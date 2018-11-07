package libdashboard

import (
	"app"
	"util/timeutil"
	"time"
	"libclient"
	"container/list"
	"util/pool"
	"util/logger"
	"net/http"
	"strconv"
	"libcommon"
	"libservice"
	"util/db"
)

// max client connection set to 1000
var p, _ = pool.NewPool(200, 0)

func StartService(config map[string]string) {
	// init db connection pool
	libservice.SetPool(db.NewPool(app.DB_POOL_SIZE))
	app.START_TIME = timeutil.GetTimestamp(time.Now())
	go startTrackerMaintainer(app.TRACKERS)
	startWebService()
}



func startTrackerMaintainer(trackers string) {

	collector := libclient.TaskCollector{
		Interval: app.SYNC_STATISTIC_INTERVAL,
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
			trackerMap[ele.Value.(string)] = app.SECRET
		}
	}
	maintainer.Maintain(trackerMap)
	go libclient.SyncTrackerAliveStatus(maintainer)
}



func startWebService() {
	http.HandleFunc("/webtracker/add", addWebTrackerHandler)
	http.HandleFunc("/webtracker/delete", deleteWebTrackerHandler)

	s := &http.Server{
		Addr: ":" + strconv.Itoa(app.HTTP_PORT),
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      0,
		MaxHeaderBytes:    1 << 20,
	}

	logger.Info("http server listen on port:", app.HTTP_PORT)
	s.ListenAndServe()
}
