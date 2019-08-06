package api

import (
	"container/list"
	"github.com/hetianyi/godfs/common"
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
	expireDetection()
}

type instanceStore struct {
	fetchTime time.Time
	instance  *common.Instance
}

func (ins *instanceStore) expired() bool {
	return gox.GetTimestamp(time.Now()) > gox.GetTimestamp(ins.fetchTime)
}

// SynchronizeFinishedTrackers is used by client mode only.
func SynchronizeFinishedTrackers(value int) int {
	countLock.Lock()
	defer countLock.Unlock()
	synced += value
	return synced
}

func tracks(clientAPI ClientAPI, server *common.Server, synchronizeOnce bool) {
	timer.Start(0, 0, common.SYNCHRONIZE_INTERVAL, func(t *timer.Timer) {
		ret, err := clientAPI.SyncInstances(server)
		if err != nil {
			logger.Error("error synchronize with tracker server: ", server.ConnectionString(), "(", server.InstanceId, ")", ": ", err)
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
				}
			}
		}
		if synchronizeOnce {
			SynchronizeFinishedTrackers(1)
			t.Destroy()
		}
	})
}

func FilterInstances(role common.Role) *list.List {
	syncLock.Lock()
	defer syncLock.Unlock()
	ret := list.New()
	for _, v := range syncInstances {
		if v.instance.Role == role {
			ret.PushBack(v.instance)
		}
	}
	return ret
}

func expireDetection() {
	// allow 2 round failure synchronization
	timer.Start(0, 0, common.SYNCHRONIZE_INTERVAL*2+time.Second*5, func(t *timer.Timer) {
		syncLock.Lock()
		defer syncLock.Unlock()
		for k, v := range syncInstances {
			if v.expired() {
				logger.Debug("instance expired: ", v.instance.ConnectionString(), "(", v.instance.InstanceId, ")")
				delete(syncInstances, k)
			}
		}
	})
}
