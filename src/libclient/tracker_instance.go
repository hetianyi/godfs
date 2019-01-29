package libclient

import (
	"app"
	"container/list"
	"libcommon/bridgev2"
	"sync"
	"util/logger"
)

type TrackerInstance struct {
	taskList    list.List
	listIteLock *sync.Mutex
	client      *bridgev2.TcpBridgeClient
	Collectors  list.List
	Ready       bool
	nextRun     bool
	ConnStr     string
	trackerUUID string
}

// Init init tracker instance and start it's task collectors
func (tracker *TrackerInstance) Init() {
	logger.Debug("init tracker instance:", tracker.ConnStr)
	tracker.listIteLock = new(sync.Mutex)
	tracker.startTaskCollector()
	tracker.nextRun = true
}

func (tracker *TrackerInstance) SetConnBridgeClient(client *bridgev2.TcpBridgeClient) {
	tracker.client = client
}

// GetTaskSize get task size in waiting list
func (tracker *TrackerInstance) GetTaskSize() int {
	return tracker.taskList.Len()
}

// startTaskCollector start eac task collectors of a tracker instance
func (tracker *TrackerInstance) startTaskCollector() {
	for ele := tracker.Collectors.Front(); ele != nil; ele = ele.Next() {
		go ele.Value.(*TaskCollector).Start(tracker)
	}
}

func (tracker *TrackerInstance) GetTask() *bridgev2.Task {
	logger.Trace("get task for tracker", tracker.ConnStr)
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

// checkTaskTypeCount check task count of this type
func (tracker *TrackerInstance) checkTaskTypeCount(taskType int) int {
	tracker.listIteLock.Lock()
	defer tracker.listIteLock.Unlock()
	count := 0
	for e := tracker.taskList.Front(); e != nil; e = e.Next() {
		if e.Value != nil && e.Value.(*bridgev2.Task).TaskType == taskType {
			count++
		}
	}
	return count
}

// ExecTask exec task
// return bool if the connection is forced close and need reconnect
func (tracker *TrackerInstance) ExecTask(task *bridgev2.Task) (bool, error) {
	logger.Debug("exec task:", task.TaskType)
	if task.TaskType == app.TaskSyncMembers {
		return TaskSyncMemberHandler(tracker)
	} else if task.TaskType == app.TaskRegisterFiles {
		return TaskRegisterFileHandler(tracker)
	} else if task.TaskType == app.TaskPullNewFiles {
		return TaskPullFileHandler(tracker)
	} else if task.TaskType == app.TaskDownloadFiles {
		TaskDownloadFileHandler(task)
	} else if task.TaskType == app.TaskSyncAllStorage {
		return TaskSyncAllStorageServerHandler(tracker)
	} else if task.TaskType == app.TaskSyncStatistic {
		return TaskSyncStatisticInfo(tracker)
	}
	return false, nil
}
