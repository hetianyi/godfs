package svc

import (
	"container/list"
	"github.com/boltdb/bolt"
	"github.com/hetianyi/godfs/api"
	"github.com/hetianyi/godfs/binlog"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/logger"
	"github.com/hetianyi/gox/timer"
	json "github.com/json-iterator/go"
	"sync"
	"time"
)

var (
	// storage servers who is being watching.
	watchingMembers map[string]*common.Server
	// binlog synchronization state of all storage servers.
	synchronizationState map[string]*common.BinlogQueryDTO
	synchronizationFlag  map[string]int64
	configKeyPrefix      = "binlogSynchronizationState:"
	syncLock             *sync.Mutex
	configChangeLock     *sync.Mutex
)

func init() {
	watchingMembers = make(map[string]*common.Server)
	synchronizationState = make(map[string]*common.BinlogQueryDTO)
	synchronizationFlag = make(map[string]int64)
	syncLock = new(sync.Mutex)
	configChangeLock = new(sync.Mutex)
}

// updateConfigChangeState updates synchronization state of each instance.
func updateConfigChangeState(instanceId string, clear bool, isLocked bool) {
	if !isLocked {
		configChangeLock.Lock()
		defer configChangeLock.Unlock()
	}

	if clear {
		for k := range synchronizationFlag {
			synchronizationFlag[k] = 0
		}
	} else {
		synchronizationFlag[instanceId] = synchronizationFlag[instanceId] + 1
	}
}

// loadSynchronizationConfig queries synchronization state of the instance.
func loadSynchronizationConfig(instanceId string) (*common.BinlogQueryDTO, error) {
	syncLock.Lock()
	defer syncLock.Unlock()

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

// checkServer checks if the server is still alive.
func checkServer(instanceId string) bool {
	syncLock.Lock()
	defer syncLock.Unlock()

	if watchingMembers[instanceId] != nil {
		return true
	}
	return false
}

// InitStorageMemberBinlogWatcher initializes timer jobs for binlog and file synchronization.
func InitStorageMemberBinlogWatcher() {
	syncLock.Lock()
	defer syncLock.Unlock()

	InitFileSynchronization()

	// timer task: save synchronization state every second.
	timer.Start(time.Second*10, time.Second, 0, func(t *timer.Timer) {
		configChanged := false
		for _, v := range synchronizationFlag {
			if v > 0 {
				configChanged = true
				break
			}
		}
		if !configChanged {
			return
		}

		configChangeLock.Lock()
		defer configChangeLock.Unlock()

		config := common.GetConfigMap()
		err := config.BatchUpdate(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(common.BUCKET_KEY_CONFIGMAP))
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
		updateConfigChangeState("", true, true)
	})

	// timer task: check and watch storage server instances
	timer.Start(time.Second*5, common.SYNCHRONIZE_INTERVAL, 0, func(t *timer.Timer) {
		ss := filterGroupMembers(api.FilterInstances(common.ROLE_STORAGE),
			common.InitializedStorageConfiguration.Group)
		if ss == nil || ss.Len() == 0 {
			return
		}

		expiredInstances := list.New()

		for k, v := range watchingMembers {
			c := false
			gox.WalkList(ss, func(item interface{}) bool {
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
		gox.WalkList(ss, func(item interface{}) bool {
			watch(&item.(*common.Instance).Server)
			return false
		})
	})
}

// watch starts to watch a single storage member server.
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

	binlogList := list.New()

	timer.Start(0, time.Second*10, 0, func(t *timer.Timer) {
		for true {

			// clear the list
			util.ClearList(binlogList)

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

				if err = DoIfNotExist(v.FileId, func() error {
					binlogList.PushBack(binlog.CreateLocalBinlog(v.FileId,
						v.FileLength, v.SourceInstance))
					return nil
				}); err != nil {
					failed++
					lastErr = err
					continue
				}
			}

			if binlogList.Len() > 0 {
				tmp := make([]*common.BingLog, binlogList.Len())
				i := 0
				gox.WalkList(binlogList, func(item interface{}) bool {
					tmp[i] = item.(*common.BingLog)
					i++
					return false
				})

				// write binlog first and dataset after.
				if err := writableBinlogManager.Write(tmp...); err != nil {
					lastErr = err
					logger.Debug("error write binlog: ", err)
				}

				gox.WalkList(binlogList, func(item interface{}) bool {
					v := item.(*common.BingLog)
					logger.Debug("add dataset...")
					if err := Add(string(v.FileId[:])); err != nil {
						failed++
						lastErr = err
						logger.Debug("error writing dataset")
						return false
					}
					logger.Debug("add dataset success")
					return false
				})
			}

			if failed == 0 {
				config.Offset = ret.Offset
				config.FileIndex = ret.FileIndex
				updateConfigChangeState(server.InstanceId, false, false)
				if len(ret.Logs) > 0 {
					logger.Debug("binlog write success")
				}
			} else {
				logger.Debug("binlog write error: ", lastErr, ", failed ", failed)
				break
			}
			if len(ret.Logs) == 0 {
				break
			}
		}
	})

}

// unWatch stops watching the storage member server.
func unWatch(server *common.Server) {
	syncLock.Lock()
	defer syncLock.Unlock()

	logger.Debug("unwatch server: ", server.ConnectionString(), "(", server.InstanceId, "): ")
	delete(watchingMembers, server.InstanceId)
}
