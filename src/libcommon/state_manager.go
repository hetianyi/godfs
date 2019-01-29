package libcommon

import (
	"app"
	"container/list"
	"libcommon/bridgev2"
	"libservicev2"
	"regexp"
	"strconv"
	"sync"
	"time"
	"util/common"
	"util/logger"
	"util/timeutil"
)

var managedStorage = make(map[string]*app.StorageDO)

// in case of when:
// client disconnect and reconnect immediately, but tracker is now not expire this storage server yet,
// server will think it is not unique and return error
var hardCheckStorage = make(map[string]byte)
var managedStorageStatistics = make(map[string]*list.List)

var operationLock = *new(sync.Mutex)

// timer task: remove expired storage servers
func ExpirationDetection() {
	timer := time.NewTicker(app.StorageClientExpireTime)
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
		logger.Debug("register storage server", storage.Host+":"+strconv.Itoa(storage.Port), "("+storage.Uuid+")")
	}
	managedStorage[storage.Uuid] = storage
	queueStatistics(storage)
	return libservicev2.SaveStorage("", *storage)
}

// expire storage server in the future
func FutureExpireStorageServer(manager *bridgev2.ConnectionManager) {
	operationLock.Lock()
	defer operationLock.Unlock()
	if manager == nil {
		return
	}
	ReleaseUUIDHolder(manager.UUID)
	storage := managedStorage[manager.UUID]
	if storage != nil {
		logger.Info("expire storage server", storage.Host+":"+strconv.Itoa(storage.Port), "("+storage.Uuid+")", "in", app.StorageClientExpireTime)
		storage.ExpireTime = timeutil.GetTimestamp(time.Now().Add(app.StorageClientExpireTime))
		managedStorage[storage.Uuid] = storage
	}
}

func HoldUUID(uuid string) {
	if !IsStorageClientUUID(uuid) {
		return
	}
	operationLock.Lock()
	defer operationLock.Unlock()
	logger.Debug("hold uuid:", uuid)
	hardCheckStorage[uuid] = 1
}

func ReleaseUUIDHolder(uuid string) {
	if !IsStorageClientUUID(uuid) {
		return
	}
	logger.Debug("release uuid holder:", uuid)
	delete(hardCheckStorage, uuid)
}

// check if instance if is unique
func IsInstanceIdUnique(uuid string) bool {
	if !IsStorageClientUUID(uuid) {
		return true
	}
	// connection is from other storage server
	if app.RunWith == 1 {
		return true
	}
	operationLock.Lock()
	defer operationLock.Unlock()
	if hardCheckStorage[uuid] == 1 {
		return false
	}
	return true
}

func IsStorageClientUUID(uuid string) bool {
	if mat, e := regexp.Match("[0-9a-z]{30}", []byte(uuid)); e == nil && mat {
		return true
	}
	return false
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
func GetAllStorageServers() []app.StorageDO {
	operationLock.Lock()
	defer operationLock.Unlock()
	var members = make([]app.StorageDO, len(managedStorage))
	index := 0
	for _, v := range managedStorage {
		members[index] = *v
		index++
	}
	return members
}

func GetSyncStatistic() []app.StorageDO {
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

func collectQueueStatistics() []app.StorageDO {
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
			Disk:          v.Disk,
			Download:      v.Download,
			Upload:        v.Upload,
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
	ret := make([]app.StorageDO, temp.Len())
	i := 0
	for ele := temp.Front(); ele != nil; ele = ele.Next() {
		ret[i] = ele.Value.(app.StorageDO)
		i++
	}
	return ret
}
