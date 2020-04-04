package api

import (
	"container/list"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/logger"
	"github.com/hetianyi/gox/timer"
	"sync"
	"time"
)

var (
	syncInstances map[string]*instanceStore
	syncLock      *sync.Mutex
	countLock     *sync.Mutex
	synced        = 0
)

func init() {
	syncLock = new(sync.Mutex)
	countLock = new(sync.Mutex)
	syncInstances = make(map[string]*instanceStore)
}

type instanceStore struct {
	fetchTime time.Time
	instance  *common.Instance
}

func (ins *instanceStore) expired() bool {
	return gox.GetTimestamp(time.Now().Add(common.SYNCHRONIZE_INTERVAL*2+time.Second*5)) < gox.GetTimestamp(ins.fetchTime)
}

func tracks(clientAPI ClientAPI, server *common.Server, synchronizeOnce bool, c chan int) {
	if !synchronizeOnce {
		go expireDetection()
	}
	timer.Start(0, common.SYNCHRONIZE_INTERVAL, 0, func(t *timer.Timer) {
		ret, err := clientAPI.SyncInstances(server)
		if err != nil {
			logger.Error("error synchronize with tracker server: ", server.ConnectionString(), ": ", err)
		} else {
			syncLock.Lock()
			defer syncLock.Unlock()

			if ret != nil && len(ret) > 0 {
				now := time.Now()
				for k, v := range ret {
					syncInstances[k] = &instanceStore{
						instance:  v,
						fetchTime: now,
					}
					if common.BootAs == common.BOOT_STORAGE {
						util.StoreSecrets(v.InstanceId, util.CollectMapKeys(v.Server.HistorySecrets)...)
					}
				}
			}
		}
		// used by client cli.
		if synchronizeOnce {
			t.Destroy()
			if c != nil {
				if err != nil {
					c <- 1
				} else {
					c <- 0
				}
			}
		}
	})
}

func FilterInstances(role common.Role) *list.List {
	syncLock.Lock()
	defer syncLock.Unlock()

	ret := list.New()
	for _, v := range syncInstances {
		if role == common.ROLE_ANY || v.instance.Role == role {
			ret.PushBack(v.instance)
		}
	}
	return ret
}

func FilterInstanceByInstanceId(instanceId string) *common.Instance {
	syncLock.Lock()
	defer syncLock.Unlock()

	for _, v := range syncInstances {
		if v.instance.InstanceId == instanceId {
			return v.instance
		}
	}
	return nil
}

// FilterReadonlyInstances gets readonly storage instance.
func FilterUploadableInstances() *list.List {
	syncLock.Lock()
	defer syncLock.Unlock()

	ret := list.New()
	for _, v := range syncInstances {
		if v.instance.Role == common.ROLE_STORAGE && v.instance.Attributes["readonly"] != "true" {
			ret.PushBack(v.instance)
		}
	}
	return ret
}

func expireDetection() {
	// allow 2 round failure synchronization
	timer.Start(0, 0, common.SYNCHRONIZE_INTERVAL, func(t *timer.Timer) {
		syncLock.Lock()
		defer syncLock.Unlock()

		for k, v := range syncInstances {
			if v.expired() {
				logger.Debug("instance expired: ", v.instance.ConnectionString())
				delete(syncInstances, k)
			}
		}
	})
}
