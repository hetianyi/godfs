package lib_tracker

import (
    "time"
    "util/logger"
    "sync"
    "common/header"
    "strconv"
    "container/list"
)

var managedStorages = make(map[string] *storageMeta)

var operationLock = *new(sync.Mutex)

type storageMeta struct {
    expireTime int64
    group string
    host string
    port int
}

// 定时任务，剔除过期的storage服务器
func ExpirationDetection() {
    operationLock.Lock()
    defer operationLock.Unlock()
    timer := time.NewTimer(time.Second * 30)
    for {
        <-timer.C
        logger.Info("exec expired detected")
        curTime := time.Now().UnixNano() / 1e6
        for k, v := range managedStorages {
            if v.expireTime <= curTime { // 过期
                delete(managedStorages, k)
                logger.Info("storage server:", k, "expired")
            }
        }
    }

}

// 添加storage服务器
func AddStorageServer(host string, port int, group string) {
    operationLock.Lock()
    defer operationLock.Unlock()
    meta := &storageMeta{expireTime: time.Now().UnixNano() / 1e6, group: group, host: host, port: port}
    managedStorages[host + ":" + strconv.Itoa(port)] = meta
}

// 获取组内成员
func GetGroupMembers(host string, port int, group string) *list.List {
    me := host + ":" + strconv.Itoa(port)
    members := list.New()
    for k, v := range managedStorages {
        if k != me && v.group == group { // 过期
            m := header.Member{BindAddr: v.host, Port: v.port}
            members.PushBack(m)
            logger.Info("storage server:", k, "expired")
        }
    }
    return members
}

