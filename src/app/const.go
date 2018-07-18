package app

const (
    BUFF_SIZE = 1024*30 // byte buffer size set to 30kb
)
var (
    RUN_WITH int            //启动模式，1：storage，2：tracker
    ASSIGN_DISK_SPACE int64
    SLICE_SIZE int64
    LOG_INTERVAL string    //log文件精度：h/d/w/m/y
    BASE_PATH string
    GROUP string
    INSTANCE_ID string
    HTTP_ENABLE bool
    MIME_TYPES_ENABLE bool
    HTTP_PORT int
)