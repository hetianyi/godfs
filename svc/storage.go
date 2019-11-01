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

// BootStorageServer starts storage server.
func BootStorageServer() {

	if err := util.ValidateStorageConfig(common.InitializedStorageConfiguration); err != nil {
		fmt.Println("Err:", err)
		os.Exit(1)
	}

	if err := util.PrepareDirs(); err != nil {
		logger.Fatal("cannot create tmp dir: ", err)
	}

	//
	if true {
		cbs, _ := json.MarshalIndent(common.InitializedStorageConfiguration, "", "  ")
		logger.Debug("\n", string(cbs))
	}

	// initialize dataset.
	initDataSet()

	// print godfs logo.
	util.PrintLogo()

	writableBinlogManager = binlog.NewXBinlogManager(binlog.LOCAL_BINLOG_MANAGER)
	if common.InitializedStorageConfiguration.EnableHttp {
		StartStorageHttpServer(common.InitializedStorageConfiguration)
	}
	// start member binlog synchronizer.
	InitStorageMemberBinlogWatcher()
	// start tcp server.
	StartStorageTcpServer()
}
