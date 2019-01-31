package libclient

import (
	"app"
	"bytes"
	"container/list"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"libcommon"
	"libcommon/bridgev2"
	"libservicev2"
	"os"
	"strconv"
	"sync"
	"time"
	"util/common"
	"util/file"
	"util/logger"
	"util/timeutil"
)

const ParallelDownload = 20
const MaxWaitDownload = 100

var GroupMembers list.List
var DownloadingFiles list.List // downloading files's id list
var memberIteLock *sync.Mutex
var addDownloadingFileLock = new(sync.Mutex)
var downloadClient *Client
var activeDownload int
var activeDownloadLock *sync.Mutex
var lockInitDownloadClient sync.Mutex
var trackerIndex = 0
var trackerIndexLock = new(sync.Mutex)

func init() {
	memberIteLock = new(sync.Mutex)
	activeDownloadLock = new(sync.Mutex)
}

type TrackerMaintainer struct {
	Collectors       list.List
	TrackerInstances list.List
	DieCallback      func(tracker string)
}

// initDownloadClient init a download client for file synchronization.
func initDownloadClient(maintainer *TrackerMaintainer) {
	lockInitDownloadClient.Lock()
	defer lockInitDownloadClient.Unlock()
	if downloadClient != nil {
		return
	}
	downloadClient = NewClient(ParallelDownload)
	downloadClient.TrackerMaintainer = maintainer
}

// getDownloadClient get download client.
func getDownloadClient() *Client {
	return downloadClient
}

// trackTaskFilter task of type TaskDownloadFiles can only put in one tracker instance.
func trackTaskFilter(allCollectors *list.List) *list.List {
	increaseTrackerIndex()
	if trackerIndex == 1 {
		return allCollectors
	}
	var ret list.List
	if allCollectors == nil {
		return nil
	}
	for ele := allCollectors.Front(); ele != nil; ele = ele.Next() {
		if !ele.Value.(*TaskCollector).Single {
			// bug: use copied object in case that the lock be use in two different goroutine which cause dead lock.
			ret.PushBack(copyTaskCollector(ele.Value.(*TaskCollector)))
		}
	}
	return &ret
}

func increaseTrackerIndex() {
	trackerIndexLock.Lock()
	defer trackerIndexLock.Unlock()
	trackerIndex++
}

// Maintain communication with tracker
// k,v => <connection string, secret>
func (maintainer *TrackerMaintainer) Maintain(trackers map[string]string) {
	if len(trackers) == 0 {
		if app.RunWith == 1 {
			logger.Warn("no trackers configured, the storage server will run in stand-alone mode.")
		} else if app.RunWith == 3 {
			logger.Warn("no trackers configured for client.")
		}
		return
	}
	for k, v := range trackers {
		go maintainer.track(k, v)
	}
}

// track connect to each tracker
func (maintainer *TrackerMaintainer) track(tracker string, secret string) {
	logger.Debug("start tracker conn with tracker server:", tracker)
	retry := 0
	// construct tracker instance and assign tasks for it
	trackerInstance := &TrackerInstance{Collectors: *trackTaskFilter(&maintainer.Collectors), ConnStr: tracker}
	trackerInstance.Init()

	// storage type client need to sync file so need a client
	if app.ClientType == 1 {
		initDownloadClient(maintainer)
	}

	// actually used by dashboard
	registerTrackerInstance(trackerInstance)
	defer unRegisterTrackerInstance(trackerInstance.ConnStr)
	defer func() {
		if maintainer.DieCallback != nil {
			maintainer.DieCallback(tracker)
		}
	}()

	for { // keep trying to connect to tracker server.
		// dashboard need to controller starting or stopping tracker instance
		if !trackerInstance.nextRun {
			break
		}
		// construct server info from tracker connection string
		serverInfo := &app.ServerInfo{}
		serverInfo.FromConnStr(tracker)
		// using new client
		client := bridgev2.NewTcpClient(serverInfo)

		e := client.Connect()

		if e == nil {
			// validate client
			respMeta, e1 := client.Validate()
			if e1 != nil {
				logger.Error("error validate with tracker", tracker+":", e1)
				client.GetConnManager().Destroy()
				// native client will break here
				if app.RunWith == 3 {
					break
				}
			} else {
				if respMeta.New4Tracker && app.RunWith == 1 {
					logger.Info("I'm new to tracker", tracker, "("+respMeta.UUID+")")
				}
				trackerInstance.Ready = true
				retry = 0
				logger.Debug("connect to tracker server success")
				trackerInstance.SetConnBridgeClient(client)
				trackerInstance.trackerUUID = respMeta.UUID
				ele := maintainer.TrackerInstances.PushBack(trackerInstance)

				for { // keep sending client statistic info to tracker server.
					// fetch a task and execute
					task := trackerInstance.GetTask()
					if task == nil {
						// controller next run
						// when tracker instance has no task and next run is disabled, this instance can stop
						if !trackerInstance.nextRun {
							logger.Debug("stop next run of tracker instance:", tracker)
							break
						}
						logger.Trace("no task available", tracker)
						time.Sleep(time.Second * 1)
						continue
					}

					// execute task
					forceClosed, e2 := trackerInstance.ExecTask(task)
					if e2 != nil {
						logger.Error("task exec error:", e2)
					} else {
						logger.Trace("task exec success:", task.TaskType)
					}

					// execute task callback func
					if task.Callback != nil {
						task.Callback(task, e2)
					}

					// if callback response is fatal, close this tracker instance
					if forceClosed {
						logger.Debug("connection is forced closed by client")
						client.GetConnManager().Destroy()
						break
					}
				}
				maintainer.TrackerInstances.Remove(ele)
				trackerInstance.Ready = false
			}
		} else {
			logger.Error("("+strconv.Itoa(retry)+") error connect to tracker server:", tracker)
			client.GetConnManager().Destroy()
		}
		retry++
		// try to connect 10 seconds later
		if trackerInstance.nextRun {
			time.Sleep(time.Second * 10)
		}
	}
}

