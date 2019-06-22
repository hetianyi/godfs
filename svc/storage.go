package svc

import (
	"fmt"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"os"
)

func BootStorageServer() {
	if err := util.ValidateStorageConfig(common.Config); err != nil {
		fmt.Println("Err:", err)
		os.Exit(1)
	}
	util.PrintLogo()
	/*cbs, _ := json.MarshalIndent(c, "", "  ")
	fmt.Println("boot storage server success!")
	fmt.Println(string(cbs))*/
	if common.Config.EnableHttp {
		StartStorageHttpServer(common.Config)
	}
	StartStorageTcpServer()
}
