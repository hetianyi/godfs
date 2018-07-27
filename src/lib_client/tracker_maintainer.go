package lib_client

import (
    "lib_common"
    "util/logger"
    "net"
    "lib_common/bridge"
    "time"
    "strconv"
    "app"
    "lib_service"
    "container/list"
    "sync"
    "util/common"
    "io"
    "encoding/json"
    "errors"
    "os"
    "util/file"
)

// storage 的任务分为：
// type 1: 定期从tracker同步members（定时任务，非持久化任务，插队任务，高优先级，任务列表中只能存在一条此类型的任务）
// type 2: 上报文件给tracker（定时任务，持久化任务，插队任务，高优先级）
// type 3: 定期向tracker服务器查询最新文件列表（定时任务，非持久化任务，插队任务，高优先级，任务列表中只能存在一条此类型的任务）
// type 4: 从其他group节点下载文件（定时任务，持久化任务，最低优先级，goroutine执行）
const ParallelDownload = 10
const MaxWaitDownload = 100
var GroupMembers list.List
var memberIteLock *sync.Mutex
var downloadClient *Client
var activeDownload int
var activeDownloadLock *sync.Mutex

func init() {
    memberIteLock = new(sync.Mutex)
    activeDownloadLock = new(sync.Mutex)
}

type IMaintainer interface {
    Maintain(trackers string)
    track(tracker string)
}

type TrackerMaintainer struct {
    Collectors list.List
}

type ITracker interface {
    Init()
    GetTaskSize() int
    GetTask() *bridge.Task
    //FailReturnTask(task *bridge.Task)
    checkTaskTypeCount(taskType int)
    startTaskCollector()
    ExecTask(task *bridge.Task, connBridge *bridge.Bridge) (bool, error)
}

type TrackerInstance struct {
    taskList list.List
    listIteLock *sync.Mutex
    connBridge *bridge.Bridge
    Collectors list.List
}


type Collector interface {
    Start(tracker *TrackerInstance)
}

type TaskCollector struct {
    startLock sync.Mutex //if the timer is already started
    Interval time.Duration    // time in Milliseconds, task exec interval.
    FirstDelay time.Duration    // time in Milliseconds, task exec first time delay.
    ExecTimes int    // the collector execute times, ExecTimes<=0 means never stop
    Name string    // collector name
    Single bool    // 是否是能一个实例运行
    Job func(tracker *TrackerInstance) // timer task
}

// 类型为TASK_DOWNLOAD_FILE的任务只能在一个trackerInstance里面执行
func trackTaskFilter(allCollectors *list.List, index int) *list.List {
    if index == 0 {
        return allCollectors
    }
    var ret list.List
    if allCollectors == nil {
        return nil
    }
    for ele := allCollectors.Front(); ele != nil; ele = ele.Next() {
        if !ele.Value.(*TaskCollector).Single {
            ret.PushBack(ele.Value)
        }
    }
    return &ret
}

// communication with tracker
func (maintainer *TrackerMaintainer) Maintain(trackers string) {
    ls := lib_common.ParseTrackers(trackers)
    if ls.Len() == 0 {
        logger.Warn("no trackers set, the storage server will run in stand-alone mode.")
        return
    }
    index := 0
    for e := ls.Front(); e != nil; e = e.Next() {
        go maintainer.track(e.Value.(string), index)
        index++
    }
}

