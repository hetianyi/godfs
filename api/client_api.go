package api

import (
	"container/list"
	"errors"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/conn"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/gpip"
	"github.com/hetianyi/gox/logger"
	json "github.com/json-iterator/go"
	"io"
	"net"
	"sync"
	"time"
)

const (
	// Default max connection count of each server.
	DefaultMaxConnectionsPerServer = 100
)

var NoStorageServerErr = errors.New("no storage available")

// Config is the APIClient config
type Config struct {
	MaxConnectionsPerServer uint                    // limit max connection for each server
	TrackerServers          []*common.Server        // tracker servers
	SynchronizeOnce         bool                    // synchronize with each tracker server only once
	SynchronizeOnceCallback chan int                // attached with `SynchronizeOnce`, for noticing client cli that whether all server is synced.
	StaticStorageServers    []*common.StorageServer // storage servers
}

// ClientAPI is godfs APIClient interface.
type ClientAPI interface {

	// SetConfig sets or refresh client server config.
	SetConfig(config *Config) // TODO

	// Upload uploads file to specific group server.
	//
	// If no group provided, it will upload file to a random server.
	Upload(src io.Reader, length int64, group string, isPrivate bool) (*common.UploadResult, error)

	// Download downloads a file from server.
	//
	// Return error can be common.NoStorageServerErr if there is no server available
	//
	// or common.NotFoundErr if the file cannot be found on the servers.
	Download(fileId string, offset int64, length int64, handler func(body io.Reader, bodyLength int64) error) error

	// Query queries file's information by fileId.
	//
	// Parameter `fileId` must be the pattern of common.FILE_ID_PATTERN
	Query(fileId string) (*common.FileInfo, error)

	// SyncInstances synchronizes instances from specific tracker server.
	SyncInstances(server *common.Server) (map[string]*common.Instance, error)

	// PushBinlog pushes binlog to tracker server.
	PushBinlog(server *common.Server, binlogs []common.BingLogDTO) error
}

// NewClient creates a new APIClient.
func NewClient() *clientAPIImpl {
	return &clientAPIImpl{
		lock:    new(sync.Mutex),
		weights: make(map[string]int64),
	}
}

// clientAPIImpl is the implementation of APIClient.
type clientAPIImpl struct {
	config  *Config
	lock    *sync.Mutex
	weights map[string]int64 // server use weights
}

func (c *clientAPIImpl) SetConfig(config *Config) {
	if config != nil {
		c.config = config
	} else {
		c.config = &Config{
			MaxConnectionsPerServer: DefaultMaxConnectionsPerServer,
		}
	}
	if c.config.MaxConnectionsPerServer <= 0 {
		c.config.MaxConnectionsPerServer = DefaultMaxConnectionsPerServer
	}
	if (c.config.TrackerServers == nil || len(c.config.TrackerServers) == 0) &&
		(c.config.StaticStorageServers == nil || len(c.config.StaticStorageServers) == 0) {
		logger.Warn("client initialized but no server provided")
	}
	if c.config.TrackerServers != nil {
		for _, s := range c.config.TrackerServers {
			conn.InitServerSettings(s, c.config.MaxConnectionsPerServer, time.Minute*5)
			tracks(c, s, config.SynchronizeOnce, config.SynchronizeOnceCallback)
		}
	}
	if c.config.StaticStorageServers != nil {
		for _, s := range c.config.StaticStorageServers {
			conn.InitServerSettings(s, c.config.MaxConnectionsPerServer, time.Minute*5)
		}
	}
}

