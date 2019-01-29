package app

import (
	"container/list"
	"sync"
	"time"
)

const (
	BufferSize   = 1024 * 30 // byte buffer size set to 30kb
	Version = "1.1.0-beta"
)

var (
	RunWith                   int // 启动模式，1：storage，2：tracker，3：client, 4:dashboard
	LogLevel                  int
	AssignDiskSpace          int64
	SliceSize                 int64
	LogInterval               string // log文件精度：h/d/w/m/y
	BasePath                  string
	Group                      string
	InstanceId                string
	Secret                     string
	AdvertiseAddress          string
	Trackers                   string
	HttpEnable                bool
	MimeTypesEnable          bool
	UploadEnable              bool
	LogEnable                 bool
	Port                       int
	AdvertisePort             int
	HttpPort                  int
	ClientType                int // client类型，1:storage client, 2:other client, 3:dashboard client
	StorageClientExpireTime = time.Second * 60
	SyncMemberInterval       = time.Second * 5 // 60
	PullNewFileInterval     = time.Second * 10
	SyncStatisticInterval    = time.Second * 6 // 65
	PathRegex                 = "^/([0-9a-zA-Z_]{1,10})/([0-9a-zA-Z_]{1,30})/([MS])/([0-9a-fA-F]{32})$"
	Md5Regex                  = "^[0-9a-fA-F]{32}$"
	UUID                       = ""

	HttpAuth = ""

	// statistic info
	IOIn            int64
	IOOut           int64
	StageIOIn      int64
	StageIOOut     int64
	Downloads       int64
	Uploads         int64
	StageDownloads int
	StageUploads   int
	StartTime      int64
	TotalFiles      int
	FinishFiles     int
	DiskUsage      int64
	Memory          uint64

	LogLevelSet    = map[string]byte{"trace": 1, "debug": 1, "info": 1, "warm": 1, "error": 1, "fatal": 1}
	LogRotationSet = map[string]byte{"h": 1, "d": 1, "m": 1, "y": 1}

	PreferredNetworks  list.List
	PreferredIPPrefix string
)

const (
	TaskSyncMembers       = 1 // storage同步自己的组内成员
	TaskRegisterFiles     = 2
	TaskPullNewFiles     = 3
	TaskDownloadFiles     = 4
	TaskSyncAllStorages = 5 // client 同步所有的storage
	TaskSyncStatistic    = 6 // dashboard同步所有的tracker统计信息
	DbPoolSize           = 20

	StatusEnabled  = 1
	STATUS_DISABLED = 0
	STATUS_DELETED  = 3

	TCP_DIALOG_TIMEOUT = time.Second * 15

	ACCESS_FLAG_NONE      = 0
	ACCESS_FLAG_LOOKBACK  = 1
	ACCESS_FLAG_ADVERTISE = 2
)

var ioinLock sync.Mutex
var iooutLock sync.Mutex
var updownLock sync.Mutex

func UpdateIOIN(len int64) {
	ioinLock.Lock()
	defer ioinLock.Unlock()
	IOIn += len
	StageIOIn += len
}
func UpdateIOOUT(len int64) {
	iooutLock.Lock()
	defer iooutLock.Unlock()
	IOOut += len
	StageIOOut += len
}

func UpdateUploads() {
	updownLock.Lock()
	defer updownLock.Unlock()
	Uploads++
	StageUploads++
}

func UpdateDownloads() {
	updownLock.Lock()
	defer updownLock.Unlock()
	Downloads++
	StageDownloads++
}

func UpdateFileTotalCount(value int) {
	updownLock.Lock()
	defer updownLock.Unlock()
	TotalFiles += value
}

func UpdateFileFinishCount(value int) {
	updownLock.Lock()
	defer updownLock.Unlock()
	FinishFiles += value
}

func UpdateDiskUsage(value int64) {
	updownLock.Lock()
	defer updownLock.Unlock()
	DiskUsage += value
}
