package libclient

import (
	"app"
	"container/list"
	"errors"
	"io"
	"libcommon/bridge"
	"libcommon/bridgev2"
	"libservice"
	"strconv"
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

func (tracker *TrackerInstance) GetTask() *bridge.Task {
	logger.Debug("get task for tracker", tracker.ConnStr)
	tracker.listIteLock.Lock()
	defer tracker.listIteLock.Unlock()
	if tracker.GetTaskSize() > 0 {
		ret := tracker.taskList.Remove(tracker.taskList.Front())
		if ret != nil {
			return ret.(*bridge.Task)
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
func (tracker *TrackerInstance) ExecTask(task *bridge.Task) (bool, error) {
	logger.Debug("exec task:", task.TaskType)
	if task.TaskType == app.TASK_SYNC_MEMBER {
		return TaskSyncMemberHandler(tracker)
	} else if task.TaskType == app.TASK_REGISTER_FILE {
		return TaskRegisterFileHandler(tracker)
	} else if task.TaskType == app.TASK_PULL_NEW_FILE {
		return TaskPullFileHandler(tracker)
	} else if task.TaskType == app.TASK_DOWNLOAD_FILE {
		logger.Debug("trying download file from other storage server...")
		if increaseActiveDownload(0) >= ParallelDownload {
			logger.Debug("ParallelDownload reached")
			// AddTask(task, tracker)
			return false, nil
		}
		fi, e1 := libservice.GetFullFileByFid(task.FileId, 0)
		if e1 != nil {
			return false, e1
		}
		if fi == nil || len(fi.Parts) == 0 {
			return false, nil
		}
		addDownloadingFile(fi.Id, false)
		go downloadFile(fi)
		return false, nil
	} else if task.TaskType == app.TASK_SYNC_ALL_STORAGES {
		regClientMeta := &bridge.OperationGetStorageServerRequest{}
		e2 := connBridge.SendRequest(bridge.O_SYNC_STORAGE, regClientMeta, 0, nil)
		if e2 != nil {
			return true, e2
		}
		e5 := connBridge.ReceiveResponse(func(response *bridge.Meta, in io.Reader) error {
			if response.Err != nil {
				return response.Err
			}
			logger.Debug("sync all storage server response from tracker", tracker.ConnStr, ": ", string(response.MetaBody))
			var validateResp = &bridge.OperationGetStorageServerResponse{}
			e3 := json.Unmarshal(response.MetaBody, validateResp)
			if e3 != nil {
				return e3
			}
			if validateResp.Status != bridge.STATUS_OK {
				return errors.New("error register to tracker server " + tracker.ConnStr + ", server response status:" + strconv.Itoa(validateResp.Status))
			}
			// connect success
			addMember(validateResp.GroupMembers)
			return nil
		})
		if e5 != nil {
			return true, e5
		}
		return false, nil
	} else if task.TaskType == app.TASK_SYNC_STATISTIC {
		regClientMeta := &bridge.OperationSyncStatisticRequest{}
		e2 := connBridge.SendRequest(bridge.O_SYNC_STATISTIC, regClientMeta, 0, nil)
		if e2 != nil {
			return true, e2
		}
		e5 := connBridge.ReceiveResponse(func(response *bridge.Meta, in io.Reader) error {
			if response.Err != nil {
				return response.Err
			}
			logger.Debug("sync statistic response:", string(response.MetaBody))
			var validateResp = &bridge.OperationSyncStatisticResponse{}
			e3 := json.Unmarshal(response.MetaBody, validateResp)
			if e3 != nil {
				return e3
			}
			if validateResp.Status != bridge.STATUS_OK {
				return errors.New("error sync statistic from tracker server, server response status:" + strconv.Itoa(validateResp.Status))
			}
			updateStatistic(tracker.ConnStr, validateResp.FileCount, validateResp.Statistic)
			return nil
		})
		if e5 != nil {
			return true, e5
		}
		return false, nil
	}
	return false, nil
}

