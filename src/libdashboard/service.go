package libdashboard

import (
	"app"
	"container/list"
	"libclient"
	"libcommon"
	"libservice"
	"net/http"
	"strconv"
	"time"
	"util/db"
	"util/logger"
	"util/pool"
	"util/timeutil"
)

// max client connection set to 1000
var p, _ = pool.NewPool(200, 0)

func StartService(config map[string]string) {
	// init db connection pool
	libservice.SetPool(db.NewPool(app.DbPoolSize))
	app.StartTime = timeutil.GetTimestamp(time.Now())
	go startTrackerMaintainer(app.Trackers)
	startWebService()
}

func startTrackerMaintainer(trackers string) {

	collector := libclient.TaskCollector{
		Interval: app.SyncStatisticInterval,
		Name:     "同步Statistic",
		Single:   false,
		Job:      libclient.SyncStatisticTaskCollector,
	}
	collectors := *new(list.List)
	collectors.PushBack(&collector)
	maintainer := &libclient.TrackerMaintainer{Collectors: collectors}

	ls := libcommon.ParseTrackers(trackers)
	trackerMap := make(map[string]string)
	if ls != nil {
		for ele := ls.Front(); ele != nil; ele = ele.Next() {
			trackerMap[ele.Value.(string)] = app.Secret
		}
	}
	maintainer.Maintain(trackerMap)
	go libclient.SyncTrackerAliveStatus(maintainer)
}

func startWebService() {
	http.HandleFunc("/dashboard/webtracker/add", addWebTrackerHandler)
	http.HandleFunc("/dashboard/webtracker/delete", deleteWebTrackerHandler)
	http.HandleFunc("/dashboard/index", indexStatistic)

	s := &http.Server{
		Addr:              ":" + strconv.Itoa(app.HttpPort),
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      0,
		MaxHeaderBytes:    1 << 20,
	}

	logger.Info("http server listen on port:", app.HttpPort)
	s.ListenAndServe()
}
