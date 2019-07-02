package svc

import (
	"github.com/gorilla/mux"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/logger"
	"net/http"
	"time"
)

// StartTrackerHttpServer starts an storage http server.
func StartTrackerHttpServer(c *common.TrackerConfig) {
	r := mux.NewRouter()
	srv := &http.Server{
		Handler: r,
		Addr:    c.BindAddress + ":" + convert.IntToStr(c.HttpPort),
		// Good practice: enforce timeouts for servers you create!
		ReadHeaderTimeout: time.Second * 15,
		WriteTimeout:      0,
		ReadTimeout:       0,
		MaxHeaderBytes:    1 << 20, // 1MB
	}
	go func() {
		logger.Info("http server listening on ", c.BindAddress, ":", c.HttpPort)
		if err := srv.ListenAndServe(); err != nil {
			logger.Fatal(err)
		}
	}()
}
