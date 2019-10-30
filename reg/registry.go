// package reg
//
//
//
//
package reg

import (
	"errors"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/logger"
	"github.com/hetianyi/gox/timer"
	"sync"
	"time"
)

var (
	// instanceSet stores all instances synchronized from tracker server.
	instanceSet    = make(map[string]*common.Instance)
	lock           = new(sync.Mutex)
	ExpirationTime = time.Second * 30 // 30s
)

// InitRegistry() starts a timer job for instance expiration detection
// in a single goroutine.
func InitRegistry() {
	go expirationDetection()
}

// Put registers a new Instance.
//
// It returns an error if the new Instance is conflict with the registered Instance,
func Put(ins *common.Instance) error {
	lock.Lock()
	defer lock.Unlock()
	if ins == nil {
		return errors.New("instance cannot be null")
	}
	if i := isInstanceConflict(ins); i != nil {
		return errors.New("instance conflict with server " + i.Server.ConnectionString())
	}
	logger.Debug("registered new instance: ", ins.InstanceId, "@", ins.Server.ConnectionString())
	ins.State = common.REGISTER_HOLD
	ins.RegisterTime = time.Now().UnixNano()
	instanceSet[ins.InstanceId] = ins
	if ins.Role == common.ROLE_STORAGE {
		util.StoreSecrets(ins.InstanceId, util.CollectMapKeys(ins.Server.HistorySecrets)...)
	}
	return nil
}

// Free sets registered instance state to FREE so it can be detected by timer expiration job.
func Free(instanceId string) {
	lock.Lock()
	defer lock.Unlock()
	ins := instanceSet[instanceId]
	if ins != nil {
		logger.Debug("free instance: ", ins.InstanceId, "@", ins.Server.ConnectionString())
		ins.RegisterTime = time.Now().UnixNano()
		ins.State = common.REGISTER_FREE
	}
}

// Remove indicates this client deregister from Registry immediately.
func Remove(ins *common.Instance) {
	lock.Lock()
	defer lock.Unlock()
	logger.Debug("deregister instance: ", ins.InstanceId, "@", ins.Server.ConnectionString())
	delete(instanceSet, ins.InstanceId)
}

// InstanceSetSnapshot takes a snapshot for current instances.
func InstanceSetSnapshot() map[string]*common.Instance {
	lock.Lock()
	defer lock.Unlock()
	snapshot := make(map[string]*common.Instance)
	for k, i := range instanceSet {
		snapshot[k] = i
	}
	return snapshot
}

// isInstanceConflict for judgement whether the new instance
// is conflict with other registered instance.
func isInstanceConflict(ins *common.Instance) *common.Instance {
	i := instanceSet[ins.InstanceId]
	if i != nil && ins.Server.InstanceId == i.Server.InstanceId &&
		i.Server.ConnectionString() != ins.Server.ConnectionString() {
		return i
	}
	return nil
}

// expirationDetection is a timer job for removing expired instance.
func expirationDetection() {
	timerTask := func() {
		lock.Lock()
		defer lock.Unlock()

		logger.Debug("current instances: ", len(instanceSet)) // TODO remove

		deadLine := time.Now().UnixNano() - int64(ExpirationTime)
		for _, i := range instanceSet {
			if i.State == common.REGISTER_FREE && i.RegisterTime <= deadLine {
				logger.Debug("instance expired: ", i.InstanceId, "@", i.Server.ConnectionString())
				delete(instanceSet, i.InstanceId)
			}
		}
	}

	timer.Start(0, ExpirationTime, 0, func(t *timer.Timer) {
		gox.Try(func() {
			timerTask()
		}, func(e interface{}) {
			logger.Error("expire err: ", e)
		})
	})
}
