package svc

import (
	"fmt"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox/logger"
	json "github.com/json-iterator/go"
	"os"
)

// BootAgentServer starts agent server.
func BootAgentServer() {

	if err := util.ValidateAgentConfig(common.InitializedAgentConfiguration); err != nil {
		fmt.Println("Err:", err)
		os.Exit(1)
	}

	if err := util.PrepareDirs(common.InitializedAgentConfiguration.TmpDir); err != nil {
		logger.Fatal("cannot create tmp dir: ", err)
	}

	//
	if true {
		cbs, _ := json.MarshalIndent(common.InitializedAgentConfiguration, "", "  ")
		logger.Debug("\n", string(cbs))
	}

	// print godfs logo.
	util.PrintLogo()

	StartAgentTcpServer()

	StartAgentHttpServer(common.InitializedAgentConfiguration)
}
