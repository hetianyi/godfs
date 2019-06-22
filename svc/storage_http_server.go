package svc

import (
	"github.com/gorilla/mux"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/logger"
	"net/http"
	"time"
)

// StartStorageHttpServer starts an storage http server.
func StartStorageHttpServer(c *common.StorageConfig) {
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
		for {
			logger.Info("http server started on port ", c.HttpPort)
			if err := srv.ListenAndServe(); err != nil {
				logger.Error("cannot start http server: ", err)
				time.Sleep(time.Second * 5)
			}
		}
	}()
}
