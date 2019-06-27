package api

import (
	"container/list"
	"errors"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/conn"
	"github.com/hetianyi/gox/gpip"
	"github.com/hetianyi/gox/logger"
	"io"
	"sync"
	"time"
)

const (
	DefaultMaxConnectionsPerServer = 100
)

var (
	NoStorageServerErr = errors.New("no storage available")
)

type Config struct {
	MaxConnectionsPerServer  uint                    // limit max connection for each server
	TrackerServers           []*common.Server        // tracker servers
	StaticStorageServers     []*common.StorageServer // storage servers
	RegisteredStorageServers []*common.StorageServer // storage servers
	// Trackers or Storages
}

type ClientAPI interface {
	Init(config *Config) error // Init initializes the ClientAPI.
	// RefreshConfig(config *Config) // TODO
	Upload(src io.Reader, group string) (bool, error)
	Download(input io.Reader) (bool, error)
}

func NewClient() *clientAPIImpl {
	return &clientAPIImpl{
		lock:    new(sync.Mutex),
		weights: make(map[string]int64),
	}
}

type clientAPIImpl struct {
	config  *Config
	lock    *sync.Mutex
	weights map[string]int64 // server use weights
}

func (c *clientAPIImpl) Init(config *Config) {
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
		}
	}
	if c.config.StaticStorageServers != nil {
		for _, s := range c.config.StaticStorageServers {
			conn.InitServerSettings(s, c.config.MaxConnectionsPerServer, time.Minute*5)
		}
	}
}

func (c *clientAPIImpl) Upload(src io.Reader, length int64, group string) (*common.UploadResult, error) {
	logger.Debug("begin to upload file")
	var exclude = list.New()                  // excluded storage list
	var selectedStorage *common.StorageServer // target server for file uploading.
	var lastErr error

	for {
		selectedStorage = c.selectStorageServer(group, exclude)
		if selectedStorage == nil {
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, NoStorageServerErr
		}
		connection, authenticated, err := conn.GetConnection(selectedStorage)
		if err != nil {
			lastErr = err
			exclude.PushBack(selectedStorage)
			continue
		}
		pip := &gpip.Pip{
			Conn: *connection,
		}
		if authenticated == nil || !authenticated.(bool) {
			if err = authenticate(pip, selectedStorage); err != nil {
				lastErr = err
				exclude.PushBack(selectedStorage)
				continue
			}
			logger.Debug("authentication success with server ", selectedStorage.ConnectionString())
		}
		defer func() {
			logger.Debug("upload success")
			conn.ReturnConnection(selectedStorage, connection, true, lastErr != nil)
		}()
		// send file body
		err = pip.Send(&common.Header{Operation: common.OPERATION_UPLOAD}, src, length)
		if err != nil {
			return nil, err
		}
		var ret *common.UploadResult
		// receive response
		err = pip.Receive(&common.Header{}, func(_header interface{}, bodyReader io.Reader, bodyLength int64) error {
			header := _header.(*common.Header)
			if header != nil {
				if header.Result == common.SUCCESS {
					ret = &common.UploadResult{
						Group:  header.Attributes["group"].(string),
						FileId: header.Attributes["fid"].(string),
						Node:   header.Attributes["instanceId"].(string),
					}
					return nil
				}
				return errors.New("upload failed: " + header.Msg)
			}
			return errors.New("upload failed: got empty response from server")
		})
		if err != nil {
			lastErr = err
		}
		return ret, err
	}
	// should never reached
	return nil, nil
}

// authenticate authenticates width storage server.
func authenticate(p *gpip.Pip, server conn.Server) error {
	logger.Debug("trying authentication with server ", server.ConnectionString())
	secret := ""
	if _, t := server.(*common.Server); t {
		secret = server.(*common.Server).Secret
	} else if _, t := server.(*common.StorageServer); t {
		secret = server.(*common.StorageServer).Secret
	}
	err := p.Send(&common.Header{
		Operation:  common.OPERATION_CONNECT,
		Attributes: map[string]interface{}{"secret": secret},
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
	if c.config.RegisteredStorageServers != nil {
		for _, s := range c.config.RegisteredStorageServers {
			if isExcluded(s, exclude) {
				continue
			}
			if group == "" || group == s.Group {
				candidates.PushBack(s)
			}
		}
	}
	if candidates.Len() == 0 {
		for _, s := range c.config.StaticStorageServers {
			if isExcluded(s, exclude) {
				continue
			}
			candidates.PushBack(s)
		}
	}
	logger.Debug("candidates: ", candidates)
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
	logger.Debug("selected storage server: ", selectedStorage.ConnectionString())
	return selectedStorage
}

func isExcluded(s *common.StorageServer, exclude *list.List) bool {
	con := false
	if exclude != nil {
		gox.WalkList(exclude, func(item interface{}) bool {
			if item.(*common.StorageServer).InstanceId == s.InstanceId {
				con = true
				return true
			}
			return false
		})
	}
	return con
}
