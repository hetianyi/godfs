package main

import (
	"app"
	"container/list"
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"libclient"
	"libcommon"
	"libcommon/bridgev2"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
	"util/file"
	"util/logger"
	"validate"
)

var checkChan chan int
var client *libclient.Client
var trackerList *list.List
var command = libclient.COMMAND_NONE

func main() {
	checkChan = make(chan int)
	abs, _ := filepath.Abs(os.Args[0])
	s, _ := filepath.Split(abs)
	s = file.FixPath(s) // client executor parent path

	// set client type
	app.CLIENT_TYPE = 2
	app.RUN_WITH = 3
	app.UUID = "NATIVE-CLIENT"

	initClientFlags()

	// config file path
	prepareClient()
	// check flag vars
	flagVarPreCheck()

	if command == libclient.COMMAND_UPLOAD ||
		command == libclient.COMMAND_DOWNLOAD ||
		command == libclient.COMMAND_INSPECT_FILE{
		client = InitClient()
	}
	libclient.ExecuteCommand(client, command)
}

func flagVarPreCheck() {
	if command == libclient.COMMAND_UPLOAD {
		if libclient.UploadFileList.Len() == 0 {
			logger.Info("no file to be upload")
			os.Exit(0)
		}
	} else if command == libclient.COMMAND_DOWNLOAD {
		if libclient.DownloadFilePath == "" {
			logger.Fatal("no download filepath specified")
			os.Exit(110)
		}
	} else if command == libclient.COMMAND_INSPECT_FILE {
		if libclient.InspectFileList.Len() == 0 {
			logger.Fatal("no md5 specified to inspect")
			os.Exit(111)
		}
	} else if command == libclient.COMMAND_UPDATE_CONFIG {
		if libclient.UpdateConfigList.Len() == 0 {
			logger.Fatal("no config provide for update")
			os.Exit(112)
		}
	}
}


