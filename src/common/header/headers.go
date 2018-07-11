package header

// Definition some data headers for data transfer.
// [operation][meta][body...]
type UploadHeader struct {
    operation [2]byte   // operations such as upload, sync, communication.
                        // 1.   1 1     和tracker通信头
                        // 2.   2 1     客户端上传通信头
                        // 3.   3 1     组内同步通信头
    metaLen [10]byte    // meta length for read meta.
                        // meta represents a json string.
    bodyLen [10]byte    // body length for read body.
                        // if operations is communication then no body.
}

type UploadMetaData struct {
    Mode int `json:"mode"`
    FileSize int64 `json:"fileSize"`
}

type CommunicationMetaData struct {
    Mode int `json:"mode"`
    Operation int `json:"operation"`// register service, register file
    BindAddr int `json:"addr"`
    Port int `json:"port"`
}

