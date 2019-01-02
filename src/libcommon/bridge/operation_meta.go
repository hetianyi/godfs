package bridge

import (
    "time"
    "app"
)

const (
	STATUS_OK                    = 0
	STATUS_BAD_SECRET            = 1
	STATUS_OPERATION_NOT_SUPPORT = 2
	STATUS_INTERNAL_SERVER_ERROR = 3
	STATUS_NOT_FOUND             = 4
	STATUS_UPLOAD_DISABLED       = 5
)

type Member struct {
    LookBackAddress string `json:"lookBackAddr"`
    Port          int    `json:"port"`
    AdvertiseAddr string `json:"addr"`
    AdvertisePort int    `json:"advertise_port"`
    InstanceId    string `json:"instance_id"`
    Group         string `json:"group"`
    HttpPort      int    `json:"httpPort"`
    HttpEnable    bool   `json:"httpEnable"`
    ReadOnly      bool   `json:"readonly"`
}

type ExpireMember struct {
    LookBackAddress string
    Port          int
    AdvertiseAddr string
    AdvertisePort int
    InstanceId    string
    Group         string
    HttpPort      int
    HttpEnable    bool
    ExpireTime    time.Time
    ReadOnly      bool
    // 1: use LookBackAddress:Port 2: use AdvertiseAddr:AdvertisePort
    AccessFlag int
}

func (expireMember *ExpireMember) From(member *Member) {
	expireMember.AdvertiseAddr = member.AdvertiseAddr
	expireMember.InstanceId = member.InstanceId
	expireMember.Group = member.Group
	expireMember.Port = member.Port
	expireMember.ReadOnly = member.ReadOnly
	expireMember.HttpPort = member.HttpPort
	expireMember.HttpEnable = member.HttpEnable
	expireMember.LookBackAddress = member.LookBackAddress
	expireMember.AdvertisePort = member.AdvertisePort
	expireMember.AccessFlag = app.ACCESS_FLAG_NONE
}


func (expireMember *ExpireMember) SwitchAccessFlag() {
	if expireMember.AccessFlag == app.ACCESS_FLAG_LOOKBACK {
		expireMember.AccessFlag = app.ACCESS_FLAG_ADVERTISE
	} else {
		expireMember.AccessFlag = app.ACCESS_FLAG_LOOKBACK
	}
}



func (expireMember *ExpireMember) GetHostAndPortByAccessFlag() (host string, port int) {
    if expireMember.AccessFlag == app.ACCESS_FLAG_NONE {
    	// if run as client, always try from advertise ip
    	if app.RUN_WITH == 3 {
			expireMember.AccessFlag = app.ACCESS_FLAG_ADVERTISE
			return expireMember.AdvertiseAddr, expireMember.AdvertisePort
		}
		expireMember.AccessFlag = app.ACCESS_FLAG_LOOKBACK
        return expireMember.LookBackAddress, expireMember.Port
    }
    if expireMember.AccessFlag == app.ACCESS_FLAG_LOOKBACK {
		return expireMember.LookBackAddress, expireMember.Port
	}
	return expireMember.AdvertiseAddr, expireMember.AdvertisePort
}


type FilePart struct {
	Fid      int    `json:"fid"`  // 分片所属文件的id
	Id       int    `json:"id"`   // 分片id
	Md5      string `json:"md5"`  // 分片md5
	FileSize int64  `json:"size"` // 文件大小
}

type File struct {
	Id       int        `json:"id"`       // 文件id
	Md5      string     `json:"md5"`      // 文件md5
	PartNum  int        `json:"partNum"`  // 文件分片数量
	Group    string     `json:"group"`    // 组id
	Instance string     `json:"instance"` // 实例id
	Parts    []FilePart `json:"parts"`    // 文件分片数组
}

// 映射任务表task中字段
type Task struct {
	FileId   int                       // file表中的id
	TaskType int                       // 任务类型，1：上报文件，2：从其他节点下载文件
	Status   int                       //任务状态
	Callback func(task *Task, e error) // callback calls when each task finish
}