func (c *clientAPIImpl) Upload(src io.Reader, length int64, group string, isPrivate bool) (*common.UploadResult, error) {
	logger.Debug("begin to upload file")
	var exclude = list.New()                  // excluded storage list
	var selectedStorage *common.StorageServer // target server for file uploading.
	var lastErr error
	var lastConn *net.Conn
	var ret *common.UploadResult
	gox.Try(func() {
		for {
			// select storage server.
			selectedStorage = c.selectStorageServer(group, exclude)
			if selectedStorage == nil {
				if lastErr == nil {
					lastErr = NoStorageServerErr
				}
				break
			}
			// get connection of this server.
			connection, authenticated, err := conn.GetConnection(selectedStorage)
			if err != nil {
				lastErr = err
				exclude.PushBack(selectedStorage)
				continue
			}
			lastConn = connection
			// construct tcp bridge.
			pip := &gpip.Pip{
				Conn: *lastConn,
			}
			// authentication with server.
			if authenticated == nil || !authenticated.(bool) {
				if err = authenticate(pip, selectedStorage); err != nil {
					lastErr = err
					exclude.PushBack(selectedStorage)
					conn.ReturnConnection(selectedStorage, lastConn, nil, true)
					lastConn = nil
					continue
				}
				logger.Debug("authentication success with server ", selectedStorage.ConnectionString())
			}
			authenticated = true
			// send file body
			err = pip.Send(&common.Header{
				Operation: common.OPERATION_UPLOAD,
				Attributes: map[string]string{
					"isPrivate": gox.TValue(isPrivate, "1", "0").(string),
				},
			}, src, length)
			if err != nil {
				lastErr = err
				conn.ReturnConnection(selectedStorage, lastConn, nil, true)
				lastConn = nil
				break
			}
			// receive response
			err = pip.Receive(&common.Header{}, func(_header interface{}, bodyReader io.Reader, bodyLength int64) error {
				header := _header.(*common.Header)
				if header != nil {
					if header.Result == common.SUCCESS {
						ret = &common.UploadResult{
							Group:    header.Attributes["group"],
							FileId:   header.Attributes["fid"],
							Instance: header.Attributes["instance"],
						}
						return nil
					}
					return errors.New("upload failed: " + header.Msg)
				}
				return errors.New("upload failed: got empty response from server")
			})
			if err != nil {
				lastErr = err
				conn.ReturnConnection(selectedStorage, lastConn, nil, true)
				lastConn = nil
				break
			}
			// upload finish
			conn.ReturnConnection(selectedStorage, lastConn, authenticated, false)
			lastErr = nil
			lastConn = nil
			logger.Debug("upload finish")
			break
		}
	}, func(e interface{}) {
		lastErr = e.(error)
	})
	// lastConn should be returned and set to nil.
	if lastConn != nil {
		conn.ReturnConnection(selectedStorage, lastConn, nil, true)
	}
	return ret, lastErr
}

