package bridge


const (
    STATUS_OK = 0
    STATUS_BAD_SECRET = 1
    STATUS_OPERATION_NOT_SUPPORT = 2
    STATUS_INTERNAL_SERVER_ERROR = 3
    STATUS_NOT_FOUND = 4
)



type Member struct {
    BindAddr string `json:"addr"`
    InstanceId string `json:"instance_id"`
    Port int `json:"port"`
}

type FilePart struct {
    Md5 string                  `json:"md5"`     // 分片md5
    FileSize int64              `json:"size"`    // 文件大小
}

type File struct {
    Id int                      `json:"id"`      // 分片md5
    Md5 string                  `json:"md5"`     // 分片md5
    PartNum int                 `json:"partNum"` // 文件分片数量
    Instance string             `json:"instance"`// 实例id
    Parts []FilePart            `json:"parts"`   // 实例id
}

// validate operation request.
type OperationValidationRequest struct {
    Secret string
}
// validate operation response.
type OperationValidationResponse struct {
    Status int
}


// register storage client operation request.
type OperationRegisterStorageClientRequest struct {
    BindAddr string      `json:"addr"`
    Group string         `json:"group"`
    InstanceId string    `json:"instance_id"`
    Port int             `json:"port"`
}
// validate operation response.
type OperationRegisterStorageClientResponse struct {
    Status int
    LookBackAddr string         `json:"backAddr"`   // tracker反视地址
    GroupMembers []Member       `json:"members"`    // 我的组内成员（不包括自己）
}


// upload file operation request.
type OperationUploadFileRequest struct {
    FileSize uint64 `json:"fileSize"` // 文件大小
    FileExt string `json:"ext"`      //文件扩展名，不包含'.'
    Md5 string `json:"md5"`          //文件md5, 如果已存在则不需要上传

}
// upload file response.
type OperationUploadFileResponse struct {
    Status int
    Path string `json:"path"`
}


// query file operation request.
type OperationQueryFileRequest struct {
    PathOrMd5 string `json:"md5"`          //文件md5, 如果已存在则不需要上传
}
// query file response.
type OperationQueryFileResponse struct {
    Status int
    Exist bool `json:"exist"`      // true:the file exists
    FileSize uint64 `json:"fileSize"` // 文件大小
}

// download file operation request.
type OperationDownloadFileRequest struct {
    Path string `json:"path"`
}
// download file response.
type OperationDownloadFileResponse struct {
    Status int
}


// register file operation request.
type OperationRegisterFileRequest struct {
    File File `json:"file"`    // 文件md5
}
// register file response.
type OperationRegisterFileResponse struct {
    Status int
}

// register file operation request.
type OperationPullFileRequest struct {
    LastId string `json:"lastId"`   // 上次同步的ID位置（tracker端的ID）
}
// register file response.
type OperationPullFileResponse struct {
    Status int
    File File                    `json:"file"`    // 文件md5
    Parts []FilePart             `json:"parts"`   // 文件分片
}