// connect to each tracker
func (maintainer *TrackerMaintainer) track(tracker string, index int) {
    logger.Info("start tracker conn with tracker server:", tracker)
    retry := 0
    for {//keep trying to connect to tracker server.
        conn, e := net.Dial("tcp", tracker)
        if e == nil {
            // validate client
            connBridge, e1 := connectAndValidate(conn)
            if e1 != nil {
                bridge.Close(conn)
                logger.Error(e1)
            } else {
                retry = 0
                logger.Debug("connect to tracker server success.")

                trackerInstance := TrackerInstance{Collectors: *trackTaskFilter(&maintainer.Collectors, index)}
                trackerInstance.Init(connBridge)

                for { // keep sending client statistic info to tracker server.
                    task := trackerInstance.GetTask()
                    if task == nil {
                        time.Sleep(time.Second * 1)
                        continue
                    }
                    forceClosed, e2 := trackerInstance.ExecTask(task)
                    if e2 != nil {
                        logger.Debug("task exec error:", e2)
                    } else {
                        logger.Trace("task exec success:", task.TaskType)
                    }
                    if task.Callback != nil {
                        task.Callback(task, e2)
                    }
                    if forceClosed {
                        logger.Debug("connection force closed by client")
                        bridge.Close(conn)
                        break
                    }
                }
            }
        } else {
            logger.Error("(" + strconv.Itoa(retry) + ")error connect to tracker server:", tracker)
        }
        retry++
        time.Sleep(time.Second * 1)
    }
}


// connect to tracker server and register client to it.
func connectAndValidate(conn net.Conn) (*bridge.Bridge, error) {
    // create bridge
    connBridge := bridge.NewBridge(conn)
    // send validate request
    e1 := connBridge.ValidateConnection("")
    if e1 != nil {
        return nil, e1
    }
    return connBridge, nil
}


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
        if collector.ExecTimes > 0 && execTimes >= collector.ExecTimes  {
            timer.Stop()
            break
        }
        time.Sleep(collector.FirstDelay)
        if collector.Name != "" {
            logger.Debug("exec task collector:", collector.Name)
        }
        common.Try(func() {
            collector.Job(tracker)
        }, func(i interface{}) {
            logger.Error("task collector return error:", i)
        })
        execTimes++
        <-timer.C
    }
}



func (tracker *TrackerInstance) Init(connBridge *bridge.Bridge) {
    tracker.listIteLock = new(sync.Mutex)
    tracker.connBridge = connBridge
    tracker.startTaskCollector()
}

// get task size in waiting list
func (tracker *TrackerInstance) GetTaskSize() int {
    return tracker.taskList.Len()
}