func (c *clientAPIImpl) Download(fileId string, offset int64, length int64, handler func(body io.Reader, bodyLength int64) error) error {
	logger.Debug("begin to download file")
	var exclude = list.New()                  // excluded storage list
	var selectedStorage *common.StorageServer // target server for file uploading.
	var lastErr error
	var lastConn *net.Conn

	fileInfo, err := util.ParseAlias(fileId)
	if err != nil {
		return err
	}
	gox.Try(func() {
		for {
			selectedStorage = c.selectStorageServer(fileInfo.Group, exclude)
			if selectedStorage == nil {
				if lastErr == nil {
					lastErr = NoStorageServerErr
				}
				break
			}
			connection, authenticated, err := conn.GetConnection(selectedStorage)
			if err != nil {
				lastErr = err
				exclude.PushBack(selectedStorage)
				continue
			}
			lastConn = connection
			pip := &gpip.Pip{
				Conn: *lastConn,
			}
			if authenticated == nil || !authenticated.(bool) {
				if err = authenticate(pip, selectedStorage); err != nil {
					lastErr = err
					exclude.PushBack(selectedStorage)
					conn.ReturnConnection(selectedStorage, lastConn, nil, true)
					lastConn = nil
					continue
				}
				logger.Debug("authentication success with server ", selectedStorage.ConnectionString())
			}
			authenticated = true
			// send file body
			err = pip.Send(&common.Header{
				Operation: common.OPERATION_DOWNLOAD,
				Attributes: map[string]string{
					"fileId": fileId,
					"offset": convert.Int64ToStr(offset),
					"length": convert.Int64ToStr(length),
				},
			}, nil, 0)
			if err != nil {
				lastErr = err
				conn.ReturnConnection(selectedStorage, lastConn, nil, true)
				lastConn = nil
				exclude.PushBack(selectedStorage)
				continue
			}
			// receive response
			err = pip.Receive(&common.Header{}, func(_header interface{}, bodyReader io.Reader, bodyLength int64) error {
				header := _header.(*common.Header)
				if header != nil {
					if header.Result == common.SUCCESS {
						return handler(bodyReader, bodyLength)
					} else if header.Result == common.NOT_FOUND {
						return common.NotFoundErr
					} else if header.Result == common.ERROR {
						return common.ServerErr
					}
					return errors.New("upload failed: " + header.Msg)
				}
				return errors.New("download failed: got empty response from server")
			})
			if err != nil {
				lastErr = err
				conn.ReturnConnection(selectedStorage, lastConn, authenticated, err != common.NotFoundErr && err != common.ServerErr)
				lastConn = nil
				exclude.PushBack(selectedStorage)
				continue
			}
			conn.ReturnConnection(selectedStorage, lastConn, authenticated, false)
			lastErr = nil
			lastConn = nil
			logger.Debug("download finish")
			break
		}
	}, func(e interface{}) {
		logger.Error(e)
	})
	if lastConn != nil {
		conn.ReturnConnection(selectedStorage, lastConn, nil, true)
	}
	return lastErr
}

func (c *clientAPIImpl) Query(fileId string) (*common.FileInfo, error) {
	logger.Debug("begin to query file")
	var exclude = list.New()                  // excluded storage list
	var selectedStorage *common.StorageServer // target server for file uploading.
	var lastErr error
	var lastConn *net.Conn
	var result *common.FileInfo

	fileInfo, err := util.ParseAlias(fileId)
	if err != nil {
		return nil, err
	}
	// TODO offline function
	gox.Try(func() {
		for {
			selectedStorage = c.selectStorageServer(fileInfo.Group, exclude)
			if selectedStorage == nil {
				if lastErr == nil {
					lastErr = NoStorageServerErr
				}
				break
			}
			connection, authenticated, err := conn.GetConnection(selectedStorage)
			if err != nil {
				lastErr = err
				exclude.PushBack(selectedStorage)
				continue
			}
			lastConn = connection
			pip := &gpip.Pip{
				Conn: *lastConn,
			}
			if authenticated == nil || !authenticated.(bool) {
				if err = authenticate(pip, selectedStorage); err != nil {
					lastErr = err
					exclude.PushBack(selectedStorage)
					conn.ReturnConnection(selectedStorage, lastConn, nil, true)
					lastConn = nil
					continue
				}
				logger.Debug("authentication success with server ", selectedStorage.ConnectionString())
			}
			authenticated = true
			// send file body
			err = pip.Send(&common.Header{
				Operation: common.OPERATION_QUERY,
				Attributes: map[string]string{
					"fileId": fileId,
				},
			}, nil, 0)
			if err != nil {
				lastErr = err
				conn.ReturnConnection(selectedStorage, lastConn, nil, true)
				lastConn = nil
				exclude.PushBack(selectedStorage)
				continue
			}
			// receive response
			err = pip.Receive(&common.Header{}, func(_header interface{}, bodyReader io.Reader, bodyLength int64) error {
				header := _header.(*common.Header)
				if header != nil {
					if header.Result == common.SUCCESS {
						infoS := header.Attributes["info"]
						result = &common.FileInfo{}
						return json.Unmarshal([]byte(infoS), result)
					} else if header.Result == common.NOT_FOUND {
						return common.NotFoundErr
					} else if header.Result == common.ERROR {
						return common.ServerErr
					}
					return errors.New("inspect failed: " + header.Msg)
				}
				return errors.New("inspect failed: got empty response from server")
			})
			if err != nil {
				lastErr = err
				conn.ReturnConnection(selectedStorage, lastConn, authenticated, err != common.NotFoundErr && err != common.ServerErr)
				lastConn = nil
				exclude.PushBack(selectedStorage)
				continue
			}
			conn.ReturnConnection(selectedStorage, lastConn, authenticated, false)
			lastErr = nil
			lastConn = nil
			logger.Debug("inspect finish")
			break
		}
	}, func(e interface{}) {
		logger.Error(e)
	})
	if lastConn != nil {
		conn.ReturnConnection(selectedStorage, lastConn, nil, true)
	}
	return result, lastErr
}

