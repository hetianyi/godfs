package libtracker

import (
	"app"
	"libcommon"
	"libcommon/bridgev2"
	"libservicev2"
	"util/common"
	"util/db"
	"util/logger"
	"util/pool"
	"net/http"
	"strconv"
	"time"
)

var p, _ = pool.NewPool(500, 0)

// Start service and listen
// 1. Start task for upload listen
// 2. Start task for communication with tracker
func StartService() {
	// prepare db connection pool
	libservicev2.SetPool(db.NewPool(app.DB_POOL_SIZE))

	uuid, e1 := libservicev2.ConfirmAppUUID(common.UUID())
	if e1 != nil {
		logger.Fatal("error persist local instance uuid:", e1)
	}
	app.UUID = uuid
	logger.Info("instance start with uuid:", app.UUID)

	go libcommon.ExpirationDetection()
	startHttpService()
	startTrackerService()
}

// tracker server start tcp listen
func startTrackerService() {
	server := bridgev2.NewServer("", app.PORT)
	server.Listen(libcommon.FutureExpireStorageServer)
}


// start http download server.
func startHttpService() {
	if !app.HTTP_ENABLE {
		logger.Info("http server disabled")
		return
	}

	http.HandleFunc("/nginx", ConfigureNginxHandler)
	http.HandleFunc("/servers", GetAllStorageServers)

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