// storeMembers storage members
func storeMembers(members []app.StorageDO) {
	memberIteLock.Lock()
	defer memberIteLock.Unlock()
	now := timeutil.GetTimestamp(time.Now())
	for e := GroupMembers.Front(); e != nil; {
		next := e.Next()
		m := e.Value.(app.StorageDO)
		if m.ExpireTime <= now {
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
			m := e.Value.(app.StorageDO)
			if a.InstanceId == m.InstanceId {
				exists = true
				m.ExpireTime = timeutil.GetTimestamp(time.Now().Add(time.Second * 61))
			}
		}
		if !exists {
			logger.Debug("add storage member server:", a)
			a.ExpireTime = timeutil.GetTimestamp(time.Now().Add(time.Second * 62))
			GroupMembers.PushBack(a)
		}
	}
}

// AddTask add timer task to list.
func AddTask(task *bridgev2.Task, tracker *TrackerInstance) bool {
	if task == nil {
		logger.Debug("can't push nil task")
		return false
	}
	logger.Trace("push task type:", strconv.Itoa(task.TaskType), "for tracker", tracker.ConnStr)
	if task.TaskType == app.TaskSyncMembers || task.TaskType == app.TaskSyncAllStorage {
		if tracker.checkTaskTypeCount(task.TaskType) == 0 {
			tracker.listIteLock.Lock()
			tracker.taskList.PushFront(task)
			tracker.listIteLock.Unlock()
			return true
		} else {
			logger.Trace("can't push task type " + strconv.Itoa(task.TaskType) + ": task type exists")
			return false
		}
	} else if task.TaskType == app.TaskRegisterFiles {
		if tracker.checkTaskTypeCount(task.TaskType) == 0 {
			tracker.listIteLock.Lock()
			tracker.taskList.PushFront(task)
			tracker.listIteLock.Unlock()
			return true
		} else {
			logger.Debug("can't push task type " + strconv.Itoa(task.TaskType) + ": task type exists")
			return false
		}
	} else if task.TaskType == app.TaskPullNewFiles {
		if tracker.checkTaskTypeCount(task.TaskType) == 0 {
			tracker.listIteLock.Lock()
			tracker.taskList.PushBack(task)
			tracker.listIteLock.Unlock()
			return true
		} else {
			logger.Debug("can't push task type 3: task type exists")
			return false
		}
	} else if task.TaskType == app.TaskDownloadFiles {
		tracker.listIteLock.Lock()
		defer tracker.listIteLock.Unlock()
		total := 0
		for e := tracker.taskList.Front(); e != nil; e = e.Next() {
			// if same download task exists then skip
			if e.Value.(*bridgev2.Task).FileId == task.FileId {
				logger.Debug("download task exists, ignore.")
				return false
			}
			if e.Value.(*bridgev2.Task).TaskType == task.TaskType {
				total++
			}
		}
		if total < MaxWaitDownload { // 限制最大并行下载任务数
			tracker.taskList.PushBack(task)
			return true
		} else {
			logger.Debug("can't push task type 4: task list full")
			return false
		}
	} else if task.TaskType == app.TaskSyncStatistic {
		if tracker.checkTaskTypeCount(task.TaskType) == 0 {
			tracker.listIteLock.Lock()
			tracker.taskList.PushFront(task)
			tracker.listIteLock.Unlock()
			return true
		} else {
			logger.Trace("can't push task type " + strconv.Itoa(task.TaskType) + ": task type exists")
			return false
		}
	}
	return false
}

