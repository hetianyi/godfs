package svc

import (
	"encoding/json"
	"fmt"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox/logger"
	"os"
)

// godfs --max-logfile-size=128 --log-rotation-interval=m --log-level=debug storage  -g G01 --preferred-network "VMware Network Adapter VMnet1" --disable-logfile --logdir="D:/logs" --data-dir="E:/data1" --http-port=8999 --http-auth="admin23:123456" --enable-mimetypes --secret="kasd3123" --port 3456 --advertise-port=6543 --allowed-domains="baidu.com,google.com" --trackers=123466@localhost:7899,abcd1234@127.0.0.1:9998
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
