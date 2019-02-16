package bridgev2

import "app"

// ConnectMeta operation meta for connect/validate
type ConnectMeta struct {
	Secret string `json:"secret"`
	UUID   string `json:"uuid"` // this is client uuid
}

// ConnectResponseMeta operation meta for connect/validate
type ConnectResponseMeta struct {
	UUID        string `json:"uuid"` // this is server uuid
	New4Tracker bool   `json:"new2Tracker"`
}

// SyncStorageMembersMeta register storage client operation request
type SyncStorageMembersMeta struct {
	// None: replaced by app.StorageDO
}

type SyncStorageMembersResponseMeta struct {
	LookBackAddr string          `json:"backAddr"` // tracker lookback addr
	GroupMembers []app.StorageDO `json:"members"`  // group members(not include self)
}

// RegisterFileMeta register file operation request.
type RegisterFileMeta struct {
	Files []app.FileVO `json:"files"` // 文件md5
}
type RegisterFileResponseMeta struct {
	LastInsertId int64 `json:"last_id"` // 文件md5
}

// PullFileMeta register file operation request.
type PullFileMeta struct {
	BaseId int64  `json:"baseId"` // 上次同步的ID位置（tracker端的ID）
	Group  string `json:"group"`
}

type PullFileResponseMeta struct {
	Files []app.FileVO
}

// UploadFileMeta upload file operation request.
type UploadFileMeta struct {
	FileSize int64  `json:"fileSize"` // file length
	FileExt  string `json:"ext"`      // file extension name, exclude '.'
	Md5      string `json:"md5"`      // file md5, if file exists, skip upload
}

type UploadFileResponseMeta struct {
	Path string `json:"path"`
}

// SyncAllStorageServerMeta register storage client operation request.(only for client)
type SyncAllStorageServerMeta struct {
}

type SyncAllStorageServerResponseMeta struct {
	Servers []app.StorageDO `json:"servers"`
}

// PullNewFileMeta register file operation request.
type PullNewFileMeta struct {
	BaseId int64  `json:"baseId"`
	Group  string `json:"group"`
}

type PullNewFileResponseMeta struct {
	Files []app.FileVO
}

// SyncStatisticMeta sync statistic operation request
type SyncStatisticMeta struct {
}

type SyncStatisticResponseMeta struct {
	Statistic []app.StorageDO `json:"statistic"`
	FileCount int             `json:"files"`
}

// QueryFileMeta query file operation request.
type QueryFileMeta struct {
	PathOrMd5 string `json:"pathMd5"` // file md5 or filePath like '/xxx/xxx/xxxxxx'
}

// QueryFileResponseMeta query file response.
type QueryFileResponseMeta struct {
	Exist bool       `json:"exist"` // true:the file exists
	File  app.FileVO `json:"file"`
}

// DownloadFileMeta download file operation request.
type DownloadFileMeta struct {
	Path   string `json:"path"`   // path like /G01/002/M/c445b10edc599617106ae8472c1446fd
	Start  int64  `json:"start"`  // length of bytes to skip
	Offset int64  `json:"offset"` // length of bytes to read, if Offset < 0 represents read all bytes left
}

// DownloadFileResponseMeta download file response.
type DownloadFileResponseMeta struct {
	Exist bool       `json:"exist"` // true:the file exists
	File  app.FileVO `json:"file"`
}

type ReadPos struct {
	PartIndex int
	PartStart int64
}

type Task struct {
	FileId   int64                     // file表中的id
	TaskType int                       // 任务类型，1：上报文件，2：从其他节点下载文件
	Status   int                       //任务状态
	Callback func(task *Task, e error) // callback calls when each task finish
}
