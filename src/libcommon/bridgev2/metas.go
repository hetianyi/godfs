package bridgev2

import "app"

// operation meta for connect/validate
type ConnectMeta struct {
	Secret string `json:"secret"`
	UUID   string `json:"uuid"` // this is client uuid
}

// operation meta for connect/validate
type ConnectResponseMeta struct {
	UUID        string `json:"uuid"` // this is server uuid
	New4Tracker bool   `json:"new"`
}

// register storage client operation request
type SyncStorageMembersMeta struct {
	// None: replaced by app.StorageDO
}

// validate operation response.
type SyncStorageMembersResponseMeta struct {
	LookBackAddr string          `json:"backAddr"` // tracker lookback addr
	GroupMembers []app.StorageDO `json:"members"`  // group members(not include self)
}

// register file operation request.
type RegisterFileMeta struct {
	Files []app.FileVO `json:"files"` // 文件md5
}
type RegisterFileResponseMeta struct {
	LastInsertId int64 `json:"last_id"` // 文件md5
}

// register file operation request.
type PullFileMeta struct {
	BaseId int64  `json:"baseId"` // 上次同步的ID位置（tracker端的ID）
	Group  string `json:"group"`
}

// register file response.
type PullFileResponseMeta struct {
	Files []app.FileVO
}

// upload file operation request.
type UploadFileMeta struct {
	FileSize uint64 `json:"fileSize"` // file length
	FileExt  string `json:"ext"`      // file extension name, exclude '.'
	Md5      string `json:"md5"`      // file md5, if file exists, skip upload
}

// upload file response.
type UploadFileResponseMeta struct {
	Path string `json:"path"`
}
