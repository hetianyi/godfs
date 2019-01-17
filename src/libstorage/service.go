package libstorage

import (
	"app"
	"container/list"
	"libclient"
	"libcommon"
	"libcommon/bridgev2"
	"libservicev2"
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

// Start service and listen
// 1. Start task for upload listen
// 2. Start task for communication with tracker
func StartService(config map[string]string) {
	trackers := config["trackers"]

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
	libservicev2.SetPool(db.NewPool(app.DB_POOL_SIZE))
	newUUID := common.UUID()
	logger.Debug("generate UUID:", newUUID)
	uuid, e1 := libservicev2.ConfirmAppUUID(newUUID)
	if e1 != nil {
		logger.Fatal("error query/persist instance uuid:", e1)
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
	// start http download server
	startHttpDownloadService()
	// start tracker maintainer
	go startTrackerMaintainer(trackers)
	// start storage server tcp service
	startStorageService()
}

func startTrackerMaintainer(trackers string) {
	collector1 := libclient.TaskCollector{
		Interval: time.Second * 10,
		Name:     "REGISTER FILES",
		Single:   false,
		Job:      libclient.QueryPushFileTaskCollector,
	}
	collector2 := libclient.TaskCollector{
		Interval:   time.Second * 10,
		Name:       "QUERY READY PUSH FILES",
		Single:     true,
		FirstDelay: time.Second * 1,
		Job:        libclient.QueryDownloadFileTaskCollector,
	}
	collector3 := libclient.TaskCollector{
		Interval: app.PULL_NEW_FILE_INTERVAL,
		Name:     "PULL NEW FILES",
		Single:   false,
		Job:      libclient.QueryNewFileTaskCollector,
	}
	collector4 := libclient.TaskCollector{
		Interval: app.SYNC_MEMBER_INTERVAL,
		Name:     "SYNCHRONIZED MEMBERS",
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
	// TODO use client object, do not use it directly
	maintainer := &libclient.TrackerMaintainer{Collectors: collectors}
	maintainer.Maintain(trackerMap)
}

// storage server start tcp listen.
func startStorageService() {
	server := bridgev2.NewServer("", app.PORT)
	server.Listen()
}

// start http download server.
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
	logger.Info("http server listening on port:", app.HTTP_PORT)
	go s.ListenAndServe()
}

// init system statistic info before boot.
func initDbStatistic() {
	for {
		statistic, e := libservicev2.QuerySystemStatistic()
		if e != nil {
			logger.Error("error query statistic info:", e)
			time.Sleep(time.Second * 5)
			continue
		} else {
			app.FILE_TOTAL = statistic.FileCount
			app.FILE_FINISH = statistic.FinishCount
			app.DISK_USAGE = statistic.DiskSpace
			logger.Info(":::statistic:::")
			logger.Info("+---------------------------+")
			logger.Info("* file count       :", app.FILE_TOTAL)
			logger.Info("* sync finish count:", app.FILE_FINISH)
			logger.Info("* disk usage       :", libcommon.HumanReadable(app.DISK_USAGE, 1000))
			logger.Info("+---------------------------+")
			break
		}
	}
}

// collect system info.
func startStatisticService() {
	timer := time.NewTicker(time.Second * 30)
	for {
		stats := &runtime.MemStats{}
		runtime.ReadMemStats(stats)
		app.MEMORY = stats.Sys
		<-timer.C
	}
}
