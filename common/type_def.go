package common

import (
	"github.com/boltdb/bolt"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/logger"
	json "github.com/json-iterator/go"
	"strings"
	"sync"
	"time"
)

type Command uint32
type Operation byte
type OperationResult byte
type BootMode uint32
type Role byte
type RegisterState byte

type StorageConfig struct {
	Trackers              []string `json:"trackers"`
	Secret                string   `json:"secret"`
	Group                 string   `json:"group"`
	BindAddress           string   `json:"bindAddress"`
	Port                  int      `json:"port"`
	AdvertiseAddress      string   `json:"advertiseAddress"`
	AdvertisePort         int      `json:"advertisePort"`
	DataDir               string   `json:"dataDir"`
	PreferredNetworks     string   `json:"preferredNetworks"`
	LogLevel              string   `json:"logLevel"`
	LogDir                string   `json:"logDir"`
	SaveLog2File          bool     `json:"saveLog2File"`
	MaxRollingLogfileSize int      `json:"maxRollingLogfileSize"`
	LogRotationInterval   string   `json:"logRotationInterval"`
	EnableHttp            bool     `json:"enableHttp"`
	HttpPort              int      `json:"httpPort"`
	EnableMimeTypes       bool     `json:"enableMimeTypes"`
	Readonly              bool     `json:"readonly"`
	PublicAccessMode      bool     `json:"publicAccessMode"`
	AllowedDomains        []string `json:"allowedDomains"`
	InstanceId            string
	HistorySecrets        map[string]string
	TmpDir                string
	ParsedTrackers        []Server
}

type TrackerConfig struct {
	Trackers              []string `json:"trackers"`
	Secret                string   `json:"secret"`
	InstanceId            string
	BindAddress           string `json:"bindAddress"`
	Port                  int    `json:"port"`
	AdvertiseAddress      string `json:"advertiseAddress"`
	AdvertisePort         int    `json:"advertisePort"`
	DataDir               string `json:"dataDir"`
	PreferredNetworks     string `json:"preferredNetworks"`
	LogLevel              string `json:"logLevel"`
	LogDir                string `json:"logDir"`
	SaveLog2File          bool   `json:"saveLog2File"`
	MaxRollingLogfileSize int    `json:"maxRollingLogfileSize"`
	LogRotationInterval   string `json:"logRotationInterval"`
	EnableHttp            bool   `json:"enableHttp"`
	HttpPort              int    `json:"httpPort"`
	HistorySecrets        map[string]string
	ParsedTrackers        []Server
}

type ClientConfig struct {
	Trackers       []string `json:"trackers"`
	Storages       []string `json:"storages"`
	LogLevel       string   `json:"logLevel"`
	Secret         string   `json:"secret"`
	PrivateUpload  bool     `json:"private_upload"`
	TestScale      int      `json:"test_scale"`
	TestThread     int      `json:"test_thread"`
	ParsedTrackers []Server
}

type Server struct {
	Host           string            `json:"host"`
	Port           uint16            `json:"port"`
	Secret         string            `json:"secret"`
	HistorySecrets map[string]string `json:"history_secret"` // 历史密码
	InstanceId     string            `json:"instanceId"`
}

type StorageServer struct {
	Server
	Group string `json:"group"`
}

func (s *Server) ConnectionString() string {
	return strings.Join([]string{s.Host, convert.Uint16ToStr(s.Port)}, ":")
}

// GetHost returns server's host.
func (s *Server) GetHost() string {
	return s.Host
}

// GetPort returns server's port.
func (s *Server) GetPort() uint16 {
	return s.Port
}

func (s *StorageServer) ToServer() *Server {
	return &Server{
		Host:       s.Host,
		Port:       s.Port,
		InstanceId: s.InstanceId,
	}
}

type Header struct {
	Operation  Operation         `json:"op"`
	Result     OperationResult   `json:"ret"`
	Msg        string            `json:"msg"`
	Attributes map[string]string `json:"ats"`
}

type UploadResult struct {
	Group    string `json:"group"`
	Instance string `json:"instance"`
	FileId   string `json:"fileId"`
}

type FileInfo struct {
	Group      string `json:"group"`
	Path       string `json:"path"`
	FileLength int64  `json:"size"`
	InstanceId string `json:"instance"`
	IsPrivate  bool   `json:"isPrivate"`
	CreateTime int64  `json:"createTime"`
}

