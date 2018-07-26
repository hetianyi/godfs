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
    "math/rand"
    "lib_common"
    "util/file"
    "lib_client"
)

// storage 的任务分为：
// type 1: 定期从tracker同步members（定时任务，非持久化任务，插队任务，高优先级，任务列表中只能存在一条此类型的任务）
// type 2: 上报文件给tracker（定时任务，持久化任务，插队任务，高优先级）
// type 3: 定期向tracker服务器查询最新文件列表（定时任务，非持久化任务，插队任务，高优先级，任务列表中只能存在一条此类型的任务）
// type 4: 从其他group节点下载文件（定时任务，持久化任务，最低优先级，goroutine执行）
const ParallelDownload = 10
var GroupMembers list.List
var memberIteLock *sync.Mutex

var clientPool *lib_client.ClientConnectionPool

func init() {
    memberIteLock = new(sync.Mutex)
    clientPool = &lib_client.ClientConnectionPool{}
    clientPool.Init(10)
}

type ITracker interface {
    Init()
    GetTaskSize() int
    GetTask() *bridge.Task
    AddTask(task *bridge.Task)
    //FailReturnTask(task *bridge.Task)
    CheckTaskTypeCount(taskType int)
    StartTaskCollector()
    QueryPersistTaskCollector()
    SyncMemberTaskCollector()
    QueryNewFileTaskCollector()
    ExecTask(task *bridge.Task, connBridge *bridge.Bridge) (bool, error)
}

type TrackerInstance struct {
    taskList list.List
    listIteLock *sync.Mutex
    connBridge *bridge.Bridge
}



func (tracker *TrackerInstance) Init(connBridge *bridge.Bridge) {
    tracker.listIteLock = new(sync.Mutex)
    tracker.connBridge = connBridge
}

// get task size in waiting list
func (tracker *TrackerInstance) GetTaskSize() int {
    return tracker.taskList.Len()
}

func addMember(members []bridge.Member) {
    memberIteLock.Lock()
    defer memberIteLock.Unlock()
    if members != nil {
        return
    }
    for i := range members {
        a := members[i]
        for e := GroupMembers.Front(); e != nil; e = e.Next() {
            m := e.Value.(*bridge.Member)
            if a.InstanceId == m.InstanceId {
                break
            }
        }
        GroupMembers.PushBack(a)
    }
}


func (tracker *TrackerInstance) GetTask() *bridge.Task {
    tracker.listIteLock.Lock()
    defer tracker.listIteLock.Unlock()
    if tracker.GetTaskSize() > 0 {
        return tracker.taskList.Remove(tracker.taskList.Front()).(*bridge.Task)
    }
    return nil
}

func (tracker *TrackerInstance) AddTask(task *bridge.Task) {
    if task == nil {
        logger.Debug("can't push nil task")
        return
    }
    if task.TaskType == app.TASK_SYNC_MEMBER {
        if tracker.CheckTaskTypeCount(task.TaskType) == 0 {
            logger.Debug("push task type 1")
            tracker.taskList.PushFront(task)
        } else {
            logger.Debug("can't push task type 1: task type exists")
        }
    } else if task.TaskType == app.TASK_REPORT_FILE {
        tracker.listIteLock.Lock()
        defer tracker.listIteLock.Unlock()
        for e := tracker.taskList.Front(); e != nil; e = e.Next() {
            if e.Value.(*bridge.Task).FileId == task.FileId {
                return
            }
        }
        logger.Debug("push task type 2")
        if tracker.taskList.Front() != nil && tracker.taskList.Front().Value.(*bridge.Task).TaskType == app.TASK_SYNC_MEMBER {
            tracker.taskList.InsertAfter(task, tracker.taskList.Front())
        } else {
            tracker.taskList.PushFront(task)
        }
    } else if task.TaskType == app.TASK_PULL_NEW_FILE {
        if tracker.CheckTaskTypeCount(task.TaskType) == 0 {
            logger.Debug("push task type 3")
            tracker.taskList.PushBack(task)
        } else {
            logger.Debug("can't push task type 3: task type exists")
        }
    } else if task.TaskType == app.TASK_DOWNLOAD_FILE {
        tracker.listIteLock.Lock()
        defer tracker.listIteLock.Unlock()
        total := 0
        for e := tracker.taskList.Front(); e != nil; e = e.Next() {
            if e.Value.(*bridge.Task).FileId == task.FileId {
                total++
            }
        }
        if total <= ParallelDownload {// 限制最大并行下载任务数
            logger.Debug("push task type 4")
            tracker.taskList.PushBack(task)
        } else {
            logger.Debug("can't push task type 4: task list full")
        }
    }
}




// check task count of this type
func (tracker *TrackerInstance) CheckTaskTypeCount(taskType int) int {
    tracker.listIteLock.Lock()
    defer tracker.listIteLock.Unlock()
    count := 0
    for e := tracker.taskList.Front(); e != nil; e = e.Next() {
        if e.Value.(*bridge.Task).TaskType == taskType {
            count++
        }
    }
    return count
}