func (c *clientAPIImpl) SyncInstances(server *common.Server) (map[string]*common.Instance, error) {
	var result = make(map[string]*common.Instance)
	connection, authenticated, err := conn.GetConnection(server)
	if err != nil {
		return nil, err
	}
	pip := &gpip.Pip{
		Conn: *connection,
	}
	if authenticated == nil || !authenticated.(bool) {
		if err = authenticate(pip, server); err != nil {
			return nil, err
		}
		logger.Debug("authentication success with server ", server.ConnectionString())
	}
	authenticated = true
	// send file body
	err = pip.Send(&common.Header{
		Operation: common.OPERATION_SYNC_INSTANCES,
	}, nil, 0)
	if err != nil {
		return nil, err
	}
	// receive response
	err = pip.Receive(&common.Header{}, func(_header interface{}, bodyReader io.Reader, bodyLength int64) error {
		header := _header.(*common.Header)
		if header != nil {
			if header.Result == common.SUCCESS {
				infoS := header.Attributes["instances"]
				instanceId := header.Attributes["instanceId"]
				var ret = make(map[string]*json.RawMessage)
				if err = json.Unmarshal([]byte(infoS), &ret); err != nil {
					return err
				}
				for k := range ret {
					var a common.Instance
					err = json.Unmarshal(*ret[k], &a)
					result[k] = &a
				}
				server.InstanceId = instanceId
				return nil
			}
			return errors.New("synchronize failed: " + header.Msg)
		}
		return errors.New("synchronize failed: got empty response from server")
	})
	if err != nil {
		return nil, err
	}
	conn.ReturnConnection(server, connection, authenticated, false)
	logger.Debug("synchronize finish, instances: ", len(result))
	return result, nil
}

func (c *clientAPIImpl) PushBinlog(server *common.Server, binlogs []common.BingLogDTO) error {
	connection, authenticated, err := conn.GetConnection(server)
	if err != nil {
		return err
	}
	pip := &gpip.Pip{
		Conn: *connection,
	}
	if authenticated == nil || !authenticated.(bool) {
		if err = authenticate(pip, server); err != nil {
			return err
		}
		logger.Debug("authentication success with server ", server.ConnectionString())
	}
	authenticated = true
	jsonAttr, _ := json.Marshal(binlogs)
	// send file body
	err = pip.Send(&common.Header{
		Operation: common.OPERATION_PUSH_BINLOGS,
		Attributes: map[string]string{
			"binlogs": string(jsonAttr),
		},
	}, nil, 0)
	if err != nil {
		return err
	}
	// receive response
	err = pip.Receive(&common.Header{}, func(_header interface{}, bodyReader io.Reader, bodyLength int64) error {
		header := _header.(*common.Header)
		if header != nil {
			if header.Result == common.SUCCESS {
				return nil
			}
			return errors.New("push failed: " + header.Msg)
		}
		return errors.New("push failed: got empty response from server")
	})
	if err != nil {
		return err
	}
	conn.ReturnConnection(server, connection, authenticated, false)
	return nil
}

