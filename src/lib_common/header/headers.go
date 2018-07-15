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


}



type Member struct {
    BindAddr string `json:"addr"`
    Port int `json:"port"`
}

// 客户端上传文件到storage的meta
type UploadRequestMeta struct {
    Secret string   `json:"secret"`  // 通信秘钥
    FileSize int64 `json:"fileSize"`
}

// upload finish response meta
type UploadResponseMeta struct {
    Status int                  `json:"status"`     // 状态
                                                    // 0:success
                                                    // 1:bad secret
                                                    // 2:operation not support
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

