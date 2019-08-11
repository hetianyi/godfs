package svc

import (
	"fmt"
	"github.com/hetianyi/godfs/binlog"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox/logger"
	json "github.com/json-iterator/go"
	"os"
)

// godfs --max-logfile-size=128 --log-rotation-interval=m --log-level=debug storage  -g G01 --preferred-network "VMware Network Adapter VMnet1" --disable-logfile --logdir="D:/logs" --data-dir="E:/data1" --http-port=8999 --http-auth="admin23:123456" --enable-mimetypes --secret="kasd3123" --port 3456 --advertise-port=6543 --allowed-domains="baidu.com,google.com" --trackers=123466@localhost:7899,abcd1234@127.0.0.1:9998
func BootStorageServer() {
	if err := util.ValidateStorageConfig(common.InitializedStorageConfiguration); err != nil {
		fmt.Println("Err:", err)
		os.Exit(1)
	}
	if err := util.PrepareDirs(); err != nil {
		logger.Fatal("cannot create tmp dir: ", err)
	}
	common.InitializedStorageConfiguration.InstanceId = util.LoadInstanceData(common.InitializedStorageConfiguration.DataDir)
	if true {
		cbs, _ := json.MarshalIndent(common.InitializedStorageConfiguration, "", "  ")
		logger.Debug("\n", string(cbs))
	}
	util.PrintLogo()
	writableBinlogManager = binlog.NewXBinlogManager(binlog.LOCAL_BINLOG_MANAGER)
	if common.InitializedStorageConfiguration.EnableHttp {
		StartStorageHttpServer(common.InitializedStorageConfiguration)
	}
	StartStorageTcpServer()
}
