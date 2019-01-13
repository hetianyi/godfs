package libcommon

import (
	"app"
	"container/list"
	"libcommon/bridge"
	"libservicev2"
	"strconv"
	"sync"
	"time"
	"util/common"
	"util/logger"
	"util/timeutil"
)

var managedStorage = make(map[string]*app.StorageDO)
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

// cache registered storage server
func CacheStorageServer(storage *app.StorageDO) error {
	operationLock.Lock()
	defer operationLock.Unlock()

	storage.ExpireTime = timeutil.GetTimestamp(time.Now().Add(time.Hour * 87600))
	if managedStorage[storage.Uuid] == nil {
		logger.Debug("register storage server", storage.Host + ":" + strconv.Itoa(storage.Port), "("+ storage.Uuid +")")
	}
	managedStorage[storage.Uuid] = storage
	queueStatistics(storage)
	return libservicev2.SaveStorage("", storage, nil)
}

// expire storage server in the future
func FutureExpireStorageServer(storage *app.StorageDO) {
	operationLock.Lock()
	defer operationLock.Unlock()
	if storage != nil {
		logger.Info("expire storage server", storage.Host + ":" + strconv.Itoa(storage.Port), "("+ storage.Uuid +")", "in", app.STORAGE_CLIENT_EXPIRE_TIME)
		storage.ExpireTime = timeutil.GetTimestamp(time.Now().Add(app.STORAGE_CLIENT_EXPIRE_TIME))
		managedStorage[storage.Uuid] = storage
	}
}

// check if instance if is unique
func IsInstanceIdUnique(storage *app.StorageDO) bool {
	operationLock.Lock()
	defer operationLock.Unlock()
	if managedStorage[storage.Uuid] != nil {
		return false
	}
	for _, v := range managedStorage {
		if v.Uuid == storage.Uuid {
			return false
		}
	}
	return true
}

// fetch group members
func GetGroupMembers(storage *app.StorageDO) []app.StorageDO {
	operationLock.Lock()
	defer operationLock.Unlock()
	var mList list.List
	for k, v := range managedStorage {
		if k != storage.Uuid && v.Group == storage.Group { // 过期
			mList.PushBack(*v)
		}
	}
	var members = make([]app.StorageDO, mList.Len())
	index := 0
	for e := mList.Front(); e != nil; e = e.Next() {
		members[index] = e.Value.(app.StorageDO)
		index++
	}
	return members
}

// fetch all storage server for client
func GetAllStorages() []app.StorageDO {
	operationLock.Lock()
	defer operationLock.Unlock()
	var members = make([]app.StorageDO, len(managedStorage))
	index := 0
	for _, v := range managedStorage {
		members[index] = *v
	}
	return members
}

func GetSyncStatistic() []bridge.ServerStatistic {
	return collectQueueStatistics()
}


func queueStatistics(storage *app.StorageDO) {
	if storage == nil {
		return
	}
	ls := managedStorageStatistics[storage.Uuid]
	if ls == nil {
		logger.Debug("statistic queue is null, create new list")
		ls = list.New()
		managedStorageStatistics[storage.Uuid] = ls
	}
	if ls.Len() >= 10 {
		logger.Debug("statistic queue full, remove head")
		ls.Remove(ls.Front())
	}
	ls.PushBack(storage)
}

func collectQueueStatistics() []bridge.ServerStatistic {
	operationLock.Lock()
	defer operationLock.Unlock()
	var temp list.List
	for _, ls := range managedStorageStatistics {
		if ls == nil || ls.Len() == 0 {
			continue
		}
		v := ls.Remove(ls.Front()).(*app.StorageDO)
		item := app.StorageDO{
			Uuid:          v.Uuid,
			AdvertiseAddr: v.Host,
			Group:         v.Group,
			InstanceId:    v.InstanceId,
			Port:          v.Port,
			HttpPort:      v.HttpPort,
			HttpEnable:    v.HttpEnable,
			TotalFiles:    v.TotalFiles,
			Finish:        v.Finish,
			IOin:          v.IOin,
			IOout:         v.IOout,
			Disk:     v.Disk,
			Download:     v.Download,
			Upload:       v.Upload,
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
