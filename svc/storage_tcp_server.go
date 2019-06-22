package svc

import (
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/gpip"
	"github.com/hetianyi/gox/logger"
	"net"
)

func StartStorageTcpServer(c *common.StorageConfig) {
	listener, err := net.Listen("tcp", c.BindAddress+":"+convert.IntToStr(c.Port))
	if err != nil {
		logger.Fatal("server started failed: ", err)
	}
	logger.Info("tcp server started on port ", c.Port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("error accepting new connection: ", err)
			continue
		}
		logger.Debug("accept a new connection")
		gox.Try(func() {
			go serverHandler(conn)
		}, func(i interface{}) {
			logger.Error("connection error:", err)
		})
	}
}

func serverHandler(conn net.Conn) {
	pip := &gpip.Pip{
		Conn: conn,
	}
	defer pip.Close()
}
