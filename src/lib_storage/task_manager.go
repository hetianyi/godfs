package lib_storage

import (
    "util/logger"
    "container/list"
    "lib_common/bridge"
    "sync"
    "time"
    "lib_service"
    "app"
    "io"
    "encoding/json"
    "errors"
    "strconv"
)

// storage 的任务分为：
// type 1: 定期从tracker同步members（定时任务，非持久化任务，插队任务，高优先级，任务列表中只能存在一条此类型的任务）
// type 2: 上报文件给tracker（定时任务，持久化任务，插队任务，高优先级）
// type 3: 定期向tracker服务器查询最新文件列表（定时任务，非持久化任务，插队任务，高优先级，任务列表中只能存在一条此类型的任务）
// type 4: 从其他group节点下载文件（定时任务，持久化任务，最低优先级，goroutine执行）
var taskList list.List
var listIteLock *sync.Mutex
const MaxTaskSize = 1000
var GroupMembers []bridge.Member

func init() {
    listIteLock = new(sync.Mutex)
}

// get task size in waiting list
func GetTaskSize() int {
    return taskList.Len()
}


func GetTask() *bridge.Task {
    listIteLock.Lock()
    defer listIteLock.Unlock()
    if GetTaskSize() > 0 {
        return taskList.Remove(taskList.Front()).(*bridge.Task)
    }
    return nil
}

func AddTask(task *bridge.Task) {
    if task == nil {
        logger.Debug("can't push nil task")
        return
    }
    if task.TaskType == app.TASK_SYNC_MEMBER {
        if checkTaskTypeCount(task.TaskType) == 0 {
            logger.Debug("push task type 1")
            taskList.PushFront(task)
        } else {
            logger.Debug("can't push task type 1: task type exists")
        }
    } else if task.TaskType == app.TASK_REPORT_FILE {
        listIteLock.Lock()
        defer listIteLock.Unlock()
        for e := taskList.Front(); e != nil; e = e.Next() {
            if e.Value.(*bridge.Task).FileId == task.FileId {
                return
            }
        }
        for e := taskList.Front(); e != nil; e = e.Next() {
            if e.Value.(*bridge.Task).TaskType == 1 {
                logger.Debug("push task type 2")
                taskList.InsertAfter(task, e)
                break
            }
        }
    } else if task.TaskType == app.TASK_PULL_NEW_FILE {
        if checkTaskTypeCount(task.TaskType) == 0 {
            logger.Debug("push task type 3")
            taskList.PushBack(task)
        } else {
            logger.Debug("can't push task type 3: task type exists")
        }
    } else if task.TaskType == app.TASK_DOWNLOAD_FILE {
        listIteLock.Lock()
        defer listIteLock.Unlock()
        for e := taskList.Front(); e != nil; e = e.Next() {
            if e.Value.(*bridge.Task).FileId == task.FileId {
                return
            }
        }
        if taskList.Len() < MaxTaskSize {// temporary set to 1000
            logger.Debug("push task type 4")
            taskList.PushBack(task)
        } else {
            logger.Debug("can't push task type 4: task list full")
        }
    }
}

// 持久化任务失败，将任务放回到任务队列尾部
func FailReturnTask(task *bridge.Task) {
    listIteLock.Lock()
    defer listIteLock.Unlock()
    if task != nil {
        return
    }
    if task.TaskType == app.TASK_PULL_NEW_FILE || task.TaskType == app.TASK_DOWNLOAD_FILE {
        logger.Debug("push back task:", task.TaskType)
        taskList.PushBack(task)
    }
}



// check task count of this type
func checkTaskTypeCount(taskType int) int {
    listIteLock.Lock()
    defer listIteLock.Unlock()
    count := 0
    for e := taskList.Front(); e != nil; e = e.Next() {
        if e.Value.(*bridge.Task).TaskType == taskType {
            count++
        }
    }
    return count
}

// 启动任务收集器
func startTaskCollector() {
    go queryPersistTaskCollector()
    go syncMemberTaskCollector()
    //go queryNewFileTaskCollector()
}

