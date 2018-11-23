package libclient

import (
	"app"
	"encoding/json"
	"libcommon/bridge"
	"libservice"
	"strconv"
	"sync"
	"time"
	"util/common"
	"util/logger"
)

var (
	managedStatistic       = make(map[string][]bridge.ServerStatistic)
	managedTrackerInstance = make(map[string]*TrackerInstance)
	managedLock            = new(sync.Mutex)
)

// trackerUUID -> host:port
func updateStatistic(trackerUUID string, fileCount int, statistic []bridge.ServerStatistic) {
	ret, _ := json.Marshal(statistic)
	logger.Info("update statistic info:( ", string(ret), ")")
	managedStatistic[trackerUUID] = statistic

	if statistic != nil {
		arr := make([]*bridge.WebStorage, len(statistic))
		for i := 0; i < len(statistic); i++ {
			if statistic[i].UUID == "" {
				continue
			}
			item := &bridge.WebStorage{
				Host:       statistic[i].AdvertiseAddr,
				Port:       statistic[i].Port,
				TotalFiles: statistic[i].TotalFiles,
				UUID:       statistic[i].UUID,

				HttpEnable: statistic[i].HttpEnable,
				HttpPort:   statistic[i].HttpPort,
				Downloads:  statistic[i].Downloads,
				Uploads:    statistic[i].Uploads,
				DiskUsage:  statistic[i].DiskUsage,
				ReadOnly:   statistic[i].ReadOnly,
				StartTime:  statistic[i].StartTime,
				InstanceId: statistic[i].InstanceId,
				Group:      statistic[i].Group,
				Status:     app.STATUS_ENABLED,
				Memory:     statistic[i].Memory,
				IOin:       statistic[i].IOin,
				IOout:      statistic[i].IOout,

				LogTime:        statistic[i].LogTime,
				StageDownloads: statistic[i].StageDownloads,
				StageUploads:   statistic[i].StageUploads,
				StageIOin:      statistic[i].StageIOin,
				StageIOout:     statistic[i].StageIOout,
			}
			arr[i] = item
		}
		e := libservice.AddWebStorage(trackerUUID, fileCount, arr)
		if e != nil {
			logger.Error("error insert web storage items:", e)
		}
	}
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
			logger.Info("start new tracker instance:", connStr)
			temp := make(map[string]string)
			temp[connStr] = secret
			trackerMaintainer.Maintain(temp)
		}
	} else {
		//logger.Info("unload tracker instance:", connStr)
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
						UpdateTrackerInstanceState(tracker.Host+":"+strconv.Itoa(tracker.Port),
							tracker.Secret, tracker.Status == app.STATUS_ENABLED, trackerMaintainer)
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
