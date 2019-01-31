package libclient

import (
	"app"
	"container/list"
	"errors"
	json "github.com/json-iterator/go"
	"io"
	"libcommon"
	"libcommon/bridgev2"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"util/file"
	"util/logger"
	"util/pool"
)

// each client has one tcp connection with storage server,
// once the connection is broken, the client will destroy.
// one client can only do 1 operation at a time.
var addLock *sync.Mutex
var ErrNoTracker = errors.New("no tracker server available")
var ErrNoStorage = errors.New("no storage server available")

func init() {
	addLock = new(sync.Mutex)
}

// Client have different meanings under different use cases.
// client usually communicate with tracker server.
type Client struct {
	TrackerMaintainer *TrackerMaintainer // tracker maintainer for client
	connPool          *pool.ClientConnectionPool
	MaxConnPerServer  int // 客户端和每个服务建立的最大连接数，web项目中建议设置为和最大线程相同的数量
}

// NewClient create a new client.
func NewClient(MaxConnPerServer int) *Client {
	logger.Debug("init native godfs client, max conn per server:", MaxConnPerServer)
	connPool := &pool.ClientConnectionPool{}
	connPool.Init(MaxConnPerServer)
	return &Client{connPool: connPool}
}

// Upload upload file to storage server.
func (client *Client) Upload(path string, group string, startTime time.Time, skipCheck bool) (string, error) {
	fi, e := file.GetFile(path)
	if e != nil {
		return "", errors.New("error upload file " + path + " due to " + e.Error())
	}
	defer fi.Close()
	fStat, e1 := fi.Stat()
	if e1 != nil {
		return "", errors.New("error stat file " + path + " due to " + e1.Error())
	}
	// if file length < 3MB, skip check
	if fStat.Size() < 3145728 {
		skipCheck = true
	}

	fileMd5 := ""
	logger.Info("upload file:", fi.Name())

	if !skipCheck {
		logger.Debug("pre check file md5:", fi.Name())
		md5, ee := file.GetFileMd5(path)
		if ee == nil {
			fileMd5 = md5
			qfi, ee1 := client.QueryFile(md5)
			if qfi != nil {
				sm := "S"
				if qfi.PartNumber > 1 {
					sm = "M"
				}
				logger.Debug("file already exists, skip upload.")
				return qfi.Group + "/" + qfi.Instance + "/" + sm + "/" + qfi.Md5, nil
			} else {
				logger.Debug("error query file info from tracker server:", ee1)
			}
		} else {
			logger.Debug("error check file md5:", ee, ", skip pre check.")
		}
	}

	var excludes list.List
	var member *app.StorageDO
	server := &app.ServerInfo{}
	var tcpClient *bridgev2.TcpBridgeClient

	for {
		// select a storage server which match the given regulation from all members
		member = selectStorageServer(group, "", &excludes, true)
		// no available storage server
		if member == nil {
			return "", ErrNoStorage
		}
		// construct server info from storage member
		server.FromStorage(member)
		tcpClient = bridgev2.NewTcpClient(server)
		// connect to storage server
		e1 := tcpClient.Connect()
		if e1 != nil {
			h, p := server.GetHostAndPortByAccessFlag()
			logger.Error("error connect to storage server", h+":"+strconv.Itoa(p), "due to:", e1.Error())
			excludes.PushBack(member)
			continue
		}
		// validate connection
		_, e2 := tcpClient.Validate()
		if e2 != nil {
			h, p := server.GetHostAndPortByAccessFlag()
			logger.Error("error validate with storage server", h+":"+strconv.Itoa(p), "due to:", e2.Error())
			excludes.PushBack(member)
			continue
		}
		// connection and validate success, continue works below
		break
	}

	h, p := server.GetHostAndPortByAccessFlag()
	logger.Info("using storage server", h+":"+strconv.Itoa(p), "("+member.Uuid+")")

	fInfo, _ := fi.Stat()
	uploadMeta := &bridgev2.UploadFileMeta{
		FileSize: fInfo.Size(),
		FileExt:  file.GetFileExt(fInfo.Name()),
		Md5:      fileMd5,
	}
	destroy := false
	resMeta, err := tcpClient.UploadFile(uploadMeta, func(manager *bridgev2.ConnectionManager, frame *bridgev2.Frame) error {
		// begin upload file body bytes
		buff, _ := bridgev2.MakeBytes(app.BufferSize, false, 0, false)
		var finish, total int64
		var stopFlag = false
		defer func() {
			stopFlag = true
			bridgev2.RecycleBytes(buff)
		}()
		total = fInfo.Size()
		finish = 0
		go libcommon.ShowPercent(&total, &finish, &stopFlag, startTime)
		for {
			len5, e4 := fi.Read(buff)
			if e4 != nil && e4 != io.EOF {
				return e4
			}
			if len5 > 0 {
				len3, e5 := manager.Conn.Write(buff[0:len5])
				finish += int64(len5)
				if e5 != nil {
					destroy = true
					return e5
				}
				if len3 != len(buff[0:len5]) {
					destroy = true
					return errors.New("could not write enough bytes")
				}
			} else {
				if e4 != io.EOF {
					return e4
				} else {
					logger.Debug("upload finish")
				}
				break
			}
		}
		return nil
	})

	if destroy {
		tcpClient.GetConnManager().Destroy()
	} else {
		tcpClient.GetConnManager().Close()
	}
	if resMeta != nil {
		return "", err
	}
	return resMeta.Path, err
}