func initClientFlags() {

	appFlag := cli.NewApp()
	appFlag.Version = app.APP_VERSION
	appFlag.Name = "godfs client cli"
	appFlag.Usage = ""

	// config file location
	appFlag.Flags = []cli.Flag {
		cli.StringFlag{
			Name:        "trackers",
			Value:       "127.0.0.1:1022",
			Usage:       "tracker servers",
			Destination: &libclient.Trackers,
		},
		cli.StringFlag{
			Name:        "log_level",
			Value:       "info",
			Usage:       "log level (trace, debug, info, warm, error, fatal)",
			Destination: &libclient.LogLevel,
		},
		cli.StringFlag{
			Name:        "log_rotation_interval",
			Value:       "d",
			Usage:       "log rotation interval h(hour),d(day),m(month),y(year)",
			Destination: &libclient.LogRotationInterval,
		},
		/*cli.BoolTFlag{
			Name:        "log_enable, le",
			Usage:       "whether enable log `LOG_ENABLE` (true, false)",
			Destination: &libclient.LogEnable,

		},*/
		cli.StringFlag{
			Name:        "secret",
			Value:       "",
			Usage:       "secret of trackers (trace, debug, info, warm, error, fatal)",
			Destination: &libclient.Secret,
		},
	}

	// sub command 'upload'
	appFlag.Commands = []cli.Command{
		{
			Name:    "upload",
			Usage:   "upload local files",
			Action:  func(c *cli.Context) error {
				command = libclient.COMMAND_UPLOAD
				abs, _ := filepath.Abs(os.Args[0])
				dir := abs + string(os.PathSeparator) + ".."
				absPath, _ := filepath.Abs(dir)
				if c.Args().First() == "*" {
					fmt.Println("upload all files of", abs)
					files, _ := ioutil.ReadDir(dir)
					for i := range files {
						if !files[i].IsDir() {
							logger.Debug("adding file:", files[i].Name())
							libclient.UploadFileList.PushBack(absPath + string(os.PathSeparator) + files[i].Name())
						}
					}
				} else {
					for i := range c.Args() {
						f := c.Args().Get(i)
						logger.Debug("adding file:", f)
						if !file.IsAbsPath(f) {
							f, _ = filepath.Abs(absPath + string(os.PathSeparator) + f)
						}
						if file.Exists(f) && file.IsFile(f) {
							libclient.UploadFileList.PushBack(f)
						} else {
							logger.Warn("file", f, "not exists or not a file, skip.")
						}
					}
				}
				return nil
			},
			Flags: []cli.Flag {
				cli.StringFlag{
					Name:        "group, g",
					Value:       "",
					Usage:       "group to be upload",
					Destination: &libclient.Group,
				},
			},
		},
		{
			Name:    "download",
			Usage:   "download a file",
			Action:  func(c *cli.Context) error {
				command = libclient.COMMAND_DOWNLOAD
				fmt.Println("download file is: ", c.Args().First())
				libclient.DownloadFilePath = c.Args().First()
				return nil
			},
			Flags: []cli.Flag {
				cli.StringFlag{
					Name:        "name, n",
					Value:       "",
					Usage:       "filename for download file, if not specified, use md5 as filename",
					Destination: &libclient.CustomFileName,
				},
			},
		},
		{
			Name:    "inspect",
			Usage:   "inspect files information by md5",
			Action:  func(c *cli.Context) error {
				command = libclient.COMMAND_INSPECT_FILE
				for i := range c.Args() {
					libclient.InspectFileList.PushBack(c.Args().Get(i))
				}
				return nil
			},
		},
		{ // this sub command is only used by client cli
			Name:    "config",
			Usage:   "client cli configuration settings operation",
			Action:  func(c *cli.Context) error {
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:  "set",
					Usage: "set client cli configuration in 'key=value' form (available keys: trackers, log_enable, log_level, log_rotation_interval, secret)",
					Action: func(c *cli.Context) error {
						command = libclient.COMMAND_UPDATE_CONFIG
						for i := range c.Args() {
							libclient.UpdateConfigList.PushBack(c.Args().Get(i))
						}
						return nil
					},
				},
				{
					Name:  "ls",
					Usage: "list client cli configurations",
					Action: func(c *cli.Context) error {
						command = libclient.COMMAND_LIST_CONFIG
						return nil
					},
				},
			},
		},
	}
	// 帮助文件模板
	cli.AppHelpTemplate = `NAME:
   {{.Name}}{{if .Usage}} - {{.Usage}}{{end}}
USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}
VERSION:
   {{.Version}}{{end}}{{end}}{{if .Description}}
DESCRIPTION:
   {{.Description}}{{end}}{{if len .Authors}}
AUTHOR{{with $length := len .Authors}}{{if ne 1 $length}}S{{end}}{{end}}:
   {{range $index, $author := .Authors}}{{if $index}}
   {{end}}{{$author}}{{end}}{{end}}{{if .VisibleCommands}}
COMMANDS:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{end}}{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
GLOBAL OPTIONS:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}{{if .Copyright}}
COPYRIGHT:
   {{.Copyright}}{{end}}
`

	err := appFlag.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}

}



