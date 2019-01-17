package app

import (
	"sync"
	"time"
)

const (
	BUFF_SIZE = 1024 * 30 // byte buffer size set to 30kb
	APP_VERSION = "1.1.0-beta"
)

var (
	RUN_WITH                   int //启动模式，1：storage，2：tracker，3：client, 4:dashboard
	LOG_LEVEL                  int
	ASSIGN_DISK_SPACE          int64
	SLICE_SIZE                 int64
	LOG_INTERVAL               string //log文件精度：h/d/w/m/y
	BASE_PATH                  string
	GROUP                      string
	INSTANCE_ID                string
	SECRET                     string
	ADVERTISE_ADDRESS          string
	TRACKERS                   string
	HTTP_ENABLE                bool
	MIME_TYPES_ENABLE          bool
	UPLOAD_ENABLE              bool
	LOG_ENABLE                 bool
	PORT                       int
	ADVERTISE_PORT             int
	HTTP_PORT                  int
	CLIENT_TYPE                int //client类型，1:storage client, 2:other client, 3:dashboard client
	STORAGE_CLIENT_EXPIRE_TIME = time.Second * 60
	SYNC_MEMBER_INTERVAL       = time.Second * 5 // 60
	PULL_NEW_FILE_INTERVAL     = time.Second * 10
	SYNC_STATISTIC_INTERVAL    = time.Second * 6 // 65
	PATH_REGEX                 = "^/([0-9a-zA-Z_]{1,10})/([0-9a-zA-Z_]{1,30})/([MS])/([0-9a-fA-F]{32})$"
	MD5_REGEX                  = "^[0-9a-fA-F]{32}$"
	UUID                       = ""

	HTTP_AUTH = ""

	// statistic info
	IOIN            int64
	IOOUT           int64
	STAGE_IOIN      int64
	STAGE_IOOUT     int64
	DOWNLOADS       int64
	UPLOADS         int64
	STAGE_DOWNLOADS int
	STAGE_UPLOADS   int
	START_TIME      int64
	FILE_TOTAL      int
	FILE_FINISH     int
	DISK_USAGE      int64
	MEMORY          uint64

	LOG_LEVEL_SETS  = map[string]byte{"trace": 1, "debug": 1, "info": 1, "warm": 1, "error": 1, "fatal": 1}
	LOG_ROTATION_SETS  = map[string]byte{"h": 1, "d": 1, "m": 1, "y": 1}
)

const (
	TASK_SYNC_MEMBER       = 1 // storage同步自己的组内成员
	TASK_REGISTER_FILE       = 2
	TASK_PULL_NEW_FILE     = 3
	TASK_DOWNLOAD_FILE     = 4
	TASK_SYNC_ALL_STORAGES = 5 // client 同步所有的storage
	TASK_SYNC_STATISTIC    = 6 // dashboard同步所有的tracker统计信息
	DB_POOL_SIZE           = 20

	STATUS_ENABLED  = 1
	STATUS_DISABLED = 0
	STATUS_DELETED  = 3

	TCP_DIALOG_TIMEOUT = time.Second * 15

	ACCESS_FLAG_NONE = 0
	ACCESS_FLAG_LOOKBACK = 1
	ACCESS_FLAG_ADVERTISE = 2
)

var ioinLock sync.Mutex
var iooutLock sync.Mutex
var updownLock sync.Mutex

func UpdateIOIN(len int64) {
	ioinLock.Lock()
	defer ioinLock.Unlock()
	IOIN += len
	STAGE_IOIN += len
}
func UpdateIOOUT(len int64) {
	iooutLock.Lock()
	defer iooutLock.Unlock()
	IOOUT += len
	STAGE_IOOUT += len
}

func UpdateUploads() {
	updownLock.Lock()
	defer updownLock.Unlock()
	UPLOADS++
	STAGE_UPLOADS++
}

func UpdateDownloads() {
	updownLock.Lock()
	defer updownLock.Unlock()
	DOWNLOADS++
	STAGE_DOWNLOADS++
}

func UpdateFileTotalCount(value int) {
	updownLock.Lock()
	defer updownLock.Unlock()
	FILE_TOTAL += value
}

func UpdateFileFinishCount(value int) {
	updownLock.Lock()
	defer updownLock.Unlock()
	FILE_FINISH += value
}

func UpdateDiskUsage(value int64) {
	updownLock.Lock()
	defer updownLock.Unlock()
	DISK_USAGE += value
}