// downloadFile sync file from other storage server.
func downloadFile(fullFi *app.FileVO) {

	addDownloadingFile(fullFi.Id, false)
	increaseActiveDownload(1)
	defer increaseActiveDownload(-1)
	defer addDownloadingFile(fullFi.Id, true)

	common.Try(func() {
		logger.Debug("sync file from other storage server, current download jobs:", increaseActiveDownload(0))
		// flag mark if download stream is dirty.
		dirty := 0
		// calculate md5
		md := md5.New()
		var start int64 = 0
		buffer, _ := bridgev2.MakeBytes(app.BufferSize, false, 0, false)
		defer bridgev2.RecycleBytes(buffer)

		for i := range fullFi.Parts {
			md.Reset()
			part := fullFi.Parts[i]
			// check if file part exists
			fInfo, e1 := os.Stat(libcommon.GetFilePathByMd5(part.Md5))
			// file part exists, skip download
			if e1 == nil || fInfo != nil {
				start += part.Size
				continue
			}
			// single part or multiple part.
			som := "S"
			if len(fullFi.Parts) > 1 {
				som = "M"
			}
			downloadPath := "/" + app.Group + "/" + fullFi.Instance + "/" + som + "/" + fullFi.Md5
			logger.Debug("download part of ", strconv.Itoa(i+1)+"/"+strconv.Itoa(len(fullFi.Parts)),
				": /"+app.Group+"/"+fullFi.Instance+"/"+som+"/"+fullFi.Md5, " -> ", part.Md5)

			e2 := downloadClient.Download(downloadPath,
				start,
				part.Size,
				true,
				nil,
				func(manager *bridgev2.ConnectionManager, frame *bridgev2.Frame, resMeta *bridgev2.DownloadFileResponseMeta) (b bool, e error) {

					// stream handler
					if part.Size != frame.BodyLength {
						return true, errors.New("download return wrong file length")
					}
					fi, e3 := libcommon.CreateTmpFile()
					if e3 != nil {
						logger.Debug("error create temp file")
						return true, e3
					}
					e4 := libcommon.WriteOut(manager.Conn, frame.BodyLength, buffer, fi, md)
					fi.Close()
					if e4 != nil {
						file.Delete(fi.Name())
						return true, e4
					}
					// check whether file md5 is correct.
					md5 := hex.EncodeToString(md.Sum(nil))
					if md5 != part.Md5 {
						file.Delete(fi.Name())
						return true, errors.New("part " + strconv.Itoa(i+1) + " download failed: incorrect file fingerprint: " + md5 + " but true is " + part.Md5)
					}
					e5 := libcommon.MoveTmpFileTo(part.Md5, fi)
					if e5 != nil {
						file.Delete(fi.Name())
						logger.Error("error move temp file")
						return false, e5
					}
					logger.Debug("synchronize file part success", strconv.Itoa(i+1)+"/"+strconv.Itoa(len(fullFi.Parts))+" -> "+part.Md5)
					return false, nil
				})
			if e2 != nil {
				logger.Error("error synchronize file part:", e2.Error())
				dirty++
			}
			start += part.Size
		}
		if dirty > 0 {
			logger.Error("error synchronize full file (" + fullFi.Md5 + "), broken parts:" + strconv.Itoa(dirty) + "/" + strconv.Itoa(len(fullFi.Parts)))
		} else {
			ee := libservicev2.UpdateFileFinishStatus(fullFi.Id, app.StatusEnabled, nil)
			if ee != nil {
				logger.Error("error update file finish state:", ee.Error())
			} else {
				logger.Info("synchronize file success (" + fullFi.Md5 + ")")
			}
		}
	}, func(i interface{}) {
		logger.Error("error synchronize file from other storage server:", i)
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
		buffer.WriteString(ele.Value.(app.StorageDO).InstanceId)
		if index != GroupMembers.Len()-1 {
			buffer.WriteString(",")
		}
		index++
	}
	logger.Debug("select download task file in members(" + buffer.String() + ")")
	return string(buffer.Bytes())
}

// addDownloadingFile mark file download state.
// fileId: file's id.
// remove: true is add and false is remove the mark from list.
func addDownloadingFile(fileId int64, remove bool) {
	addDownloadingFileLock.Lock()
	defer addDownloadingFileLock.Unlock()
	exist := false
	for ele := DownloadingFiles.Front(); ele != nil; ele = ele.Next() {
		if ele.Value.(int64) == fileId {
			exist = true
			break
		}
	}
	if remove {
		for ele := DownloadingFiles.Front(); ele != nil; ele = ele.Next() {
			if ele.Value.(int64) == fileId {
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

func existsDownloadingFile(fileId int64) bool {
	for ele := DownloadingFiles.Front(); ele != nil; ele = ele.Next() {
		if ele.Value.(int64) == fileId {
			return true
		}
	}
	return false
}