type ReadPos struct {
	PartIndex int
	PartStart int64
}

type ServerStatistic struct {
	UUID          string `json:"uuid"`
	AdvertiseAddr string `json:"addr"`
	Group         string `json:"group"`
	InstanceId    string `json:"instance_id"`
	Port          int    `json:"port"`
	HttpPort      int    `json:"httpPort"`
	HttpEnable    bool   `json:"httpEnable"`
	// 统计信息
	TotalFiles int   `json:"files"`
	Finish     int   `json:"finish"`
	StartTime  int64 `json:"startTime"`
	Downloads  int   `json:"downloads"`
	Uploads    int   `json:"uploads"`

	StageDownloads int `json:"stageDownloads"`
	StageUploads   int `json:"stageUploads"`

	IOin       int64  `json:"in"`
	IOout      int64  `json:"out"`
	StageIOin  int64  `json:"stageIOin"`
	StageIOout int64  `json:"stageIOout"`
	DiskUsage  int64  `json:"disk"`
	Memory     uint64 `json:"mem"`
	ReadOnly   bool   `json:"readonly"`
	LogTime    int64  `json:"logTime"`
}

// validate operation request.
type OperationValidationRequest struct {
	Secret string `json:"secret"`
	UUID   string `json:"uuid"`
}

// validate operation response.
type OperationValidationResponse struct {
	Status int    `json:"status"`
	UUID   string `json:"uuid"`
	IsNew  bool   `json:"isnew"` // tracker是否标志新client
}

// register storage client operation request.
type OperationRegisterStorageClientRequest struct {
	UUID          string `json:"uuid"`
	AdvertiseAddr string `json:"addr"`
	Group         string `json:"group"`
	InstanceId    string `json:"instance_id"`
	AdvertisePort int    `json:"advertise_port"`
	Port          int    `json:"port"`
	HttpPort      int    `json:"httpPort"`
	HttpEnable    bool   `json:"httpEnable"`
	// 统计信息
	TotalFiles int   `json:"files"`
	Finish     int   `json:"finish"`
	StartTime  int64 `json:"startTime"`
	LogTime    int64 `json:"logTime"`
	Downloads  int   `json:"downloads"`
	Uploads    int   `json:"uploads"`
	IOin       int64 `json:"in"`
	IOout      int64 `json:"out"`

	StageDownloads int   `json:"stageDownloads"`
	StageUploads   int   `json:"stageUploads"`
	StageIOin      int64 `json:"stageIOin"`
	StageIOout     int64 `json:"stageIOout"`

	DiskUsage int64  `json:"disk"`
	Memory    uint64 `json:"mem"`
	ReadOnly  bool   `json:"readonly"`

	// add at 2019/01/02
    // consider using in docker stack environment
    // if 'advertise_addr' is not satisfied for docker stack environment when a group has multiple instances,
    // then the client(include storage client, java client and native client) can only get a single address of a group,
    // this address is usually the 'advertise_addr' parameter specified in docker compose file.
    // client always use LookBackAddress to synchronize files between each other first, the 'advertise_addr' is secondary choice.
	LookBackAddress string `json:"lookBackAddr"`
}

// validate operation response.
type OperationRegisterStorageClientResponse struct {
	Status       int      `json:"status"`
	LookBackAddr string   `json:"backAddr"` // tracker反视地址
	GroupMembers []Member `json:"members"`  // 我的组内成员（不包括自己）
}

// register storage client operation request.(only for client)
type OperationGetStorageServerRequest struct {
}

// validate operation response.
type OperationGetStorageServerResponse struct {
	Status       int      `json:"status"`
	GroupMembers []Member `json:"members"` // 我的组内成员（不包括自己）
}

