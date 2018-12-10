package libstorage

import (
	"app"
	"container/list"
	"crypto/md5"
	"io"
	"libclient"
	"libcommon"
	"libcommon/bridge"
	"libservice"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"time"
	"util/common"
	"util/db"
	"util/logger"
	"util/pool"
	"util/timeutil"
)

// max client connection set to 1000
var p, _ = pool.NewPool(200, 0)

// sys secret
var secret string

// sys config
var cfg map[string]string

// tasks put in this list
var fileRegisterList list.List

// Start service and listen
// 1. Start task for upload listen
// 2. Start task for communication with tracker
func StartService(config map[string]string) {
	cfg = config
	trackers := config["trackers"]
	port := config["port"]
	secret = config["secret"]

	// set client type
	app.CLIENT_TYPE = 1
	app.START_TIME = timeutil.GetTimestamp(time.Now())
	app.DOWNLOADS = 0
	app.UPLOADS = 0
	app.IOIN = 0
	app.IOOUT = 0
	app.STAGE_DOWNLOADS = 0
	app.STAGE_UPLOADS = 0
	app.STAGE_IOIN = 0
	app.STAGE_IOOUT = 0
	app.FILE_TOTAL = 0
	app.FILE_FINISH = 0

	// init db connection pool
	libservice.SetPool(db.NewPool(app.DB_POOL_SIZE))
	newUUID := common.UUID()
	logger.Debug("generate UUID:", newUUID)
	e1 := libservice.ConfirmLocalInstanceUUID(newUUID)
	if e1 != nil {
		logger.Fatal("error persist local instance uuid:", e1)
	}

	uuid, e2 := libservice.GetLocalInstanceUUID()
	if e2 != nil {
		logger.Fatal("error fetch local instance uuid:", e2)
	}
	app.UUID = uuid
	logger.Info("instance start with uuid:", app.UUID)
	if app.INSTANCE_ID == "" {
		// 2018/12/10 if instance_id not set, use app.UUID as instance_id instead.
		// this feature is mainly used in the docker clustering environment.
		logger.Warn("app instance_id not set, use app.UUID as instance_id instead")
		app.INSTANCE_ID = app.UUID
	}

	initDbStatistic()
	// start statistic service.
	go startStatisticService()

	startHttpDownloadService()
	go startTrackerMaintainer(trackers)
	startStorageService(port)
}

func startTrackerMaintainer(trackers string) {
	collector1 := libclient.TaskCollector{
		Interval: time.Second * 10,
		Name:     "推送本地新文件到tracker",
		Single:   false,
		Job:      libclient.QueryPushFileTaskCollector,
	}
	collector2 := libclient.TaskCollector{
		Interval:   time.Second * 10,
		Name:       "查询待同步文件",
		Single:     true,
		FirstDelay: time.Second * 1,
		Job:        libclient.QueryDownloadFileTaskCollector,
	}
	collector3 := libclient.TaskCollector{
		Interval: app.PULL_NEW_FILE_INTERVAL,
		Name:     "拉取tracker新文件",
		Single:   false,
		Job:      libclient.QueryNewFileTaskCollector,
	}
	collector4 := libclient.TaskCollector{
		Interval: app.SYNC_MEMBER_INTERVAL,
		Name:     "同步storage成员",
		Single:   false,
		Job:      libclient.SyncMemberTaskCollector,
	}
	collectors := *new(list.List)
	collectors.PushBack(&collector1)
	collectors.PushBack(&collector2)
	collectors.PushBack(&collector3)
	collectors.PushBack(&collector4)

	ls := libcommon.ParseTrackers(trackers)
	trackerMap := make(map[string]string)
	if ls != nil {
		for ele := ls.Front(); ele != nil; ele = ele.Next() {
			trackerMap[ele.Value.(string)] = app.SECRET
		}
	}
	maintainer := &libclient.TrackerMaintainer{Collectors: collectors}
	maintainer.Maintain(trackerMap)
}

