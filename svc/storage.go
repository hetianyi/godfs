package svc

import (
	"fmt"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"os"
)

func BootStorageServer(c *common.StorageConfig) {
	if err := util.ValidateStorageConfig(c); err != nil {
		fmt.Println("Err:", err)
		os.Exit(1)
	}
	/*cbs, _ := json.MarshalIndent(c, "", "  ")
	fmt.Println("boot storage server success!")
	fmt.Println(string(cbs))*/
	if c.EnableHttp {
		StartStorageHttpServer(c)
	}
	StartStorageTcpServer(c)
}
