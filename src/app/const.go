package app

const (
    BUFF_SIZE = 1024*30 // byte buffer size set to 30kb
)
var (
    ASSIGN_DISK_SPACE int64
    SLICE_SIZE int64
    LOG_INTERVAL string    //log文件精度：h/d/w/m/y
    BASE_PATH string
    GROUP string
    INSTANCE_ID string
)