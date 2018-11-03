package libdashboard

import (
	"app"
	"util/timeutil"
	"time"
	"libclient"
	"container/list"
	"util/logger"
	"libcommon/bridge"
	"net"
	"strconv"
	"util/common"
)

func StartService(config map[string]string) {
	// set client type
	app.CLIENT_TYPE = 3
	app.START_TIME = timeutil.GetTimestamp(time.Now())
	go startTrackerMaintainer("127.0.0.1:1022")
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
	tryTimes := 0
	for {
		common.Try(func() {
			listener, e := net.Listen("tcp", ":"+strconv.Itoa(8080))
			logger.Info("service listening on port:", 8080)
			if e != nil {
				panic(e)
			} else {
				// keep accept connections.
				for {
					conn, _ := listener.Accept()
					bridge.Close(conn)
				}
			}
		}, func(i interface{}) {
			logger.Error("["+strconv.Itoa(tryTimes)+"] error shutdown service duo to:", i)
			time.Sleep(time.Second * 10)
		})
	}
}
