package libclient

import (
    "libcommon/bridge"
    "util/logger"
    "encoding/json"
    "sync"
    "util/common"
    "libservice"
    "strconv"
    "time"
    "app"
)

var (
    managedStatistic = make(map[string][]bridge.ServerStatistic)
    managedTrackerInstance = make(map[string]*TrackerInstance)
    managedLock = new(sync.Mutex)
)

func updateStatistic(trackerUUID string, statistic []bridge.ServerStatistic) {
    ret, _ := json.Marshal(statistic)
    logger.Info("update statistic info:( ", string(ret), ")")
    managedStatistic[trackerUUID] = statistic
}


// register tracker instance when start track with tracker
func registerTrackerInstance(instance *TrackerInstance) {
    if instance == nil {
        return
    }
    managedLock.Lock()
    defer managedLock.Unlock()
    if managedTrackerInstance[instance.ConnStr] != nil {
        logger.Info("tracker instance already started, ignore registration")
    } else {
        managedTrackerInstance[instance.ConnStr] = instance
    }
}

// when delete a tracker, agent must remove the tracker instance and disconnect from tracker
func unRegisterTrackerInstance(connStr string) {
    managedLock.Lock()
    defer managedLock.Unlock()
    if managedTrackerInstance[connStr] == nil {
        logger.Info("tracker instance not exists")
    } else {
        delete(managedTrackerInstance, connStr)
    }
}

// update nextRun flag of tracker instance
// secret is optional
func UpdateTrackerInstanceState(connStr string, secret string, nextRun bool, trackerMaintainer *TrackerMaintainer) {
    managedLock.Lock()
    defer managedLock.Unlock()
    if managedTrackerInstance[connStr] == nil {
        logger.Info("tracker instance not exists")
        if nextRun {
            temp := make(map[string]string)
            temp[connStr] = secret
            trackerMaintainer.Maintain(temp)
        }
    } else {
        logger.Info("unload tracker instance:", connStr)
        ins := managedTrackerInstance[connStr]
        ins.nextRun = nextRun
    }
}


func SyncTrackerAliveStatus(trackerMaintainer *TrackerMaintainer) {
    timer := time.NewTicker(app.SYNC_STATISTIC_INTERVAL + 3)
    execTimes := 0
    for {
        common.Try(func() {
            trackers, e := libservice.GetAllWebTrackers()
            if e != nil {
                logger.Error(e)
            } else {
                if trackers != nil && trackers.Len() > 0 {
                    for ele := trackers.Front(); ele != nil; ele = ele.Next() {
                        tracker := ele.Value.(*bridge.WebTracker)
                        if tracker.Status == 0 {
                            UpdateTrackerInstanceState(tracker.Host + ":" + strconv.Itoa(tracker.Port),
                                tracker.Secret, tracker.Status == 1, trackerMaintainer)
                        }
                    }
                }
            }
        }, func(i interface{}) {
            logger.Error("error fetch web tracker status:", i)
        })
        execTimes++
        <-timer.C
    }
}


