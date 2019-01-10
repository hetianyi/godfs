package bridgev2

import "app"

// operation meta for connect/validate
type ConnectMeta struct {
    Secret string `json:"secret"`
    UUID   string `json:"uuid"` // this is client uuid
}

// operation meta for connect/validate
type ConnectResponseMeta struct {
    UUID        string `json:"uuid"` // this is server uuid
    New4Tracker bool   `json:"new"`
}

// register storage client operation request
type SyncStorageMembersMeta struct {
    UUID          string `json:"uuid"`
    Group         string `json:"group"`
    InstanceId    string `json:"instance_id"`
    AdvertiseAddr string `json:"addr"`
    AdvertisePort int    `json:"advertise_port"`
    Port          int    `json:"port"`
    HttpPort      int    `json:"httpPort"`
    HttpEnable    bool   `json:"httpEnable"`
    // statistic info
    TotalFiles int   `json:"files"`
    Finish     int   `json:"finish"`
    StartTime  int64 `json:"startTime"`
    LogTime    int64 `json:"logTime"`
    Downloads  int   `json:"downloads"`
    Uploads    int   `json:"uploads"`
    IOin       int64 `json:"in"`
    IOout      int64 `json:"out"`

    StageDownloads int   `json:"stageDownloads"`
    StageUploads   int   `json:"stageUploads"`
    StageIOin      int64 `json:"stageIOin"`
    StageIOout     int64 `json:"stageIOout"`

    DiskUsage int64  `json:"disk"`
    Memory    uint64 `json:"mem"`
    ReadOnly  bool   `json:"readonly"`

    // add at 2019/01/02
    // consider using in docker stack environment
    // if 'advertise_addr' is not satisfied for docker stack environment when a group has multiple instances,
    // then the client(include storage client, java client and native client) can only get a single address of a group,
    // this address is usually the 'advertise_addr' parameter specified in docker compose file.
    // client always use LookBackAddress to synchronize files between each other first, the 'advertise_addr' is secondary choice.
    Host string `json:"host"`
}

// validate operation response.
type SyncStorageMembersResponseMeta struct {
    LookBackAddr string   `json:"backAddr"` // tracker lookback addr
    GroupMembers []app.Member `json:"members"`  // group members(not include self)
}





