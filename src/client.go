package main

import (
	"app"
	"container/list"
	"errors"
	"flag"
	"fmt"
	"io"
	"libclient"
	"libcommon"
	"libcommon/bridge"
	"libcommon/bridgev2"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"util/file"
	"util/logger"
	"util/timeutil"
	"validate"
)

var client *libclient.Client
var trackerList *list.List
var checkChan chan int

// 对于客户端，只提供类似于mysql的客户端，每个client与所有的tracker建立单个连接进行数据同步
// client和每个storage server最多建立一个连接
// 三方客户端可以开发成为一个连接池
// echo \"$(ls -m)\" |xargs client -u
// TODO support custom download path in command line.
// path structure:
// /usr/local/godfs
//              |- /bin/client
//              |- /conf/client.conf
// /usr/bin/client -> /usr/local/godfs/bin/client
func main() {

	checkChan = make(chan int)
	abs, _ := filepath.Abs(os.Args[0])
	s, _ := filepath.Split(abs)
	s = file.FixPath(s) // client executor parent path

	// set client type
	app.CLIENT_TYPE = 2
	app.RUN_WITH = 3
	app.UUID = "NATIVE-CLIENT"

	// config file path
	config := prepareClient()
	logger.SetEnable(config.LogEnable)
	validate.SetSystemLogLevel(config.LogLevel)

	if *uploadFile != "" || *downFile != "" {
		client = InitClient()
	}

}
