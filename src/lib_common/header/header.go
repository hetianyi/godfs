package header

// operation code mapped first 2 bytes
var OperationHeadByteMap = make(map[int] []byte)

func init() {
    OperationHeadByteMap[0] = []byte{1,1}   //for tracker server: register storage client //add upload client in the future
    OperationHeadByteMap[1] = []byte{2,2}   //for tracker server: register file from client
    OperationHeadByteMap[2] = []byte{3,1}   //for storage server: upload file from client
    OperationHeadByteMap[4] = []byte{3,2}   //for all kinds of client        : response
    OperationHeadByteMap[8] = []byte{3,3}   //for all kinds of first request : connection validate
    OperationHeadByteMap[3] = []byte{4,2}   //for storage server: sync file from another storage server
    OperationHeadByteMap[5] = []byte{4,3}   //for storage server: query if file exists
    OperationHeadByteMap[6] = []byte{4,4}   //for storage server: download file



}


// 当连接初次建立时发送的头部信息，用于校验secret
type ConnectionHead struct {
    Secret string   `json:"secret"`  // 通信秘钥
}

// 当连接初次建立时发送的头部信息，用于校验secret
type ConnectionHeadResponse struct {
    Status int                  `json:"status"`     // 状态
                                                    // 0:success
                                                    // 1:bad secret
                                                    // 2:operation not support
                                                    // 3:server failed, will not close connection
}


type Member struct {
    BindAddr string `json:"addr"`
    InstanceId string `json:"instance_id"`
    Port int `json:"port"`
}

// 客户端上传文件到storage的meta
type UploadRequestMeta struct {
    FileSize int64 `json:"fileSize"` // 文件大小
    FileExt string `json:"ext"`      //文件扩展名，不包含'.'
    Md5 string `json:"md5"`          //文件md5, 如果已存在则不需要上传
}

// upload finish response meta
type UploadResponseMeta struct {
    Status int                  `json:"status"`     // 状态
                                                    // 0:success
                                                    // 1:bad secret
                                                    // 2:operation not support
                                                    // 3:server failed, will not close connection
    Path string `json:"path"`
    FileSize int64 `json:"size"`
    Exist bool                  `json:"exist"`      // true:the file exists
                                                    // true:the file does not exists
}


// 客户端查询文件是否存在的meta
type QueryFileRequestMeta struct {
    PathOrMd5 string `json:"md5"`          //文件md5, 如果已存在则不需要上传
}
// 客户端查询文件是否存在的meta
type QueryFileResponseMeta struct {
    Status int                  `json:"status"`     // 状态
                                                    // 0:success
                                                    // 1:bad secret
                                                    // 2:operation not support
                                                    // 3:server failed, will not close connection
    Exist bool                  `json:"exist"`      // true:the file exists
                                                    // true:the file does not exists
}

// 客户端下载文件的meta
type DownloadFileRequestMeta struct {
    Path string `json:"path"`
}
// 客户端下载文件的meta
type DownloadFileResponseMeta struct {
    Status int                  `json:"status"`     // 状态
                                                    // 0:success
                                                    // 1:bad secret
                                                    // 2:operation not support
                                                    // 3:server failed, will not close connection
                                                    // 4:file not found
}




// storage将自己注册到tracker的meta
// TODO add more statistic info about storage server
type CommunicationRegisterStorageRequestMeta struct {
    Secret string        `json:"secret"`  // 通信秘钥
    BindAddr string      `json:"addr"`
    Group string         `json:"group"`
    InstanceId string    `json:"instance_id"`
    Port int             `json:"port"`
}

// tracker响应storage注册自己的meta
type CommunicationRegisterStorageResponseMeta struct {
    Status int                  `json:"status"`     // 状态
                                                    // 0:success
                                                    // 1:bad secret
                                                    // 2:operation not support
                                                    // 3:server failed, will not close connection
    LookBackAddr string         `json:"backAddr"`   // tracker反视地址
    GroupMembers []Member       `json:"members"`    // 我的组内成员（不包括自己）
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

// storage将文件注册到tracker的meta
type CommunicationRegisterFileRequestMeta struct {
    File File                    `json:"file"`    // 文件md5
}

// storage将文件注册到tracker的meta
type CommunicationRegisterFileResponseMeta struct {
    status int                   `json:"status"`   // 状态，0：成功，其他失败
}

type CommunicationPullFileRequestMeta struct {
    LastId string                `json:"lastId"`   // 上次同步的ID位置（tracker端的ID）
}
type CommunicationPullFileResponseMeta struct {
    File File                    `json:"file"`    // 文件md5
    Parts []FilePart             `json:"parts"`   // 文件分片
}


