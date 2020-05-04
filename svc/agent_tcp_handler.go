package svc

import (
	"github.com/hetianyi/godfs/api"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/logger"
	"github.com/logrusorgru/aurora"
	"net"
	"time"
)

func StartAgentTcpServer() {

	_, err := net.Listen("tcp",
		common.InitializedAgentConfiguration.BindAddress+":"+
			convert.IntToStr(common.InitializedAgentConfiguration.Port))
	if err != nil {
		logger.Fatal(err)
	}

	time.Sleep(time.Millisecond * 50)

	logger.Info(" tcp server listening on ",
		common.InitializedAgentConfiguration.BindAddress, ":",
		common.InitializedAgentConfiguration.Port)
	logger.Info("my instance id: ", common.InitializedAgentConfiguration.InstanceId)
	logger.Info(aurora.BrightGreen("::: agent server started :::"))

	// running in cluster mode.
	if common.InitializedAgentConfiguration.ParsedTrackers != nil &&
		len(common.InitializedAgentConfiguration.ParsedTrackers) > 0 {
		servers := make([]*common.Server, len(common.InitializedAgentConfiguration.ParsedTrackers))
		for i, s := range common.InitializedAgentConfiguration.ParsedTrackers {
			servers[i] = &s
		}
		config := &api.Config{
			MaxConnectionsPerServer: MaxConnPerServer,
			SynchronizeOnce:         false,
			TrackerServers:          servers,
		}
		InitializeClientAPI(config)
	}

	// TODO tcp agent is not ready yet.
	/*for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("error accepting new connection: ", err)
			continue
		}
		logger.Debug("accept a new connection")
		go storageClientConnHandler(conn)
	}*/
}
