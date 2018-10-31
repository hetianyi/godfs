package app

import (
	"sync"
	"time"
)

const (
	BUFF_SIZE = 1024 * 30 // byte buffer size set to 30kb
)

var (
	RUN_WITH                     int //启动模式，1：storage，2：tracker，3：client
	ASSIGN_DISK_SPACE            int64
	SLICE_SIZE                   int64
	LOG_INTERVAL                 string //log文件精度：h/d/w/m/y
	BASE_PATH                    string
	GROUP                        string
	INSTANCE_ID                  string
	SECRET                       string
	ADVERTISE_ADDRESS            string
	TRACKERS                     string
	HTTP_ENABLE                  bool
	MIME_TYPES_ENABLE            bool
	UPLOAD_ENABLE                bool
	LOG_ENABLE                   bool
	PORT                         int
	HTTP_PORT                    int
	CLIENT_TYPE                  int //client类型，1：storage client，2：other client
	STORAGE_CLIENT_EXPIRE_TIME   = time.Second * 60
	SYNC_MEMBER_INTERVAL         = time.Second * 30
	PULL_NEW_FILE_INTERVAL       = time.Second * 10 //每5s取一次同步任务
	QUWEY_DOWNLOAD_FILE_INTERVAL = time.Second * 15 //每5s取一次同步任务
	PATH_REGEX                   = "^/([0-9a-zA-Z_]{1,10})/([0-9a-zA-Z_]{1,10})/([MS])/([0-9a-fA-F]{32})$"
	MD5_REGEX                   = "^[0-9a-fA-F]{32}$"
	UUID                         = ""

	HTTP_AUTH = ""

	// statistic info
	IOIN        int64
	IOOUT       int64
	DOWNLOADS   int
	UPLOADS     int
	START_TIME  int64
	FILE_TOTAL  int
	FILE_FINISH int
	DISK_USAGE  int64
	MEMORY      uint64
)

const (
	TASK_SYNC_MEMBER       = 1 // storage同步自己的组内成员
	TASK_REPORT_FILE       = 2
	TASK_PULL_NEW_FILE     = 3
	TASK_DOWNLOAD_FILE     = 4
	TASK_SYNC_ALL_STORAGES = 5 // client 同步所有的storage
	DB_POOL_SIZE           = 20
)

var ioinLock sync.Mutex
var iooutLock sync.Mutex
var updownLock sync.Mutex

func UpdateIOIN(len int64) {
	ioinLock.Lock()
	defer ioinLock.Unlock()
	IOIN += len
}
func UpdateIOOUT(len int64) {
	iooutLock.Lock()
	defer iooutLock.Unlock()
	IOOUT += len
}

func UpdateUploads() {
	updownLock.Lock()
	defer updownLock.Unlock()
	UPLOADS++
}

func UpdateDownloads() {
	updownLock.Lock()
	defer updownLock.Unlock()
	DOWNLOADS++
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
