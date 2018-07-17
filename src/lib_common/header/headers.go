package header

import "container/list"

// operation code mapped first 2 bytes
var OperationHeadByteMap = make(map[int] []byte)

func init() {
    OperationHeadByteMap[0] = []byte{1,1}   //for tracker server: register storage client
    OperationHeadByteMap[1] = []byte{2,2}   //for tracker server: register file from client
    OperationHeadByteMap[2] = []byte{3,1}   //for storage server: upload file from client
    OperationHeadByteMap[4] = []byte{3,2}   //for all kinds of client        : response
    OperationHeadByteMap[3] = []byte{4,2}   //for storage server: sync file from another storage server
    OperationHeadByteMap[5] = []byte{4,3}   //for storage server: query if file exists
    OperationHeadByteMap[6] = []byte{4,4}   //for storage server: download file


}



type Member struct {
    BindAddr string `json:"addr"`
    Port int `json:"port"`
}

// 客户端上传文件到storage的meta
type UploadRequestMeta struct {
    Secret string   `json:"secret"`  // 通信秘钥
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
    Secret string   `json:"secret"`  // 通信秘钥
    Md5 string `json:"md5"`          //文件md5, 如果已存在则不需要上传
}

// 客户端下载文件的meta
type DownloadFileRequestMeta struct {
    Secret string   `json:"secret"`  // 通信秘钥
    Path string `json:"path"`
}




// storage将自己注册到tracker的meta
type CommunicationRegisterStorageRequestMeta struct {
    Secret string   `json:"secret"`  // 通信秘钥
    BindAddr string `json:"addr"`
    Group string    `json:"group"`
    Port int        `json:"port"`
}
// tracker响应storage注册自己的meta
type CommunicationRegisterStorageResponseMeta struct {
    Status int                  `json:"status"`     // 状态
                                                    // 0:success
                                                    // 1:bad secret
    LookBackAddr string         `json:"backAddr"`   // tracker反看地址
    GroupMembers *list.List     `json:"members"`    // 我的组内成员（不包括自己）
}

// storage将文件注册到tracker的meta
type CommunicationRegisterFileRequestMeta struct {
    BindAddr int `json:"addr"`
    Port int `json:"port"`
}