// upload file operation request.
type OperationUploadFileRequest struct {
	FileSize uint64 `json:"fileSize"` // 文件大小
	FileExt  string `json:"ext"`      //文件扩展名，不包含'.'
	Md5      string `json:"md5"`      //文件md5, 如果已存在则不需要上传
}

// upload file response.
type OperationUploadFileResponse struct {
	Status int    `json:"status"`
	Path   string `json:"path"`
}

// query file operation request.
type OperationQueryFileRequest struct {
	PathOrMd5 string `json:"md5"` //文件md5, 如果已存在则不需要上传
}

// query file response.
type OperationQueryFileResponse struct {
	Status int   `json:"status"`
	Exist  bool  `json:"exist"` // true:the file exists
	File   *File `json:"file"`
}

// download file operation request.
type OperationDownloadFileRequest struct {
	Path   string `json:"path"`   // path like /G01/002/M/c445b10edc599617106ae8472c1446fd
	Start  int64  `json:"start"`  // length of bytes to skip
	Offset int64  `json:"offset"` // length of bytes to read, if Offset < 0 represents read all bytes left
}

// download file response.
type OperationDownloadFileResponse struct {
	Status int `json:"status"`
}

// register file operation request.
type OperationRegisterFileRequest struct {
	Files []File `json:"files"` // 文件md5
}

// register file response.
type OperationRegisterFileResponse struct {
	Status int `json:"status"`
}

// register file operation request.
type OperationPullFileRequest struct {
	BaseId int    `json:"baseId"` // 上次同步的ID位置（tracker端的ID）
	Group  string `json:"group"`
}

// register file response.
type OperationPullFileResponse struct {
	Status int `json:"status"`
	Files  []File
}

type OperationSyncStatisticRequest struct {
}

type OperationSyncStatisticResponse struct {
	Status    int               `json:"status"`
	Statistic []ServerStatistic `json:"statistic"`
	FileCount int               `json:"files"`
}

type TrackerConfig struct {
	UUID          string
	TrackerSyncId int
	LocalPushId   int
}

type WebTracker struct {
	Id      int       `json:"id"`
	UUID    string    `json:"uuid"`
	Host    string    `json:"host"`
	Port    int       `json:"port"`
	Status  int       `json:"status"` // 0: disabled,  1:enabled, 3: deleted
	Secret  string    `json:"secret"`
	Remark  string    `json:"remark"`
	AddTime time.Time `json:"addTime"`
}

type WebStorage struct {
	Id         int    `json:"id"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Status     int    `json:"status"` // 0: disabled,  1:enabled, 3: deleted
	Tracker    int    `json:"tracker"`
	TotalFiles int    `json:"totalFiles"`
	UUID       string `json:"uuid"`

	IOin  int64 `json:"in"`
	IOout int64 `json:"out"`

	Group      string `json:"group"`
	InstanceId string `json:"instance_id"`
	HttpPort   int    `json:"httpPort"`
	HttpEnable bool   `json:"httpEnable"`
	StartTime  int64  `json:"startTime"`
	Downloads  int    `json:"downloads"`
	Uploads    int    `json:"uploads"`
	DiskUsage  int64  `json:"disk"`
	Memory     uint64 `json:"mem"`
	ReadOnly   bool   `json:"readonly"`

	StageDownloads int   `json:"stageDownloads"`
	StageUploads   int   `json:"stageUploads"`
	StageIOin      int64 `json:"stageIOin"`
	StageIOout     int64 `json:"stageIOout"`
	LogTime        int64 `json:"logTime"`
    AdvertisePort int    `json:"advertise_port"`

    LookBackAddress string `json:"lookBackAddr"`
}

type IndexStatistic struct {
	IOin      int64  `json:"ioin"`
	IOout     int64  `json:"ioout"`
	Downloads int    `json:"downloads"`
	Uploads   int    `json:"uploads"`
	Tracker   int    `json:"tracker"`
	Storage   int    `json:"storage"`
	Files     int    `json:"files"`
	UpTime    string `json:"up_time"`
}
