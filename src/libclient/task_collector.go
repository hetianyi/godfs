package libclient

import (
	"app"
	"libcommon/bridge"
	"libservice"
	"sync"
	"time"
	"util/common"
	"util/logger"
)

// storage 的任务分为：
// type 1: 定期从tracker同步members（定时任务，非持久化任务，插队任务，高优先级，任务列表中只能存在一条此类型的任务）
// type 2: 上报文件给tracker（定时任务，持久化任务，插队任务，高优先级）
// type 3: 定期向tracker服务器查询最新文件列表（定时任务，非持久化任务，插队任务，高优先级，任务列表中只能存在一条此类型的任务）
// type 4: 从其他group节点下载文件（定时任务，持久化任务，最低优先级，goroutine执行）

type TaskCollector struct {
	startLock  sync.Mutex                     // if the timer is already started
	Interval   time.Duration                  // time in Milliseconds, task exec interval.
	FirstDelay time.Duration                  // time in Milliseconds, task exec first time delay.
	ExecTimes  int                            // the collector execute times, ExecTimes<=0 means never stop
	Name       string                         // collector name
	Single     bool                           // 是否是能一个实例运行
	Job        func(tracker *TrackerInstance) // timer task
}


// copy a task collector in case of share lock
func copyTaskCollector(collector *TaskCollector) *TaskCollector {
	if collector == nil {
		return nil
	}
	return &TaskCollector{
		Interval:   collector.Interval,
		FirstDelay: collector.FirstDelay,
		ExecTimes:  collector.ExecTimes,
		Name:       collector.Name,
		Single:     collector.Single,
		Job:        collector.Job,
	}

}


// start task collectors of a tracker instance
func (collector *TaskCollector) Start(tracker *TrackerInstance) {
	if collector.Job == nil {
		logger.Error("no task assigned to this collector")
		return
	}
	collector.startLock.Lock()
	if collector.Interval <= 0 {
		collector.Interval = time.Millisecond * 10000
	}
	if collector.FirstDelay <= 0 {
		collector.FirstDelay = time.Millisecond * 0
	}
	timer := time.NewTicker(collector.Interval)
	execTimes := 0
	for {
		if collector.ExecTimes > 0 && execTimes >= collector.ExecTimes {
			logger.Debug("stop collector \""+collector.Name+"\"", "of tracker", tracker.ConnStr, "because of max execute times reached.")
			timer.Stop()
			break
		}
		time.Sleep(collector.FirstDelay)
		if collector.Name != "" {
			logger.Trace("exec task collector:", collector.Name, "of tracker", tracker.ConnStr)
		}
		common.Try(func() {
			collector.Job(tracker)
		}, func(i interface{}) {
			logger.Error("task collector \""+collector.Name+"\" return error:", i)
		})
		execTimes++
		<-timer.C
	}
}



// task collectors below
// ------------------------------------------------

// task collector: query files uploaded through this instance and push to all trackers
func QueryPushFileTaskCollector(tracker *TrackerInstance) {
	if tracker.client == nil {
		return
	}
	task, e1 := libservice.GetLocalPushFileTask(app.TASK_REGISTER_FILE, tracker.trackerUUID)
	if e1 != nil {
		logger.Error(e1)
		return
	}
	if task != nil {
		AddTask(task, tracker)
	}
}

// task collector: query files need to sync from other members
func QueryDownloadFileTaskCollector(tracker *TrackerInstance) {
	members := collectMemberInstanceId()
	// no member, no server for download.
	if members == "" {
		logger.Debug("no storage server available, skip collect download task")
		return
	}
	taskList, e1 := libservice.GetDownloadFileTask(app.TASK_DOWNLOAD_FILE)
	if e1 != nil {
		logger.Error(e1)
		return
	}
	if taskList == nil || taskList.Len() == 0 {
		logger.Debug("no file need to sync.")
		return
	}
	for ele := taskList.Front(); ele != nil; ele = ele.Next() {
		if !existsDownloadingFile(ele.Value.(*bridge.Task).FileId) {
			AddTask(ele.Value.(*bridge.Task), tracker)
		}
	}
}

// task collector: sync member info from trackers
func SyncMemberTaskCollector(tracker *TrackerInstance) {
	task := &bridge.Task{TaskType: app.TASK_SYNC_MEMBER}
	AddTask(task, tracker)
}

// task collector: query new files of other members from tracker
func QueryNewFileTaskCollector(tracker *TrackerInstance) {
	task := &bridge.Task{TaskType: app.TASK_PULL_NEW_FILE}
	AddTask(task, tracker)
}

// task collector: get all storage info tracker(used by native client)
func SyncAllStorageServersTaskCollector(tracker *TrackerInstance) {
	task := &bridge.Task{TaskType: app.TASK_SYNC_ALL_STORAGES}
	AddTask(task, tracker)
}

// task collector: sync statistic info of storage from tracker
func SyncStatisticTaskCollector(tracker *TrackerInstance) {
	task := &bridge.Task{TaskType: app.TASK_SYNC_STATISTIC}
	AddTask(task, tracker)
}

// task collectors end
// ------------------------------------------------