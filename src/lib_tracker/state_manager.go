package lib_tracker

import (
    "time"
    "util/logger"
    "sync"
    "strconv"
    "util/timeutil"
    "container/list"
    "app"
    "lib_common/bridge"
    "encoding/json"
)

var managedStorages = make(map[string] *storageMeta)

var operationLock = *new(sync.Mutex)

type storageMeta struct {
    ExpireTime int64
    Group string
    InstanceId string
    Host string
    Port int
}

// 定时任务，剔除过期的storage服务器
func ExpirationDetection() {
    timer := time.NewTicker(time.Second * app.STORAGE_CLIENT_EXPIRE_TIME)
    for {
        <-timer.C
        logger.Debug("exec expired detected")
        curTime := time.Now().UnixNano() / 1e6
        for k, v := range managedStorages {
            if v.ExpireTime <= curTime { // 过期
                delete(managedStorages, k)
                logger.Info("storage server:", k, "expired")
            }
        }
    }
}

// 添加storage服务器
func AddStorageServer(meta *bridge.OperationRegisterStorageClientRequest) {
    operationLock.Lock()
    defer operationLock.Unlock()
    key := meta.BindAddr + ":" + strconv.Itoa(meta.Port)
    holdMeta := &storageMeta{
        ExpireTime: timeutil.GetTimestamp(time.Now().Add(time.Hour * 876000)),//set to 100 years
        Group: meta.Group,
        InstanceId: meta.InstanceId,
        Host: meta.BindAddr,
        Port: meta.Port,
    }
    logger.Debug("register storage server:", key)
    managedStorages[key] = holdMeta
    //js, _ := json.Marshal(*managedStorages[key])
    //fmt.Println(string(js))
}

// 执行即将过期storage服务器
// 通常是storage客户端和tracker服务器断开连接时
func FutureExpireStorageServer(meta *bridge.OperationRegisterStorageClientRequest) {
    operationLock.Lock()
    defer operationLock.Unlock()
    if meta != nil {
        s,_ := json.Marshal(meta)
        logger.Info("expire storage client:", s)
        key := meta.BindAddr + ":" + strconv.Itoa(meta.Port)
        holdMeta := &storageMeta{
            ExpireTime: timeutil.GetTimestamp(time.Now().Add(time.Second * app.STORAGE_CLIENT_EXPIRE_TIME)),//set to 100 years
            Group: meta.Group,
            InstanceId: meta.InstanceId,
            Host: meta.BindAddr,
            Port: meta.Port,
        }
        managedStorages[key] = holdMeta
    }
}

// check if instance if is unique
func IsInstanceIdUnique(meta *bridge.OperationRegisterStorageClientRequest) bool {
    key := meta.BindAddr + ":" + strconv.Itoa(meta.Port)
    for k, v := range managedStorages {
        if k != key && v.Group == meta.Group && v.InstanceId == meta.InstanceId {
            return false
        }
    }
    return true
}

// 获取组内成员
func GetGroupMembers(meta *bridge.OperationRegisterStorageClientRequest) []bridge.Member {
    key := meta.BindAddr + ":" + strconv.Itoa(meta.Port)
    var mList list.List
    for k, v := range managedStorages {
        if k != key && v.Group == meta.Group { // 过期
            m := bridge.Member{BindAddr: v.Host, Port: v.Port, InstanceId: v.InstanceId, Group: v.Group}
            mList.PushBack(m)
        }
    }
    var members = make([]bridge.Member, mList.Len())
    index := 0
    for e := mList.Front(); e != nil; e = e.Next() {
        members[index] = e.Value.(bridge.Member)
        index++
    }
    return members
}


// 获取组内成员
func GetAllStorages() []bridge.Member {
    var mList list.List
    for _, v := range managedStorages {
        m := bridge.Member{BindAddr: v.Host, Port: v.Port, InstanceId: v.InstanceId, Group: v.Group}
        mList.PushBack(m)
    }
    var members = make([]bridge.Member, mList.Len())
    index := 0
    for e := mList.Front(); e != nil; e = e.Next() {
        members[index] = e.Value.(bridge.Member)
        index++
    }
    return members
}