func prepareClient() *app.ClientConfig {
	user, e := user.Current()
	if e != nil {
		fmt.Println("cannot get system user:", e)
		os.Exit(100)
	}
	basDir := user.HomeDir + string(os.PathSeparator) + ".godfs"
	app.BASE_PATH = basDir
	if !file.Exists(basDir) {
		if e1 := file.CreateDir(basDir); e != nil {
			fmt.Println("cannot create directory:", e1)
			os.Exit(101)
		}
	}
	confFilePath := basDir + string(os.PathSeparator) + "config.json"
	if !file.Exists(confFilePath) {
		config := &app.ClientConfig{
			Trackers: []string{"127.0.0.1:1022"},
			LogLevel: "info",
			LogRotationInterval: "m",
			Secret: "",
		}
		if e2 := libclient.WriteConf(config); e2 != nil {
			fmt.Println("error write config file:", e2)
			os.Exit(102)
		}
	}

	logFilePath := basDir + string(os.PathSeparator) + "logs"
	if !file.Exists(logFilePath) {
		if e1 := file.CreateDir(logFilePath); e != nil {
			fmt.Println("cannot create directory:", e1)
			os.Exit(103)
		}
	}

	config, e3 := libclient.ReadConf()
	if e != nil {
		fmt.Println("error read config file:", e3.Error())
		os.Exit(104)
	}

	app.BASE_PATH = basDir
	app.TRACKERS = strings.Join(config.Trackers, ",")
	if libclient.Trackers != "" {
		app.TRACKERS = libclient.Trackers
		config.Trackers = strings.Split(app.TRACKERS, ",")
	}

	app.SECRET = strings.TrimSpace(config.Secret)
	if libclient.Secret != "" {
		app.SECRET = libclient.Secret
		config.Secret = app.SECRET
	}

	// check log_rotation_interval
	logRotationInterval := strings.ToLower(strings.TrimSpace(config.LogRotationInterval))
	if app.LOG_ROTATION_SETS[logRotationInterval] == 0 {
		logRotationInterval = "d"
	}
	config.LogRotationInterval = logRotationInterval

	if libclient.LogRotationInterval != "" {
		if app.LOG_ROTATION_SETS[libclient.LogRotationInterval] == 0 {
			libclient.LogRotationInterval = "d"
			config.LogRotationInterval = libclient.LogRotationInterval
		}
	}

	app.LOG_INTERVAL = config.LogRotationInterval

	// enable log config
	app.LOG_ENABLE = libclient.LogEnable
	logger.SetEnable(app.LOG_ENABLE)

	// check log_level
	logLevel := strings.ToLower(strings.TrimSpace(config.LogLevel))
	if app.LOG_LEVEL_SETS[logLevel] == 0 {
		logLevel = "info"
	}
	config.LogLevel = logLevel

	if libclient.LogLevel != "" {
		if app.LOG_LEVEL_SETS[libclient.LogLevel] == 0 {
			libclient.LogLevel = "info"
			config.LogLevel = libclient.LogLevel
		}
	}
	validate.SetSystemLogLevel(config.LogLevel)
	return config
}


func InitClient() *libclient.Client {
	client := libclient.NewClient(50)
	collector := libclient.TaskCollector{
		Interval:   time.Millisecond * 30,
		FirstDelay: 0,
		ExecTimes:  1,
		Name:       "::: synchronize storage server instances :::",
		Job:        clientMonitorCollector,
	}
	collectors := new(list.List)
	collectors.PushBack(&collector)
	maintainer := &libclient.TrackerMaintainer{Collectors: *collectors}
	client.TrackerMaintainer = maintainer
	maintainer.DieCallback = func(tracker string) {
		logger.Debug("finish a tracker:", tracker)
		checkChan <- 1
	}

	trackerList := libcommon.ParseTrackers(app.TRACKERS)
	trackerMap := make(map[string]string)
	if trackerList != nil {
		for ele := trackerList.Front(); ele != nil; ele = ele.Next() {
			trackerMap[ele.Value.(string)] = app.SECRET
		}
	}
	maintainer.Maintain(trackerMap)
	logger.Info("synchronizing with trackers...")
	for i := 0; i < trackerList.Len(); i++ {
		<-checkChan
	}
	logger.Info("finish synchronizing with all trackers")

	// check storage members
	if libclient.GroupMembers.Len() == 0 &&
		(command == libclient.COMMAND_DOWNLOAD || command == libclient.COMMAND_UPLOAD) {
		logger.Fatal("cannot upload or download file, no storage server available")
	}
	return client
}


func clientMonitorCollector(tracker *libclient.TrackerInstance) {
	logger.Debug("create sync task for tracker:", tracker.ConnStr)
	task := &bridgev2.Task{
		TaskType: app.TASK_SYNC_ALL_STORAGES,
	}
	libclient.AddTask(task, tracker)
}