type Instance struct {
	Server
	Role         Role              `json:"role"`
	Attributes   map[string]string `json:"ats"`
	RegisterTime int64             `json:"ts"`
	State        RegisterState     `json:"state"`
}

type InstanceMap struct {
	Instances map[string]Instance `json:"instances"`
}

type BingLog struct {
	SourceInstance [8]byte // file source instance
	FileLength     [8]byte // file length
	FileId         []byte  // fileId
}

type BingLogDTO struct {
	SourceInstance string
	FileLength     int64
	FileId         string
}

// FileId is a file
type FileId struct {
	FileId     string
	InstanceId string
	Timestamp  time.Time
}

// BinlogQueryDTO is the binlog query entity between storage servers.
type BinlogQueryDTO struct {
	FileIndex int   `json:"fileIndex"`
	Offset    int64 `json:"offset"`
}

// BinlogQueryResultDTO is the binlog query result entity between storage servers.
type BinlogQueryResultDTO struct {
	BinlogQueryDTO
	Logs []BingLogDTO `json:"logs"`
}

type ConfigMap struct {
	db *bolt.DB
}

var configMapLock = new(sync.Mutex)

func NewConfigMap(path string) (*ConfigMap, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists([]byte(BUCKET_KEY_CONFIGMAP))
		if e != nil {
			return e
		}
		_, e = tx.CreateBucketIfNotExists([]byte(BUCKET_KEY_FAILED_BINLOG_POS))
		if e != nil {
			return e
		}
		if BootAs == BOOT_TRACKER {
			_, e := tx.CreateBucketIfNotExists([]byte(BUCKET_KEY_FILEID))
			if e != nil {
				return nil
			}
		}
		return e
	})
	return &ConfigMap{db}, err
}

func (c *ConfigMap) BatchUpdate(w func(tx *bolt.Tx) error) error {
	configMapLock.Lock()
	defer func() {
		configMapLock.Unlock()
		if err := recover(); err != nil {
			logger.Error("error performing action BatchUpdate: ", err)
		}
	}()
	return c.db.Batch(func(tx *bolt.Tx) error {
		return w(tx)
	})
}

func (c *ConfigMap) PutConfig(key string, value []byte) error {
	configMapLock.Lock()
	defer func() {
		configMapLock.Unlock()
		if err := recover(); err != nil {
			logger.Error("error performing action PutConfig: ", err)
		}
	}()

	return c.db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_KEY_CONFIGMAP))
		b.Put([]byte(key), value)
		return nil
	})
}

func (c *ConfigMap) GetConfig(key string) (ret []byte, err error) {
	err = c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_KEY_CONFIGMAP))
		ret = b.Get([]byte(key))
		return nil
	})
	return
}

func (c *ConfigMap) PutFile(binlogs []BingLogDTO) error {
	configMapLock.Lock()
	defer func() {
		configMapLock.Unlock()
		if err := recover(); err != nil {
			logger.Error("error performing action PutFile: ", err)
		}
	}()

	return c.db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_KEY_FILEID))
		for _, log := range binlogs {
			if err := b.Put([]byte(log.FileId), []byte(convert.Int64ToStr(log.FileLength))); err != nil {
				return err
			}
		}
		return nil
	})
}

func (c *ConfigMap) GetFile(key string) (ret []byte, err error) {
	err = c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_KEY_FILEID))
		ret = b.Get([]byte(key))
		return nil
	})
	return
}

func (c *ConfigMap) PutFailedBinlogPos(binlogPos *BinlogQueryDTO) error {
	configMapLock.Lock()
	defer func() {
		configMapLock.Unlock()
		if err := recover(); err != nil {
			logger.Error("error performing action PutFailedBinlogPos: ", err)
		}
	}()

	bs, err := json.Marshal(binlogPos)
	if err != nil {
		return err
	}
	return c.db.Batch(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(BUCKET_KEY_FAILED_BINLOG_POS)).Put(bs, nil)
	})
}

func (c *ConfigMap) IteratorFailedBinlog(iterator func(c *bolt.Cursor) error) error {
	return c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_KEY_FAILED_BINLOG_POS))
		return iterator(b.Cursor())
	})
}
