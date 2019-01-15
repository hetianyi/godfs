package libclient

import (
	"app"
	"container/list"
	"libcommon/bridge"
	"libcommon/bridgev2"
	"sync"
	"util/logger"
)

type TrackerInstance struct {
	taskList    list.List
	listIteLock *sync.Mutex
	client  *bridgev2.TcpBridgeClient
	Collectors  list.List
	Ready       bool
	nextRun     bool
	ConnStr     string
	trackerUUID string
}

// init tracker instance and start it's task collectors
func (tracker *TrackerInstance) Init() {
	logger.Debug("init tracker instance:", tracker.ConnStr)
	tracker.listIteLock = new(sync.Mutex)
	tracker.startTaskCollector()
	tracker.nextRun = true
}

func (tracker *TrackerInstance) SetConnBridgeClient(client *bridgev2.TcpBridgeClient) {
	tracker.client = client
}

// get task size in waiting list
func (tracker *TrackerInstance) GetTaskSize() int {
	return tracker.taskList.Len()
}

// start eac task collectors of a tracker instance
func (tracker *TrackerInstance) startTaskCollector() {
	for ele := tracker.Collectors.Front(); ele != nil; ele = ele.Next() {
		go ele.Value.(*TaskCollector).Start(tracker)
	}
}

func (tracker *TrackerInstance) GetTask() *bridgev2.Task {
	logger.Debug("get task for tracker", tracker.ConnStr)
	tracker.listIteLock.Lock()
	defer tracker.listIteLock.Unlock()
	if tracker.GetTaskSize() > 0 {
		ret := tracker.taskList.Remove(tracker.taskList.Front())
		if ret != nil {
			return ret.(*bridgev2.Task)
		}
	}
	return nil
}


// check task count of this type
func (tracker *TrackerInstance) checkTaskTypeCount(taskType int) int {
	tracker.listIteLock.Lock()
	defer tracker.listIteLock.Unlock()
	count := 0
	for e := tracker.taskList.Front(); e != nil; e = e.Next() {
		if e.Value != nil && e.Value.(*bridge.Task).TaskType == taskType {
			count++
		}
	}
	return count
}


// exec task
// return bool if the connection is forced close and need reconnect
func (tracker *TrackerInstance) ExecTask(task *bridgev2.Task) (bool, error) {
	logger.Debug("exec task:", task.TaskType)
	if task.TaskType == app.TASK_SYNC_MEMBER {
		return TaskSyncMemberHandler(tracker)
	} else if task.TaskType == app.TASK_REGISTER_FILE {
		return TaskRegisterFileHandler(tracker)
	} else if task.TaskType == app.TASK_PULL_NEW_FILE {
		return TaskPullFileHandler(tracker)
	} else if task.TaskType == app.TASK_DOWNLOAD_FILE {
		TaskDownloadFileHandler(task)
	} else if task.TaskType == app.TASK_SYNC_ALL_STORAGES {
		return TaskSyncAllStorageServerHandler(tracker)
	} else if task.TaskType == app.TASK_SYNC_STATISTIC {
		return TaskSyncStatisticInfo(tracker)
	}
	return false, nil
}

