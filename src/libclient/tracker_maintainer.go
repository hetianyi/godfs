package libclient

import (
	"app"
	"bytes"
	"container/list"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"libcommon"
	"libcommon/bridge"
	"libservice"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
	"util/common"
	"util/file"
	"util/logger"
	"util/timeutil"
)

// storage 的任务分为：
// type 1: 定期从tracker同步members（定时任务，非持久化任务，插队任务，高优先级，任务列表中只能存在一条此类型的任务）
// type 2: 上报文件给tracker（定时任务，持久化任务，插队任务，高优先级）
// type 3: 定期向tracker服务器查询最新文件列表（定时任务，非持久化任务，插队任务，高优先级，任务列表中只能存在一条此类型的任务）
// type 4: 从其他group节点下载文件（定时任务，持久化任务，最低优先级，goroutine执行）
const ParallelDownload = 50
const MaxWaitDownload = 100

var GroupMembers list.List
var DownloadingFiles list.List // 正在下载的文件Id列表
var memberIteLock *sync.Mutex
var addDownloadingFileLock = new(sync.Mutex)
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
	Collectors       list.List
	TrackerInstances list.List
}

type ITracker interface {
	Init()
	SetConnBridge()
	GetTaskSize() int
	GetTask() *bridge.Task
	//FailReturnTask(task *bridge.Task)
	checkTaskTypeCount(taskType int)
	startTaskCollector()
	ExecTask(task *bridge.Task, connBridge *bridge.Bridge) (bool, error)
}

type TrackerInstance struct {
	taskList    list.List
	listIteLock *sync.Mutex
	connBridge  *bridge.Bridge
	Collectors  list.List
	Ready       bool
}

type Collector interface {
	Start(tracker *TrackerInstance)
}

