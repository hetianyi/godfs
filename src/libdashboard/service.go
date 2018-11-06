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
)

// max client connection set to 1000
var p, _ = pool.NewPool(200, 0)

func StartService(config map[string]string) {
	app.START_TIME = timeutil.GetTimestamp(time.Now())
	go startTrackerMaintainer(app.TRACKERS)
	startWebService()
}



func startTrackerMaintainer(trackers string) {

	collector := libclient.TaskCollector{
		Interval: app.PULL_NEW_FILE_INTERVAL,
		Name:     "同步Statistic",
		Single:   false,
		Job:      libclient.SyncStatisticTaskCollector,
	}
	collectors := *new(list.List)
	collectors.PushBack(&collector)
	maintainer := &libclient.TrackerMaintainer{Collectors: collectors}
	maintainer.Maintain(trackers)
}



func startWebService() {
	//http.HandleFunc("/download/", DownloadHandler)
	s := &http.Server{
		Addr: ":" + strconv.Itoa(app.HTTP_PORT),
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      0,
		MaxHeaderBytes:    1 << 20,
	}
	logger.Info("http server listen on port:", app.HTTP_PORT)
	s.ListenAndServe()
}