// QueryFile query file from tracker server.
func (client *Client) QueryFile(pathOrMd5 string) (*app.FileVO, error) {
	logger.Debug("query file info:", pathOrMd5)
	var result *app.FileVO
	ls := libcommon.ParseTrackers(app.Trackers)
	trackerMap := make(map[string]string)
	if ls != nil {
		for ele := ls.Front(); ele != nil; ele = ele.Next() {
			trackerMap[ele.Value.(string)] = app.Secret
		}
	}
	for k := range trackerMap {
		server := &app.ServerInfo{}
		server.FromConnStr(k)
		tcpClient := bridgev2.NewTcpClient(server)
		// connect to tracker server
		e1 := tcpClient.Connect()
		if e1 != nil {
			h, p := server.GetHostAndPortByAccessFlag()
			logger.Error("error connect to tracker server", h+":"+strconv.Itoa(p), "due to:", e1.Error())
			tcpClient.GetConnManager().Destroy()
			continue
		}
		// validate connection
		_, e2 := tcpClient.Validate()
		if e2 != nil {
			h, p := server.GetHostAndPortByAccessFlag()
			logger.Error("error validate with tracker server", h+":"+strconv.Itoa(p), "due to:", e2.Error())
			tcpClient.GetConnManager().Destroy()
			continue
		}
		meta := &bridgev2.QueryFileMeta{PathOrMd5: pathOrMd5}
		resMeta, e3 := tcpClient.QueryFile(meta)
		if e3 != nil {
			h, p := server.GetHostAndPortByAccessFlag()
			logger.Debug("error query file from tracker server", h+":"+strconv.Itoa(p), "due to:", e2.Error())
			tcpClient.GetConnManager().Destroy()
			continue
		}
		if resMeta == nil || !resMeta.Exist {
			h, p := server.GetHostAndPortByAccessFlag()
			logger.Debug("query file returns no result from tracker server", h+":"+strconv.Itoa(p))
			tcpClient.GetConnManager().Close()
			continue
		}
		result = &resMeta.File
		tcpClient.GetConnManager().Close()
		break
	}
	return result, nil
}

// DownloadFile download file part.
func (client *Client) DownloadFile(path string,
	start int64,
	offset int64,
	bodyWriterHandler func(manager *bridgev2.ConnectionManager, frame *bridgev2.Frame, resMeta *bridgev2.DownloadFileResponseMeta) (bool, error)) error {

	path = strings.TrimSpace(path)
	if strings.Index(path, "/") != 0 {
		path = "/" + path
	}
	if mat, _ := regexp.Match(app.PathRegex, []byte(path)); !mat {
		return errors.New("file path format error")
	}
	return client.Download(path, start, offset, true, new(list.List), bodyWriterHandler)
}

