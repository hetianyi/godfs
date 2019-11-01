package svc

import (
	"container/list"
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/hetianyi/godfs/api"
	"github.com/hetianyi/godfs/binlog"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/logger"
	"github.com/hetianyi/gox/timer"
	"sync"
	"time"
)

var (
	// storage servers who is being watching.
	watchingMembers map[string]*common.Server
	// binlog synchronization state of all storage servers.
	synchronizationState map[string]*common.BinlogQueryDTO
	configChanged        = false
	configKeyPrefix      = "binlogSynchronizationState:"
	syncLock             *sync.Mutex
	configChangeLock     *sync.Mutex
)

func init() {
	watchingMembers = make(map[string]*common.Server)
	synchronizationState = make(map[string]*common.BinlogQueryDTO)
	syncLock = new(sync.Mutex)
	configChangeLock = new(sync.Mutex)
}

func updateConfigChangeState(changed bool) {
	configChangeLock.Lock()
	defer configChangeLock.Unlock()

	if changed || (!changed && !configChanged) {
		configChanged = changed
		return
	}
}

func loadSynchronizationConfig(instanceId string) (*common.BinlogQueryDTO, error) {

	if synchronizationState[instanceId] != nil {
		return synchronizationState[instanceId], nil
	}

	config := common.GetConfigMap()
	bs, err := config.GetConfig(configKeyPrefix + instanceId)
	if err != nil {
		return nil, err
	}
	ret := &common.BinlogQueryDTO{}
	if json.Unmarshal(bs, ret); err != nil {
		return nil, err
	}
	synchronizationState[instanceId] = ret
	return ret, nil
}

func checkServer(instanceId string) bool {
	syncLock.Lock()
	defer syncLock.Unlock()

	if watchingMembers[instanceId] != nil {
		return true
	}
	return false
}

func InitStorageMemberBinlogWatcher() {
	syncLock.Lock()
	defer syncLock.Unlock()

	// timer task: save synchronization state every second.
	timer.Start(time.Second*5, time.Second, 0, func(t *timer.Timer) {

		if !configChanged || len(synchronizationState) == 0 {
			return
		}

		config := common.GetConfigMap()

		err := config.BatchUpdateConfig(func(b *bolt.Bucket) error {
			for k, v := range synchronizationState {
				bs, err := json.Marshal(v)
				if err != nil {
					logger.Debug(err)
					continue
				}
				if err := b.Put([]byte(configKeyPrefix+k), bs); err != nil {
					logger.Debug("error save config")
					continue
				}
			}
			return nil
		})

		if err != nil {
			logger.Error(err)
		} else {
			logger.Debug("save synchronization state success")
		}
		updateConfigChangeState(false)
	})

	// timer task: check storage server instances
	timer.Start(time.Second*5, common.SYNCHRONIZE_INTERVAL, 0, func(t *timer.Timer) {
		ss := api.FilterInstances(common.ROLE_STORAGE)
		if ss == nil || ss.Len() == 0 {
			return
		}

		sameGroupInstances := list.New()
		expiredInstances := list.New()
		// filter group
		for ele := ss.Front(); ele != nil; ele = ele.Next() {
			if ele.Value.(*common.Instance).Attributes["group"] == common.InitializedStorageConfiguration.Group &&
				ele.Value.(*common.Instance).InstanceId != common.InitializedStorageConfiguration.InstanceId {
				sameGroupInstances.PushBack(ele.Value)
			}
		}

		for k, v := range watchingMembers {
			c := false
			gox.WalkList(sameGroupInstances, func(item interface{}) bool {
				if item.(*common.Instance).InstanceId == k {
					c = true
					return true
				}
				return false
			})
			if !c {
				expiredInstances.PushBack(&v)
			}
		}
		gox.WalkList(expiredInstances, func(item interface{}) bool {
			unWatch(item.(*common.Server))
			return false
		})
		gox.WalkList(sameGroupInstances, func(item interface{}) bool {
			watch(&item.(*common.Instance).Server)
			return false
		})
	})
}

func watch(server *common.Server) {
	syncLock.Lock()
	defer syncLock.Unlock()

	logger.Debug("watching storage server: ",
		server.ConnectionString(), "(", server.InstanceId, ")")

	if watchingMembers[server.InstanceId] != nil {
		logger.Debug("storage server is already in watch: ",
			server.ConnectionString(), "(", server.InstanceId, ")")
		return
	}

	watchingMembers[server.InstanceId] = server

	timer.Start(0, time.Second*5, 0, func(t *timer.Timer) {
		for true {
			// make sure this member is not expired.
			if !checkServer(server.InstanceId) {
				t.Destroy()
				break
			}

			// get state
			config, err := loadSynchronizationConfig(server.InstanceId)
			if err != nil {
				logger.Debug("error load synchronization config: ", err)
				break
			}

			ret, err := clientAPI.SyncBinlog(server, config)
			if err != nil {
				logger.Debug("error synchronize binlog from storage server: ",
					server.ConnectionString(), "(", server.InstanceId, "): ", err)
				break
			}

			if ret.FileIndex == config.FileIndex && ret.Offset == config.Offset {
				logger.Debug("nothing changed")
				break
			}

			logger.Debug("synchronize ", len(ret.Logs), " binlogs from ",
				server.ConnectionString(), "(", server.InstanceId, ")")

			failed := 0
			var lastErr error
			for _, v := range ret.Logs {
				if v.SourceInstance == common.InitializedStorageConfiguration.InstanceId {
					// binlog is mime, so skip.
					continue
				}
				c, err := Contains(v.FileId)
				if err != nil {
					failed++
					lastErr = err
					logger.Debug("error query local binlog: ", v.FileId, ":", err)
					continue
				}
				if !c {
					if err = writableBinlogManager.Write(binlog.CreateLocalBinlog(v.FileId,
						v.FileLength, v.SourceInstance, time.Now(), 0)); err != nil {
						failed++
						lastErr = err
						logger.Debug("error write binlog: ", err)
					}
					// write to dataset
					// TODO write when file synchronize success.
				} else {
					logger.Debug("binlog already exists: ", v.FileId)
				}
			}

			if failed == 0 {
				config.Offset = ret.Offset
				config.FileIndex = ret.FileIndex
				updateConfigChangeState(true)
				logger.Debug("binlog write success")
			} else {
				logger.Debug("binlog write error: ", lastErr, ", failed ", failed)
			}
			if len(ret.Logs) == 0 {
				break
			}
		}
	})

}

func unWatch(server *common.Server) {
	syncLock.Lock()
	defer syncLock.Unlock()

	delete(watchingMembers, server.InstanceId)
}