// 查询本地持久化任务收集器
func queryPersistTaskCollector() {
    timer := time.NewTicker(time.Second * 100)
    for {
        <-timer.C
        logger.Debug("fetch tasks from db")
        taskList, e1 := lib_service.GetSyncTask()
        if e1 != nil {
            logger.Error(e1)
            continue
        }
        for e := taskList.Front(); e != nil; e = e.Next() {
            AddTask(e.Value.(*bridge.Task))
        }
    }
}

func syncMemberTaskCollector() {
    timer := time.NewTicker(time.Second * 100)
    for {
        logger.Debug("add sync member task")
        task := &bridge.Task{TaskType: app.TASK_SYNC_MEMBER}
        AddTask(task)
        <-timer.C
    }
}

func queryNewFileTaskCollector() {
    timer := time.NewTicker(time.Second * 30)
    for {
        <-timer.C
        logger.Debug("fetch tasks")
        task := &bridge.Task{TaskType: app.TASK_PULL_NEW_FILE}
        AddTask(task)
    }
}

// exec task
// return bool as if the connection is forced close and need reconnect
func ExecTask(task *bridge.Task, connBridge *bridge.Bridge) (bool, error) {
    logger.Debug("exec task:", task.TaskType)
    if task.TaskType == app.TASK_SYNC_MEMBER {
        // register storage client to tracker server
        regClientMeta := &bridge.OperationRegisterStorageClientRequest {
            BindAddr: app.BIND_ADDRESS,
            Group: app.GROUP,
            InstanceId: app.INSTANCE_ID,
            Port: app.PORT,
        }
        // reg client
        e2 := connBridge.SendRequest(bridge.O_REG_STORAGE, regClientMeta, 0, nil)
        if e2 != nil {
            return true, e2
        }
        e5 := connBridge.ReceiveResponse(func(response *bridge.Meta, in io.Reader) error {
            if response.Err != nil {
                return response.Err
            }
            logger.Debug(string(response.MetaBody))
            var validateResp = &bridge.OperationRegisterStorageClientResponse{}
            e3 := json.Unmarshal(response.MetaBody, validateResp)
            if e3 != nil {
                return e3
            }
            if validateResp.Status != bridge.STATUS_OK {
                return errors.New("error register to tracker server, server response status:" + strconv.Itoa(validateResp.Status))
            }
            // connect success
            GroupMembers = validateResp.GroupMembers
            return nil
        })
        if e5 != nil {
            return true, e5
        }
        return false, nil
    } else if task.TaskType == app.TASK_REPORT_FILE {
        fi, e1 := lib_service.GetFullFileByFid(task.FileId)
        if e1 != nil {
            return false, e1
        }
        // register storage client to tracker server
        regFileMeta := &bridge.OperationRegisterFileRequest {
            File: fi,
        }
        // reg client
        e2 := connBridge.SendRequest(bridge.O_REG_FILE, regFileMeta, 0, nil)
        if e2 != nil {
            return true, e2
        }
        e5 := connBridge.ReceiveResponse(func(response *bridge.Meta, in io.Reader) error {
            if response.Err != nil {
                return response.Err
            }
            var regResp = &bridge.OperationRegisterFileResponse{}
            e3 := json.Unmarshal(response.MetaBody, regResp)
            if e3 != nil {
                return e3
            }
            if regResp.Status != bridge.STATUS_OK {
                return errors.New("error register file "+ strconv.Itoa(task.FileId) +" to tracker server, server response status:" + strconv.Itoa(regResp.Status))
            }
            e7 := lib_service.FinishSyncTask(task.FileId)
            if e7 != nil {
                return e7
            }
            return nil
        })
        if e5 != nil {
            return true, e5
        }
        return false, nil
    }
    return false, nil
}




/*
func start() {
    timer := time.NewTicker(time.Second * app.SYNC_INTERVAL)
    for {
        <-timer.C
        logger.Debug("fetch sync tasks")

    }
}*/