// Download download file from other storage server.
func (client *Client) Download(path string,
	start int64,
	offset int64,
	fromSrc bool,
	excludes *list.List,
	bodyWriterHandler func(manager *bridgev2.ConnectionManager, frame *bridgev2.Frame, resMeta *bridgev2.DownloadFileResponseMeta) (bool, error)) error {

	group := regexp.MustCompile(app.PathRegex).ReplaceAllString(path, "${1}")
	instanceId := regexp.MustCompile(app.PathRegex).ReplaceAllString(path, "${2}")
	if excludes == nil {
		excludes = new(list.List)
	}
	var member *app.StorageDO
	var server = &app.ServerInfo{}

	for {
		// select storage server
		if fromSrc {
			member = selectStorageServer(group, instanceId, excludes, false)
		} else {
			member = selectStorageServer(group, "", excludes, false)
		}
		if member != nil {
			excludes.PushBack(member)
		} else {
			if !fromSrc {
				return ErrNoStorage
			} else {
				logger.Debug("source server is not available(" + instanceId + ")")
				fromSrc = false
				continue
			}
		}

		server.FromStorage(member)
		h, p := server.GetHostAndPortByAccessFlag()
		if fromSrc {
			logger.Debug("try to download file from source server:", h+":"+strconv.Itoa(p))
		} else {
			logger.Debug("try to download file from storage server:", h+":"+strconv.Itoa(p))
		}

		tcpClient := bridgev2.NewTcpClient(server)
		// connect to storage server
		e1 := tcpClient.Connect()
		if e1 != nil {
			h, p := server.GetHostAndPortByAccessFlag()
			logger.Error("error connect to storage server", h+":"+strconv.Itoa(p), "due to:", e1.Error())
			tcpClient.GetConnManager().Destroy()
			continue
		}
		// validate connection
		_, e2 := tcpClient.Validate()
		if e2 != nil {
			h, p := server.GetHostAndPortByAccessFlag()
			logger.Error("error validate with storage server", h+":"+strconv.Itoa(p), "due to:", e2.Error())
			tcpClient.GetConnManager().Destroy()
			continue
		}

		meta := &bridgev2.DownloadFileMeta{
			Path:   path,
			Start:  start,
			Offset: offset,
		}
		resMeta, frame, e3 := tcpClient.DownloadFile(meta)
		if e3 != nil || resMeta == nil || !resMeta.Exist {
			logger.Error("error download from storage server", h+":"+strconv.Itoa(p), "due to: file not found")
			tcpClient.GetConnManager().Destroy()
			return client.Download(path, start, offset, false, excludes, bodyWriterHandler)
		}
		bs, _ := json.Marshal(resMeta.File)
		logger.Debug("download file info:", string(bs))

		retry, e5 := bodyWriterHandler(tcpClient.GetConnManager(), frame, resMeta)
		if e5 != nil {
			tcpClient.GetConnManager().Destroy()
			logger.Error("error download from storage server", h+":"+strconv.Itoa(p), "due to:", e5.Error())
			if retry {
				return client.Download(path, start, offset, false, excludes, bodyWriterHandler)
			} else {
				break
			}
		}
		tcpClient.GetConnManager().Close()
		break
	}
	return nil
}

// selectStorageServer select a storage server matching given group and instanceId
// excludes contains fail storage and not gonna use this time.
func selectStorageServer(group string, instanceId string, excludes *list.List, upload bool) *app.StorageDO {
	memberIteLock.Lock()
	defer memberIteLock.Unlock()
	var pick list.List
	for ele := GroupMembers.Front(); ele != nil; ele = ele.Next() {
		b := ele.Value.(app.StorageDO)
		if containsMember(&b, excludes) || (upload && b.ReadOnly) {
			continue
		}
		match1 := false
		match2 := false
		if group == "" || group == b.Group {
			match1 = true
		}
		if instanceId == "" || instanceId == b.InstanceId {
			match2 = true
		}
		if match1 && match2 {
			pick.PushBack(b)
		}
	}
	if pick.Len() == 0 {
		return nil
	}
	rd := rand.Intn(pick.Len())
	index := 0
	for ele := pick.Front(); ele != nil; ele = ele.Next() {
		if index == rd {
			s := ele.Value.(app.StorageDO)
			return &s
		}
		index++
	}
	return nil
}

// containsMember query if a list contains the given storage server.
func containsMember(mem *app.StorageDO, excludes *list.List) bool {
	if excludes == nil {
		return false
	}
	for ele := excludes.Front(); ele != nil; ele = ele.Next() {
		if ele.Value.(*app.StorageDO).Uuid == mem.Uuid {
			return true
		}
	}
	return false
}
