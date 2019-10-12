package svc

import (
	"encoding/json"
	"fmt"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/reg"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox/logger"
	"os"
)

func BootTrackerServer() {
	if err := util.ValidateTrackerConfig(common.InitializedTrackerConfiguration); err != nil {
		fmt.Println("Err:", err)
		os.Exit(1)
	}
	common.InitializedTrackerConfiguration.InstanceId = util.LoadInstanceData(common.TRACKER_CONFIG_MAP_KEY)
	if true {
		cbs, _ := json.MarshalIndent(common.InitializedTrackerConfiguration, "", "  ")
		logger.Debug("\n", string(cbs))
	}
	util.PrintLogo()
	if common.InitializedTrackerConfiguration.EnableHttp {
		StartTrackerHttpServer(common.InitializedTrackerConfiguration)
	}
	reg.InitRegistry()
	StartTrackerTcpServer()
}
