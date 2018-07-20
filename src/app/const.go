package app

import "time"

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
    SECRET string
    HTTP_ENABLE bool
    MIME_TYPES_ENABLE bool
    HTTP_PORT int
    STORAGE_CLIENT_EXPIRE_TIME time.Duration = 60
    REG_STORAGE_INTERVAL       time.Duration = 30
    SYNC_INTERVAL              time.Duration = 5 //每5s取一次同步任务
    PATH_REGEX = "^/([0-9a-zA-Z_]{1,10})/([0-9a-zA-Z_]{1,10})/([MS])/([0-9a-fA-F]{32})$"
)