type TaskCollector struct {
	startLock  sync.Mutex                     //if the timer is already started
	Interval   time.Duration                  // time in Milliseconds, task exec interval.
	FirstDelay time.Duration                  // time in Milliseconds, task exec first time delay.
	ExecTimes  int                            // the collector execute times, ExecTimes<=0 means never stop
	Name       string                         // collector name
	Single     bool                           // 是否是能一个实例运行
	Job        func(tracker *TrackerInstance) // timer task
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
func (maintainer *TrackerMaintainer) Maintain(trackers string) *list.List {
	ls := libcommon.ParseTrackers(trackers)
	if ls.Len() == 0 {
		if app.RUN_WITH == 1 {
			logger.Warn("no trackers configured, the storage server will run in stand-alone mode.")
		} else if app.RUN_WITH == 3 {
			logger.Warn("no trackers configured for client.")
		}
		return ls
	}
	index := 0
	for e := ls.Front(); e != nil; e = e.Next() {
		go maintainer.track(e.Value.(string), index)
		index++
	}
	return ls
}

// connect to each tracker
func (maintainer *TrackerMaintainer) track(tracker string, index int) {
	logger.Debug("start tracker conn with tracker server:", tracker)
	retry := 0
	trackerInstance := &TrackerInstance{Collectors: *trackTaskFilter(&maintainer.Collectors, index)}
	trackerInstance.Init()
	initDownloadClient(maintainer)
	// for test
	//go startTimer1()
	for { //keep trying to connect to tracker server.
		conn, e := net.Dial("tcp", tracker)
		if e == nil {
			// validate client
			connBridge, e1 := connectAndValidate(conn)
			if e1 != nil {
				bridge.Close(conn)
				logger.Error(e1)
				if app.RUN_WITH == 3 {
					break
				}
			} else {
				/*trackerConfig, te := lib_service.GetTrackerConfig(connBridge.UUID)
				  if te != nil {
				      logger.Error(te)
				      time.Sleep(time.Second * 10)
				      continue
				  }
				  if trackerConfig == nil {
				      trackerConfig = &bridge.TrackerConfig{UUID: connBridge.UUID, MasterSyncId: 0, LocalPushId: 0}
				  }*/
				ele := maintainer.TrackerInstances.PushBack(trackerInstance)
				trackerInstance.Ready = true
				retry = 0
				logger.Debug("connect to tracker server success.")
				trackerInstance.SetConnBridge(connBridge)
				for { // keep sending client statistic info to tracker server.
					task := trackerInstance.GetTask()
					if task == nil {
						time.Sleep(time.Second * 1)
						continue
					}
					forceClosed, e2 := trackerInstance.ExecTask(task)
					if e2 != nil {
						logger.Error("task exec error:", e2)
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
				maintainer.TrackerInstances.Remove(ele)
				trackerInstance.Ready = false
			}
		} else {
			logger.Error("("+strconv.Itoa(retry)+")error connect to tracker server:", tracker)
		}
		retry++
		time.Sleep(time.Second * 10)
	}
}

// connect to tracker server and register client to it.
func connectAndValidate(conn net.Conn) (*bridge.Bridge, error) {
	// create bridge
	connBridge := bridge.NewBridge(conn)
	// send validate request
	isNew, e1 := connBridge.ValidateConnection("")
	if e1 != nil {
		connBridge.Close()
		return nil, e1
	}
	if isNew && app.CLIENT_TYPE == 1 {
		logger.Info("I'm new to tracker:", connBridge.GetConn().RemoteAddr().String(), "[", connBridge.UUID, "]")
		e2 := libservice.UpdateTrackerSyncId(connBridge.UUID, 0, nil)
		if e2 != nil {
			connBridge.Close()
			return nil, e2
		}
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
		if collector.ExecTimes > 0 && execTimes >= collector.ExecTimes {
			logger.Debug("stop collector \"" + collector.Name + "\" because of max execute times reached.")
			timer.Stop()
			break
		}
		time.Sleep(collector.FirstDelay)
		if collector.Name != "" {
			logger.Trace("exec task collector:", collector.Name)
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

func (tracker *TrackerInstance) Init() {
	tracker.listIteLock = new(sync.Mutex)
	tracker.startTaskCollector()
}

func (tracker *TrackerInstance) SetConnBridge(connBridge *bridge.Bridge) {
	tracker.connBridge = connBridge
}

// get task size in waiting list
func (tracker *TrackerInstance) GetTaskSize() int {
	return tracker.taskList.Len()
}

func addMember(members []bridge.Member) {
	memberIteLock.Lock()
	defer memberIteLock.Unlock()
	now := timeutil.GetTimestamp(time.Now())
	for e := GroupMembers.Front(); e != nil; {
		next := e.Next()
		m := e.Value.(*bridge.ExpireMember)
		if timeutil.GetTimestamp(m.ExpireTime) <= now {
			GroupMembers.Remove(e)
		}
		e = next
	}
	if members == nil {
		return
	}
	for i := range members {
		a := members[i]
		exists := false
		for e := GroupMembers.Front(); e != nil; e = e.Next() {
			m := e.Value.(*bridge.ExpireMember)
			if a.InstanceId == m.InstanceId {
				exists = true
				m.ExpireTime = time.Now().Add(time.Second * 61)
			}
		}
		if !exists {
			logger.Debug("add storage member server:", a)
			b := &bridge.ExpireMember{}
			b.From(&a)
			b.ExpireTime = time.Now().Add(time.Second * 61)
			GroupMembers.PushBack(b)
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
			logger.Trace("can't push task type " + strconv.Itoa(task.TaskType) + ": task type exists")
			return false
		}
	} else if task.TaskType == app.TASK_REPORT_FILE {
		if tracker.checkTaskTypeCount(task.TaskType) == 0 {
			logger.Trace("push task type 2")
			if tracker.taskList.Front() != nil && tracker.taskList.Front().Value.(*bridge.Task).TaskType == app.TASK_SYNC_MEMBER {
				tracker.taskList.InsertAfter(task, tracker.taskList.Front())
			} else {
				tracker.taskList.PushFront(task)
			}
			return true
		} else {
			logger.Debug("can't push task type " + strconv.Itoa(task.TaskType) + ": task type exists")
			return false
		}
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
			// if same download task exists then skip
			if e.Value.(*bridge.Task).FileId == task.FileId {
				logger.Debug("download task exists, ignore.")
				return false
			}
			if e.Value.(*bridge.Task).TaskType == task.TaskType {
				total++
			}
		}
		if total < MaxWaitDownload { // 限制最大并行下载任务数
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
		regClientMeta := &bridge.OperationRegisterStorageClientRequest{
			BindAddr:   app.BIND_ADDRESS,
			Group:      app.GROUP,
			InstanceId: app.INSTANCE_ID,
			Port:       app.PORT,
			TotalFiles: app.FILE_TOTAL,
			Finish:     app.FILE_FINISH,
			IOin:       app.IOIN,
			IOout:      app.IOOUT,
			DiskUsage:  app.DISK_USAGE,
			Downloads:  app.DOWNLOADS,
			Uploads:    app.UPLOADS,
			StartTime:  app.START_TIME,
			Memory:     app.MEMORY,
			ReadOnly:   !app.UPLOAD_ENABLE,
		}
		// reg client
		e2 := connBridge.SendRequest(bridge.O_SYNC_MEMBERS, regClientMeta, 0, nil)
		if e2 != nil {
			return true, e2
		}
		e5 := connBridge.ReceiveResponse(func(response *bridge.Meta, in io.Reader) error {
			if response.Err != nil {
				return response.Err
			}
			//logger.Debug(string(response.MetaBody))
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
		files, e1 := libservice.GetFilesBasedOnId(task.FileId)
		if e1 != nil {
			return false, e1
		}
		if files == nil || files.Len() == 0 {
			return false, nil
		}
		fs := make([]bridge.File, files.Len())
		i := 0
		maxId := 0
		for ele := files.Front(); ele != nil; ele = ele.Next() {
			fs[i] = *ele.Value.(*bridge.File)
			if maxId < fs[i].Id {
				maxId = fs[i].Id
			}
			i++
		}
		// register storage client to tracker server
		regFileMeta := &bridge.OperationRegisterFileRequest{
			Files: fs,
		}
		logger.Info("register", files.Len(), "files to tracker server")
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
				return errors.New("error register file " + strconv.Itoa(task.FileId) + " to tracker server, server response status:" + strconv.Itoa(regResp.Status))
			}
			// update table trackers and set local_push_fid to new id
			e7 := libservice.FinishLocalFilePushTask(maxId, tracker.connBridge.UUID)
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
		config, e1 := libservice.GetTrackerConfig(tracker.connBridge.UUID)
		if e1 != nil {
			return false, e1
		}
		if config == nil {
			config = &bridge.TrackerConfig{TrackerSyncId: 0}
		}
		// register storage client to tracker server
		pullMeta := &bridge.OperationPullFileRequest{
			BaseId: config.TrackerSyncId,
		}
		logger.Debug("try to pull new file from tracker server:", tracker.connBridge.GetConn().RemoteAddr().String(), ", base id is", config.TrackerSyncId)
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
				return errors.New("error register file " + strconv.Itoa(task.FileId) + " to tracker server, server response status:" + strconv.Itoa(pullResp.Status))
			}

			files := pullResp.Files
			if len(files) > 0 {
				logger.Info("pull", len(files), "files from tracker server:", tracker.connBridge.GetConn().RemoteAddr().String())
			} else {
				logger.Debug("no file pull from tracker server:", tracker.connBridge.GetConn().RemoteAddr().String())
			}
			return libservice.StorageAddTrackerPulledFile(files, tracker.connBridge.UUID)
		})
		if e5 != nil {
			return true, e5
		}
		return false, nil
	} else if task.TaskType == app.TASK_DOWNLOAD_FILE {
		logger.Debug("trying download file from other storage server...")
		if increaseActiveDownload(0) >= ParallelDownload {
			logger.Debug("ParallelDownload reached")
			//AddTask(task, tracker)
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

//TODO ERROR - 2018-08-20 09:46:03,879 [tracker_maintainer.go:232] task collector "推送本地新文件到tracker" return error: interface conversion: interface {} is nil, not *bridge.Task
//TODO  ERROR - 2018-08-20 09:46:03,901 [tracker_maintainer.go:232] task collector "拉取tracker新文件" return error: interface conversion: interface {} is nil, not *bridge.Task

// 查询推送文件到tracker的任务收集器
func QueryPushFileTaskCollector(tracker *TrackerInstance) {
	if tracker.connBridge == nil {
		return
	}
	task, e1 := libservice.GetLocalPushFileTask(app.TASK_REPORT_FILE, tracker.connBridge.UUID)
	if e1 != nil {
		logger.Error(e1)
		return
	}
	if task != nil {
		AddTask(task, tracker)
	}
}

// TODO 标记多次下载失败的任务文件
// 查询本地持久化任务收集器
func QueryDownloadFileTaskCollector(tracker *TrackerInstance) {
	members := collectMemberInstanceId()
	// no member, no server for download.
	if members == "" {
		return
	}
	taskList, e1 := libservice.GetDownloadFileTask(app.TASK_DOWNLOAD_FILE)
	if e1 != nil {
		logger.Error(e1)
		return
	}
	if taskList.Len() == 0 {
		logger.Debug("no file need to sync.")
	}
	for e := taskList.Front(); e != nil; e = e.Next() {
		if !existsDownloadingFile(e.Value.(*bridge.Task).FileId) {
			AddTask(e.Value.(*bridge.Task), tracker)
		}
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

func initDownloadClient(maintainer *TrackerMaintainer) {
	lockInitDownloadClient.Lock()
	defer lockInitDownloadClient.Unlock()
	if downloadClient != nil {
		return
	}
	// TODO MaxConnPerServer
	downloadClient = NewClient(ParallelDownload)
	downloadClient.TrackerMaintainer = maintainer
}

func getDownloadClient() *Client {
	return downloadClient
}

// TODO download from source first
func downloadFile(fullFi *bridge.File) {
	increaseActiveDownload(1)
	defer increaseActiveDownload(-1)
	defer addDownloadingFile(fullFi.Id, true)
	common.Try(func() {
		logger.Debug("sync file from other storage server, current download thread:", increaseActiveDownload(0))
		dirty := 0
		// calculate md5
		md := md5.New()
		var start int64 = 0
		buffer, _ := bridge.MakeBytes(app.BUFF_SIZE, false, 0, false)
		defer bridge.RecycleBytes(buffer)
		for i := range fullFi.Parts {
			md.Reset()
			part := fullFi.Parts[i]
			// check if file part exists
			fInfo, e1 := os.Stat(libcommon.GetFilePathByMd5(part.Md5))
			// file part exists, skip download
			if e1 == nil || fInfo != nil {
				start += part.FileSize
				continue
			}
			// begin download
			som := "S"
			if len(fullFi.Parts) > 1 {
				som = "M"
			}
			logger.Debug("download part of ", strconv.Itoa(i+1)+"/"+strconv.Itoa(len(fullFi.Parts)), ": /"+app.GROUP+"/"+fullFi.Instance+"/"+som+"/"+fullFi.Md5, " -> ", part.Md5)
			e2 := download("/"+app.GROUP+"/"+fullFi.Instance+"/"+som+"/"+fullFi.Md5,
				start, part.FileSize, true, new(list.List), getDownloadClient(),
				func(realPath string, fileLen uint64, reader io.Reader) error {
					if uint64(part.FileSize) != fileLen {
						return errors.New("download return wrong file length")
					}
					fi, e3 := libcommon.CreateTmpFile()
					if e3 != nil {
						return e3
					}
					e4 := libcommon.WriteOut(reader, int64(fileLen), buffer, fi, md, increDownloadBytes)
					fi.Close()
					if e4 != nil {
						file.Delete(fi.Name())
						return e4
					}
					// check whether file md5 is correct. TODO need to test
					md5 := hex.EncodeToString(md.Sum(nil))
					if md5 != part.Md5 {
						file.Delete(fi.Name())
						return errors.New("part " + strconv.Itoa(i+1) + "download error: file fingerprint confirm failed: " + md5 + " but true is " + part.Md5)
					}
					e5 := libcommon.MoveTmpFileTo(part.Md5, fi)
					if e5 != nil {
						file.Delete(fi.Name())
						return e5
					}
					logger.Debug("download part success", strconv.Itoa(i+1)+"/"+strconv.Itoa(len(fullFi.Parts))+" ->"+part.Md5)
					return nil
				})
			if e2 != nil {
				logger.Error(e2)
				dirty++
			}
			start += part.FileSize
		}
		if dirty > 0 {
			logger.Error("error download full file, broken parts:" + strconv.Itoa(dirty) + "/" + strconv.Itoa(len(fullFi.Parts)))
		} else {
			ee := libservice.UpdateFileStatus(fullFi.Id)
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

func increaseActiveDownload(value int) int {
	activeDownloadLock.Lock()
	defer activeDownloadLock.Unlock()
	activeDownload += value
	return activeDownload
}

func collectMemberInstanceId() string {
	memberIteLock.Lock()
	defer memberIteLock.Unlock()
	var buffer bytes.Buffer
	index := 0
	for ele := GroupMembers.Front(); ele != nil; ele = ele.Next() {
		buffer.WriteString(ele.Value.(*bridge.ExpireMember).InstanceId)
		if index != GroupMembers.Len()-1 {
			buffer.WriteString(",")
		}
	}
	logger.Debug("select download task file in members(" + buffer.String() + ")")
	return string(buffer.Bytes())
}

func addDownloadingFile(fileId int, remove bool) {
	addDownloadingFileLock.Lock()
	defer addDownloadingFileLock.Unlock()
	exist := false
	for ele := DownloadingFiles.Front(); ele != nil; ele = ele.Next() {
		if ele.Value.(int) == fileId {
			exist = true
		}
	}
	if remove {
		for ele := DownloadingFiles.Front(); ele != nil; ele = ele.Next() {
			if ele.Value.(int) == fileId {
				DownloadingFiles.Remove(ele)
				break
			}
		}
	} else {
		if !exist {
			DownloadingFiles.PushBack(fileId)
		}
	}
}

func existsDownloadingFile(fileId int) bool {
	for ele := DownloadingFiles.Front(); ele != nil; ele = ele.Next() {
		if ele.Value.(int) == fileId {
			return true
		}
	}
	return false
}

// for test
var increLock sync.Mutex
var lastBytesSecond = 0

func increDownloadBytes(n int) {
	increLock.Lock()
	defer increLock.Unlock()
	lastBytesSecond += n
}

func startTimer1() {
	timer := time.NewTicker(time.Second)
	for {
		reportDownloadStatus()
		lastBytesSecond = 0
		<-timer.C
	}
}

func reportDownloadStatus() {
	fmt.Print("\n\n-----------------------------------------------------------------\n")
	fmt.Print("连接总数：", downloadClient.connPool.totalActiveConn, ", 当前并行下载数：", increaseActiveDownload(0), ", 当前下载速率：", lastBytesSecond/1024, "kb/s")
	fmt.Print("\n-----------------------------------------------------------------\n\n")
}
