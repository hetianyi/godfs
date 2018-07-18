package lib_tracker

import (
    "time"
    "util/logger"
    "sync"
    "strconv"
    "lib_common/header"
    "util/timeutil"
    "container/list"
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
    timer := time.NewTicker(time.Second * 30)
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
func AddStorageServer(meta *header.CommunicationRegisterStorageRequestMeta) {
    operationLock.Lock()
    defer operationLock.Unlock()
    key := meta.BindAddr + ":" + strconv.Itoa(meta.Port)
    holdMeta := &storageMeta{
        ExpireTime: timeutil.GetTimestamp(time.Now().Add(time.Minute)),
        Group: meta.Group,
        InstanceId: meta.InstanceId,
        Host: meta.BindAddr,
        Port: meta.Port,
    }
    managedStorages[key] = holdMeta
    //js, _ := json.Marshal(*managedStorages[key])
    //fmt.Println(string(js))
}

// check if instance if is unique
func IsInstanceIdUnique(meta *header.CommunicationRegisterStorageRequestMeta) bool {
    key := meta.BindAddr + ":" + strconv.Itoa(meta.Port)
    for k, v := range managedStorages {
        if k != key && v.Group == meta.Group && v.InstanceId == meta.InstanceId {
            return false
        }
    }
    return true
}

// 获取组内成员
func GetGroupMembers(meta *header.CommunicationRegisterStorageRequestMeta) []header.Member {
    key := meta.BindAddr + ":" + strconv.Itoa(meta.Port)
    var mList list.List
    for k, v := range managedStorages {
        if k != key && v.Group == meta.Group { // 过期
            m := header.Member{BindAddr: v.Host, Port: v.Port, InstanceId: v.InstanceId}
            mList.PushBack(m)
        }
    }
    var members = make([]header.Member, mList.Len())
    index := 0
    for e := mList.Front(); e != nil; e = e.Next() {
        members[index] = e.Value.(header.Member)
        index++
    }
    return members
}