func addMember(members []bridge.Member) {
    memberIteLock.Lock()
    defer memberIteLock.Unlock()
    if members == nil {
        return
    }
    for i := range members {
        a := members[i]
        exists := false
        for e := GroupMembers.Front(); e != nil; e = e.Next() {
            m := e.Value.(*bridge.Member)
            if a.InstanceId == m.InstanceId {
                exists = true
                break
            }
        }
        if !exists {
            logger.Debug("add storage member server:", a)
            GroupMembers.PushBack(&a)
        }
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

// 某些任务不能多个tracker instance重复执行，只能选择其中一个予以执行
func AddTask(task *bridge.Task, tracker *TrackerInstance) bool {
    if task == nil {
        logger.Debug("can't push nil task")
        return false
    }
    if task.TaskType == app.TASK_SYNC_MEMBER || task.TaskType == app.TASK_SYNC_ALL_STORAGES {
        if tracker.checkTaskTypeCount(task.TaskType) == 0 {
            logger.Trace("push task type:", strconv.Itoa(task.TaskType))
            tracker.taskList.PushFront(task)
            return true
        } else {
            logger.Debug("can't push task type "+ strconv.Itoa(task.TaskType) +": task type exists")
            return false
        }
    } else if task.TaskType == app.TASK_REPORT_FILE {
        tracker.listIteLock.Lock()
        defer tracker.listIteLock.Unlock()
        for e := tracker.taskList.Front(); e != nil; e = e.Next() {
            if e.Value.(*bridge.Task).FileId == task.FileId {
                return false
            }
        }
        logger.Trace("push task type 2")
        if tracker.taskList.Front() != nil && tracker.taskList.Front().Value.(*bridge.Task).TaskType == app.TASK_SYNC_MEMBER {
            tracker.taskList.InsertAfter(task, tracker.taskList.Front())
        } else {
            tracker.taskList.PushFront(task)
        }
        return true
    } else if task.TaskType == app.TASK_PULL_NEW_FILE {
        if tracker.checkTaskTypeCount(task.TaskType) == 0 {
            logger.Trace("push task type 3")
            tracker.taskList.PushBack(task)
            return true
        } else {
            logger.Debug("can't push task type 3: task type exists")
            return false
        }
    } else if task.TaskType == app.TASK_DOWNLOAD_FILE {
        tracker.listIteLock.Lock()
        defer tracker.listIteLock.Unlock()
        total := 0
        for e := tracker.taskList.Front(); e != nil; e = e.Next() {
            if e.Value.(*bridge.Task).TaskType == task.TaskType {
                total++
            }
        }
        if total < MaxWaitDownload {// 限制最大并行下载任务数
            logger.Trace("push task type 4")
            tracker.taskList.PushBack(task)
            return true
        } else {
            logger.Debug("can't push task type 4: task list full")
            return false
        }
    }
    return false
}




// check task count of this type
func (tracker *TrackerInstance) checkTaskTypeCount(taskType int) int {
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
func (tracker *TrackerInstance) startTaskCollector() {
    for ele := tracker.Collectors.Front(); ele != nil; ele = ele.Next() {
        go ele.Value.(*TaskCollector).Start(tracker)
    }
    //go tracker.QueryPersistTaskCollector()
    //go tracker.SyncMemberTaskCollector()
    //go tracker.QueryNewFileTaskCollector()
}


// exec task
// return bool if the connection is forced close and need reconnect
func (tracker *TrackerInstance) ExecTask(task *bridge.Task) (bool, error) {
    connBridge := *tracker.connBridge
    logger.Trace("exec task:", task.TaskType)
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
        if increateActiveDownload(0) >= ParallelDownload {
            logger.Debug("ParallelDownload reached")
            AddTask(task, tracker)
            return false, nil
        }
        fi, e1 := lib_service.GetFullFileByFid(task.FileId, 0)
        if e1 != nil {
            return false, e1
        }
        if fi == nil || len(fi.Parts) == 0 {
            return false, nil
        }
        go downloadFile(fi)
        return false, nil
    } else if task.TaskType == app.TASK_SYNC_ALL_STORAGES {
        regClientMeta := &bridge.OperationGetStorageServerRequest {}
        e2 := connBridge.SendRequest(bridge.O_SYNC_STORAGE, regClientMeta, 0, nil)
        if e2 != nil {
            return true, e2
        }
        e5 := connBridge.ReceiveResponse(func(response *bridge.Meta, in io.Reader) error {
            if response.Err != nil {
                return response.Err
            }
            logger.Debug("sync storage server response:", string(response.MetaBody))
            var validateResp = &bridge.OperationGetStorageServerResponse{}
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
    }
    return false, nil
}







// task collectors below
// ------------------------------------------------


// 查询本地持久化任务收集器
func QueryPushFileTaskCollector(tracker *TrackerInstance) {
    taskList, e1 := lib_service.GetTask(app.TASK_REPORT_FILE)
    if e1 != nil {
        logger.Error(e1)
        return
    }
    for e := taskList.Front(); e != nil; e = e.Next() {
        AddTask(e.Value.(*bridge.Task), tracker)
    }
}
// 查询本地持久化任务收集器
func QueryDownloadFileTaskCollector(tracker *TrackerInstance) {
    taskList, e1 := lib_service.GetTask(app.TASK_DOWNLOAD_FILE)
    if e1 != nil {
        logger.Error(e1)
        return
    }
    for e := taskList.Front(); e != nil; e = e.Next() {
        AddTask(e.Value.(*bridge.Task), tracker)
    }
}

func SyncMemberTaskCollector(tracker *TrackerInstance) {
    task := &bridge.Task{TaskType: app.TASK_SYNC_MEMBER}
    AddTask(task, tracker)
}

func QueryNewFileTaskCollector(tracker *TrackerInstance) {
    task := &bridge.Task{TaskType: app.TASK_PULL_NEW_FILE}
    AddTask(task, tracker)
}


func SyncAllStorageServersTaskCollector(tracker *TrackerInstance) {
    task := &bridge.Task{TaskType: app.TASK_SYNC_ALL_STORAGES}
    AddTask(task, tracker)
}
// task collectors end
// ------------------------------------------------


var lockInitDownloadClient sync.Mutex
func initDownloadClient() {
    lockInitDownloadClient.Lock()
    defer lockInitDownloadClient.Unlock()
    if downloadClient != nil {
        return
    }
    // TODO MaxConnPerServer
    downloadClient = NewClient(10)
}

func getDownloadClient() *Client {
    if downloadClient == nil {
        initDownloadClient()
    }
    return downloadClient
}

func downloadFile(fi *bridge.File) {
    increateActiveDownload(1)
    defer increateActiveDownload(-1)
    common.Try(func() {
        logger.Debug("downloading file from other storage server, current download:", increateActiveDownload(0))
        member := selectStorageServerByInstance(fi.Instance)
        if member == nil {
            logger.Error(NO_STORAGE_ERROR)
        }
        dirty := 0
        var start int64 = 0
        buffer := make([]byte, app.BUFF_SIZE)
        for i := range fi.Parts {
            part := fi.Parts[i]
            // check if file part exists
            fInfo, e1 := os.Stat(lib_common.GetFilePathByMd5(part.Md5))
            // file part exists, skip download
            if e1 == nil || fInfo != nil {
                continue
            }
            // begin download
            som := "S"
            if len(fi.Parts) > 1 {
                som = "M"
            }
            logger.Debug("download part of " + strconv.Itoa(i) + "/" + strconv.Itoa(len(fi.Parts)) + ": /" + app.GROUP + "/" + fi.Instance + "/" + som + "/" + fi.Md5)
            e2 := download("/" + app.GROUP + "/" + fi.Instance + "/" + som + "/" + fi.Md5,
                start, part.FileSize, true, getDownloadClient(),
                func(fileLen uint64, reader io.Reader) error {
                    if uint64(part.FileSize) != fileLen {
                        return errors.New("download return wrong file length")
                    }
                    fi, e3 := lib_common.CreateTmpFile()
                    if e3 != nil {
                        return e3
                    }
                    e4 := lib_common.WriteOut(reader, int64(fileLen), buffer, fi)
                    fi.Close()
                    if e4 != nil {
                        file.Delete(fi.Name())
                        return e4
                    }
                    e5 := lib_common.MoveTmpFileTo(part.Md5, fi)
                    if e5 != nil {
                        file.Delete(fi.Name())
                        return e5
                    }
                    return nil
                })
            if e2 != nil {
                logger.Error(e2)
                dirty++
            }
            start += part.FileSize
        }
        if dirty > 0 {
            logger.Error("error download full file, broken parts:" +strconv.Itoa(dirty))
        } else {
            ee := lib_service.UpdateFileStatus(fi.Id)
            if ee != nil {
                logger.Error(ee)
            } else {
                logger.Info("download file success")
            }
        }
    }, func(i interface{}) {
        logger.Error("error download file from other storage server:", i)
    })

}

// TODO consider expire storage members
// select a storage server matching given group and instanceId
// excludes contains fail storage and not gonna use this time.
func selectStorageServerByInstance(instanceId string) *bridge.Member {
    for ele := GroupMembers.Front(); ele != nil; ele = ele.Next() {
        b := ele.Value.(*bridge.Member)
        if instanceId == b.InstanceId {
            return b
        }
    }
    return nil
}

func increateActiveDownload(value int) int {
    activeDownloadLock.Lock()
    defer activeDownloadLock.Unlock()
    activeDownload++
    return activeDownload
}