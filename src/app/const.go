package app

import (
    "time"
)

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
    BIND_ADDRESS string
    TRACKERS string
    HTTP_ENABLE bool
    MIME_TYPES_ENABLE bool
    LOG_ENABLE bool
    PORT int
    HTTP_PORT int
    CLIENT_TYPE int //client类型，1：storage client，2：other client
    STORAGE_CLIENT_EXPIRE_TIME time.Duration = 60
    REG_STORAGE_INTERVAL       time.Duration = 30
    SYNC_INTERVAL              time.Duration = 5 //每5s取一次同步任务
    PATH_REGEX = "^/([0-9a-zA-Z_]{1,10})/([0-9a-zA-Z_]{1,10})/([MS])/([0-9a-fA-F]{32})$"
    UUID = ""
)

const (
    TASK_SYNC_MEMBER = 1 // storage同步自己的组内成员
    TASK_REPORT_FILE = 2
    TASK_PULL_NEW_FILE = 3
    TASK_DOWNLOAD_FILE = 4
    TASK_SYNC_ALL_STORAGES = 5 // client 同步所有的storage
    DB_Pool_SIZE = 10
)