// upload listen
func startStorageService(port string) {
	pt := libcommon.ParsePort(port)
	if pt > 0 {
		tryTimes := 0
		for {
			common.Try(func() {
				listener, e := net.Listen("tcp", ":"+strconv.Itoa(pt))
				logger.Info("service listening on port:", pt)
				if e != nil {
					panic(e)
				} else {
					// keep accept connections.
					for {
						conn, e1 := listener.Accept()
						if e1 == nil {
							ee := p.Exec(func() {
								clientHandler(conn)
							})
							// maybe the poll is full
							if ee != nil {
								logger.Error(ee)
								bridge.Close(conn)
							}
						} else {
							logger.Info("accept new conn error", e1)
							if conn != nil {
								bridge.Close(conn)
							}
						}
					}
				}
			}, func(i interface{}) {
				logger.Error("["+strconv.Itoa(tryTimes)+"] error shutdown service duo to:", i)
				time.Sleep(time.Second * 10)
			})
		}
	}
}

// http download service
func startHttpDownloadService() {
	if !app.HTTP_ENABLE {
		logger.Info("http server disabled")
		return
	}

	http.HandleFunc("/download/", DownloadHandler)
	if app.UPLOAD_ENABLE {
		http.HandleFunc("/upload", WebUploadHandlerV1)
	} else {
		logger.Info("upload is disabled.")
	}

	s := &http.Server{
		Addr: ":" + strconv.Itoa(app.HTTP_PORT),
		// ReadTimeout:    10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      0,
		MaxHeaderBytes:    1 << 20,
	}
	logger.Info("http server listen on port:", app.HTTP_PORT)
	go s.ListenAndServe()
}

// accept a new connection for file upload
// the connection will keep till it is broken
// 文件同步策略：
// 文件上传成功将任务写到本地文件storage_task.data作为备份
// 将任务通知到tracker服务器，通知成功，tracker服务进行广播
// 其他storage定时取任务，将任务
func clientHandler(conn net.Conn) {
	defer func() {
		logger.Info("close connection from server")
		bridge.Close(conn)
	}()
	common.Try(func() {
		// calculate md5
		md := md5.New()
		connBridge := bridge.NewBridge(conn)
		for {
			error := connBridge.ReceiveRequest(func(request *bridge.Meta, in io.ReadCloser) error {
				// return requestRouter(request, &bodyBuff, md, connBridge, conn)
				if request.Err != nil {
					return request.Err
				}
				// route
				if request.Operation == bridge.O_CONNECT {
					return validateClientHandler(request, connBridge)
				} else if request.Operation == bridge.O_UPLOAD {
					return uploadHandler(request, md, conn, connBridge)
				} else if request.Operation == bridge.O_QUERY_FILE {
					return QueryFileHandler(request, connBridge, 1)
				} else if request.Operation == bridge.O_DOWNLOAD_FILE {
					return downloadFileHandler(request, connBridge)
				} else {
					return bridge.OPERATION_NOT_SUPPORT_ERROR
				}
			})
			if error != nil {
				logger.Error(error)
				break
			}
		}
	}, func(i interface{}) {
		logger.Error("connection error:", i)
	})
}

func initDbStatistic() {
	for {
		files, finish, disk, e := libservice.QueryStatistic()
		if e != nil {
			logger.Error("error query statistic info:", e)
			continue
		} else {
			app.FILE_TOTAL = files
			app.FILE_FINISH = finish
			app.DISK_USAGE = disk
			logger.Info(":::statistic:::")
			logger.Info("+---------------------------+")
			logger.Info("* file count       :", app.FILE_TOTAL)
			logger.Info("* sync finish count:", app.FILE_FINISH)
			logger.Info("* disk usage       :", libcommon.HumanReadable(app.DISK_USAGE, 1000)) //TODO not right
			logger.Info("+---------------------------+")
			break
		}
	}
}

func startStatisticService() {
	timer := time.NewTicker(time.Second * 30)
	for {
		stats := &runtime.MemStats{}
		runtime.ReadMemStats(stats)
		app.MEMORY = stats.Sys
		<-timer.C
	}
}
