package libcommon

import (
	"app"
	"container/list"
	"libcommon/bridge"
	"strconv"
	"sync"
	"time"
	"util/common"
	"util/logger"
	"util/timeutil"
)

var managedStorage = make(map[string]*app.Member)
var managedStorageStatistics = make(map[string]*list.List)

var operationLock = *new(sync.Mutex)


// timer task: remove expired storage servers
func ExpirationDetection() {
	timer := time.NewTicker(app.STORAGE_CLIENT_EXPIRE_TIME)
	for {
		<-timer.C
		logger.Debug("detecting expiration")
		curTime := timeutil.GetTimestamp(time.Now())
		operationLock.Lock()
		common.Try(func() {
			for k, v := range managedStorage {
				if v.ExpireTime <= curTime { // expired
					delete(managedStorage, k)
					logger.Info("storage server:", k, "expired finally")
				}
			}
		}, func(i interface{}) {})
		operationLock.Unlock()
	}
}

// 添加storage服务器
func SaveStorageServer(member *app.Member) {
	operationLock.Lock()
	defer operationLock.Unlock()

	member.ExpireTime = timeutil.GetTimestamp(time.Now().Add(time.Hour * 8760000))
	if managedStorage[member.Uuid] == nil {
		logger.Debug("register storage server", member.Host + ":" + strconv.Itoa(member.Port), "("+ member.Uuid +")")
	}
	managedStorage[member.Uuid] = member
	queueStatistics(member)
}

// expire storage server in the future
func FutureExpireStorageServer(member *app.Member) {
	operationLock.Lock()
	defer operationLock.Unlock()
	if member != nil {
		logger.Info("expire storage server", member.Host + ":" + strconv.Itoa(member.Port), "("+ member.Uuid +")", "in", app.STORAGE_CLIENT_EXPIRE_TIME)
		member.ExpireTime = timeutil.GetTimestamp(time.Now().Add(app.STORAGE_CLIENT_EXPIRE_TIME))
		managedStorage[member.Uuid] = member
	}
}

// check if instance if is unique
func IsInstanceIdUnique(member *app.Member) bool {
	operationLock.Lock()
	defer operationLock.Unlock()
	if managedStorage[member.Uuid] != nil {
		return false
	}
	for _, v := range managedStorage {
		if v.Uuid == member.Uuid {
			return false
		}
	}
	return true
}

// fetch group members
func GetGroupMembers(member *app.Member) []app.Member {
	operationLock.Lock()
	defer operationLock.Unlock()
	var mList list.List
	for k, v := range managedStorage {
		if k != member.Uuid && v.Group == member.Group { // 过期
			mList.PushBack(*v)
		}
	}
	var members = make([]app.Member, mList.Len())
	index := 0
	for e := mList.Front(); e != nil; e = e.Next() {
		members[index] = e.Value.(app.Member)
		index++
	}
	return members
}

// fetch all storage server for client
func GetAllStorages() []app.Member {
	operationLock.Lock()
	defer operationLock.Unlock()
	var members = make([]app.Member, len(managedStorage))
	index := 0
	for _, v := range managedStorage {
		members[index] = *v
	}
	return members
}

func GetSyncStatistic() []bridge.ServerStatistic {
	return collectQueueStatistics()
}


func queueStatistics(member *app.Member) {
	if member == nil {
		return
	}
	ls := managedStorageStatistics[member.Uuid]
	if ls == nil {
		logger.Debug("statistic queue is null, create new list")
		ls = list.New()
		managedStorageStatistics[member.Uuid] = ls
	}
	if ls.Len() >= 10 {
		logger.Debug("statistic queue full, remove head")
		ls.Remove(ls.Front())
	}
	ls.PushBack(member)
}

func collectQueueStatistics() []bridge.ServerStatistic {
	operationLock.Lock()
	defer operationLock.Unlock()
	var temp list.List
	for _, ls := range managedStorageStatistics {
		if ls == nil || ls.Len() == 0 {
			continue
		}
		v := ls.Remove(ls.Front()).(*app.Member)
		item := app.CachedStorageServerStatistic{
			UUID:          v.Uuid,
			AdvertiseAddr: v.Host,
			Group:         v.Group,
			InstanceId:    v.InstanceId,
			Port:          v.Port,
			HttpPort:      v.HttpPort,
			HttpEnable:    v.HttpEnable,
			TotalFiles:    v.T,
			Finish:        v.Finish,
			IOin:          v.IOin,
			IOout:         v.IOout,
			DiskUsage:     v.DiskUsage,
			Downloads:     v.Downloads,
			Uploads:       v.Uploads,
			StartTime:     v.StartTime,
			Memory:        v.Memory,
			ReadOnly:      v.ReadOnly,

			LogTime:        v.LogTime,
			StageDownloads: v.StageDownloads,
			StageUploads:   v.StageUploads,
			StageIOin:      v.StageIOin,
			StageIOout:     v.StageIOout,
		}
		temp.PushBack(item)
	}
	ret := make([]bridge.ServerStatistic, temp.Len())
	i := 0
	for ele := temp.Front(); ele != nil; ele = ele.Next() {
		ret[i] = ele.Value.(bridge.ServerStatistic)
		i++
	}
	return ret
}