// authenticate authenticates with server.
func authenticate(p *gpip.Pip, server conn.Server) error {
	logger.Debug("trying to authenticate with server ", server.ConnectionString())
	secret := ""
	if _, t := server.(*common.Server); t {
		secret = server.(*common.Server).Secret
	} else if _, t := server.(*common.StorageServer); t {
		secret = server.(*common.StorageServer).Secret
	}

	// validate with instance info
	var instance *common.Instance
	if common.BootAs == common.BOOT_TRACKER {
		conf := common.InitializedTrackerConfiguration
		advPort, _ := convert.StrToUint16(convert.IntToStr(conf.AdvertisePort))
		instance = &common.Instance{
			Server: common.Server{
				Host:       conf.AdvertiseAddress,
				Port:       advPort,
				Secret:     conf.Secret,
				InstanceId: conf.InstanceId,
			},
			Role: common.ROLE_TRACKER,
		}
	} else if common.BootAs == common.BOOT_STORAGE {
		conf := common.InitializedStorageConfiguration
		advPort, _ := convert.StrToUint16(convert.IntToStr(conf.AdvertisePort))
		instance = &common.Instance{
			Server: common.Server{
				Host:       conf.AdvertiseAddress,
				Port:       advPort,
				Secret:     conf.Secret,
				InstanceId: conf.InstanceId,
			},
			Role: common.ROLE_STORAGE,
			Attributes: map[string]string{
				"group": conf.Group,
			},
		}
	} /* else if common.BootAs == common.BOOT_PROXY {} */
	info, err := json.Marshal(instance)
	if err != nil {
		return err
	}

	err = p.Send(&common.Header{
		Operation: common.OPERATION_CONNECT,
		Attributes: map[string]string{
			"secret":   secret,
			"instance": string(info),
		},
	}, nil, 0)
	if err != nil {
		return err
	}
	return p.Receive(&common.Header{}, func(_header interface{}, bodyReader io.Reader, bodyLength int64) error {
		header := _header.(*common.Header)
		if header.Result != common.SUCCESS {
			return errors.New("authentication failed with server: " + server.ConnectionString())
		}
		return nil
	})
}

// selectStorageServer selects proper storage server.
func (c *clientAPIImpl) selectStorageServer(group string, exclude *list.List) *common.StorageServer {
	c.lock.Lock()
	defer c.lock.Unlock()
	logger.Debug("select storage server...")
	var candidates = list.New()
	// if registered storage server is not empty, use it first.
	syncStorages := FilterInstances(common.ROLE_STORAGE)
	if syncStorages.Len() > 0 {
		for ele := syncStorages.Front(); ele != nil; ele = ele.Next() {
			s := ele.Value.(*common.Instance)
			if isExcluded(s.Server, exclude) {
				continue
			}
			sg := ""
			if s.Attributes != nil {
				sg = s.Attributes["group"]
			}
			if group == "" || group == sg {
				candidates.PushBack(&common.StorageServer{
					Server: s.Server,
					Group:  sg,
				})
			}
		}
	}
	// if no candidate server, choose from static storage servers.
	// static server usually has no group configured, so here ignores the group.
	if candidates.Len() == 0 {
		for _, s := range c.config.StaticStorageServers {
			if isExcluded(s.Server, exclude) {
				continue
			}
			candidates.PushBack(s)
		}
	}
	// select smallest weights of storage server.
	var selectedStorage *common.StorageServer
	gox.WalkList(candidates, func(item interface{}) bool {
		if selectedStorage == nil {
			selectedStorage = item.(*common.StorageServer)
			return false
		}
		if c.weights[item.(*common.StorageServer).InstanceId] < c.weights[selectedStorage.InstanceId] {
			selectedStorage = item.(*common.StorageServer)
			return false
		}
		return false
	})
	if selectedStorage != nil {
		logger.Debug("selected storage server: ", selectedStorage.ConnectionString())
	}
	return selectedStorage
}

// isExcluded judges whether a storage server is in the exclude list.
func isExcluded(s common.Server, exclude *list.List) bool {
	if exclude == nil {
		return false
	}
	con := false
	gox.WalkList(exclude, func(item interface{}) bool {
		if item.(*common.StorageServer).InstanceId == s.InstanceId {
			con = true
			return true
		}
		return false
	})
	return con
}
