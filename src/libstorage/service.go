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

// StartService Start service and listen
// 1. Start task for upload listen
// 2. Start task for communication with tracker
func StartService(config map[string]string) {
	trackers := config["trackers"]

	// set client type
	app.ClientType = 1
	app.StartTime = timeutil.GetTimestamp(time.Now())
	app.Downloads = 0
	app.Uploads = 0
	app.IOIn = 0
	app.IOOut = 0
	app.StageDownloads = 0
	app.StageUploads = 0
	app.StageIOIn = 0
	app.StageIOOut = 0
	app.TotalFiles = 0
	app.FinishFiles = 0

	// init db connection pool
	libservicev2.SetPool(db.NewPool(app.DbPoolSize))
	newUUID := common.UUID()
	logger.Debug("generate UUID:", newUUID)
	uuid, e1 := libservicev2.ConfirmAppUUID(newUUID)
	if e1 != nil {
		logger.Fatal("error query/persist instance uuid:", e1)
	}

	app.UUID = uuid
	logger.Info("instance start with uuid:", app.UUID)
	if app.InstanceId == "" {
		// 2018/12/10 if instance_id not set, use app.UUID as instance_id instead.
		// this feature is mainly used in the docker clustering environment.
		logger.Warn("app instance_id not set, use app.UUID as instance_id instead")
		app.InstanceId = app.UUID
	}

	initDbStatistic()
	// start statistic service.
	go startStatisticsService()
	go app.BusyPointService()
	// start http download server
	startHttpService()
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
		Interval: app.PullNewFileInterval,
		Name:     "PULL NEW FILES",
		Single:   false,
		Job:      libclient.QueryNewFileTaskCollector,
	}
	collector4 := libclient.TaskCollector{
		Interval: app.SyncMemberInterval,
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
			trackerMap[ele.Value.(string)] = app.Secret
		}
	}
	// TODO use client object, do not use it directly
	maintainer := &libclient.TrackerMaintainer{Collectors: collectors}
	maintainer.Maintain(trackerMap)
}

// startStorageService storage server start tcp listen.
func startStorageService() {
	server := bridgev2.NewServer("", app.Port)
	server.Listen()
}

// startHttpService start http download server.
func startHttpService() {
	if !app.HttpEnable {
		logger.Info("http server disabled")
		return
	}

	http.HandleFunc("/download/", DownloadHandler)
	if app.UploadEnable {
		http.HandleFunc("/upload", WebUploadHandlerV1)
		http.HandleFunc("/upload/", WebUploadHandlerV1)
	} else {
		logger.Info("upload is disabled.")
	}

	s := &http.Server{
		Addr: ":" + strconv.Itoa(app.HttpPort),
		// ReadTimeout:    10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      0,
		MaxHeaderBytes:    1 << 20,
	}
	logger.Info("http server listening on port:", app.HttpPort)
	go s.ListenAndServe()
}

// initDbStatistic init system statistic info before boot.
func initDbStatistic() {
	for {
		statistic, e := libservicev2.QuerySystemStatistic()
		if e != nil {
			logger.Error("error query statistic info:", e)
			time.Sleep(time.Second * 5)
			continue
		} else {
			app.TotalFiles = statistic.FileCount
			app.FinishFiles = statistic.FinishCount
			app.DiskUsage = statistic.DiskSpace
			logger.Info(":::statistic:::")
			logger.Info("+---------------------------+")
			logger.Info("* file count       :", app.TotalFiles)
			logger.Info("* sync finish count:", app.FinishFiles)
			logger.Info("* disk usage       :", libcommon.HumanReadable(app.DiskUsage, 1000))
			logger.Info("+---------------------------+")
			break
		}
	}
}

// startStatisticService collect system info.
func startStatisticsService() {
	timer := time.NewTicker(time.Second * 30)
	for {
		stats := &runtime.MemStats{}
		runtime.ReadMemStats(stats)
		app.Memory = stats.Sys
		<-timer.C
	}
}