// 启动任务收集器
func (tracker *TrackerInstance) StartTaskCollector() {
    go tracker.QueryPersistTaskCollector()
    go tracker.SyncMemberTaskCollector()
    go tracker.QueryNewFileTaskCollector()
}

// 查询本地持久化任务收集器
func (tracker *TrackerInstance) QueryPersistTaskCollector() {
    timer := time.NewTicker(time.Second * 10)
    for {
        <-timer.C
        logger.Debug("add task: fetch tasks from db")
        taskList, e1 := lib_service.GetSyncTask()
        if e1 != nil {
            logger.Error(e1)
            continue
        }
        for e := taskList.Front(); e != nil; e = e.Next() {
            tracker.AddTask(e.Value.(*bridge.Task))
        }
    }
}

func (tracker *TrackerInstance) SyncMemberTaskCollector() {
    timer := time.NewTicker(time.Second * 20)
    for {
        logger.Debug("add task: sync storage member")
        task := &bridge.Task{TaskType: app.TASK_SYNC_MEMBER}
        tracker.AddTask(task)
        <-timer.C
    }
}

func (tracker *TrackerInstance) QueryNewFileTaskCollector() {
    timer := time.NewTicker(time.Second * 5)
    for {
        <-timer.C
        logger.Debug("add task: pull files from tracker")
        task := &bridge.Task{TaskType: app.TASK_PULL_NEW_FILE}
        tracker.AddTask(task)
    }
}

// exec task
// return bool if the connection is forced close and need reconnect
func (tracker *TrackerInstance) ExecTask(task *bridge.Task) (bool, error) {
    connBridge := *tracker.connBridge
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
            addMember(validateResp.GroupMembers)
            return nil
        })
        if e5 != nil {
            return true, e5
        }
        return false, nil
    } else if task.TaskType == app.TASK_REPORT_FILE {
        fi, e1 := lib_service.GetFullFileByFid(task.FileId, 2)
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
    } else if task.TaskType == app.TASK_PULL_NEW_FILE {
        logger.Debug("trying pull new file from tracker...")
        baseId, e1 := lib_service.GetSyncId()
        if e1 != nil {
            return false, e1
        }
        // register storage client to tracker server
        pullMeta := &bridge.OperationPullFileRequest {
            BaseId: baseId,
        }
        // reg client
        e2 := connBridge.SendRequest(bridge.O_PULL_NEW_FILES, pullMeta, 0, nil)
        if e2 != nil {
            return true, e2
        }
        e5 := connBridge.ReceiveResponse(func(response *bridge.Meta, in io.Reader) error {
            if response.Err != nil {
                return response.Err
            }
            var pullResp = &bridge.OperationPullFileResponse{}
            e3 := json.Unmarshal(response.MetaBody, pullResp)
            if e3 != nil {
                return e3
            }
            if pullResp.Status != bridge.STATUS_OK {
                return errors.New("error register file "+ strconv.Itoa(task.FileId) +" to tracker server, server response status:" + strconv.Itoa(pullResp.Status))
            }

            files := pullResp.Files
            logger.Debug("got", len(files), "new files")
            for i := range files {
                eee := lib_service.StorageAddRemoteFile(&files[i])
                if eee != nil {
                    return nil
                }
            }
            return nil
        })
        if e5 != nil {
            return true, e5
        }
        return false, nil
    } else if task.TaskType == app.TASK_DOWNLOAD_FILE {
        fi, e1 := lib_service.GetFullFileByFid(task.FileId, 0)
        if e1 != nil {
            return false, e1
        }
        if fi != nil || len(fi.Parts) == 0 {
            return false, nil
        }
        e2 := downloadFile(fi)
        if e2 != nil {
            return false, e2
        }
        e3 := lib_service.UpdateFileStatus(fi.Id)
        return false, e3
    }
    return false, nil
}


func downloadFile(fi *bridge.File) error {
    logger.Debug("downloading file from other storage server...")
    for i := range fi.Parts {
        part := fi.Parts[i]
        pFile, e1 := file.GetFile(lib_common.GetFilePathByMd5(part.Md5))
        if e1 != nil {
            return e1
        }
        // TODO 复用lib_client 作为下载客户端
        lib_cli

    }
}


func groupFiles(fi *bridge.File) map[string]list.List {
    var group = make(map[string]list.List)
    for i := range fi.Parts {
        key := fi.Instance
    }
}


// select a storage server matching given group and instanceId
// excludes contains fail storage and not gonna use this time.
func selectStorageServer(instanceId string) *bridge.Member {
    for i := range GroupMembers {
        b := GroupMembers[i]
        if instanceId == b.InstanceId {
            return &b
        }
    }
    return nil
}