package header

import "container/list"

var (
    COM_REG_STORAGE = []byte{1,1}
    COM_REG_FILE = []byte{1,1}
    COM_UPLOAD_FILE = []byte{2,1}
)


type Member struct {
    BindAddr string `json:"addr"`
    Port int `json:"port"`
}

// 客户端上传文件到storage的meta
type UploadRequestMeta struct {
    Secret string   `json:"secret"`  // 通信秘钥
    FileSize int64 `json:"fileSize"`
}

// 客户端上传文件到storage的meta
type ResponseMeta struct {
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

