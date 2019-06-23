package svc

import (
	"encoding/json"
	"fmt"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox/logger"
	"os"
)

func BootStorageServer() {
	if err := util.ValidateStorageConfig(common.Config); err != nil {
		fmt.Println("Err:", err)
		os.Exit(1)
	}
	if err := util.PrepareDirs(); err != nil {
		logger.Fatal("cannot create tmp dir: ", err)
	}
	common.Config.InstanceId = util.LoadInstanceData()
	util.PrintLogo()
	cbs, _ := json.MarshalIndent(common.Config, "", "  ")
	fmt.Println("boot storage server success!")
	fmt.Println(string(cbs))
	if common.Config.EnableHttp {
		StartStorageHttpServer(common.Config)
	}
	StartStorageTcpServer()
}
