package main

import (
	"app"
	"container/list"
	"errors"
	json "github.com/json-iterator/go"
	"fmt"
	"github.com/urfave/cli"
	"io"
	"libclient"
	"libcommon"
	"libcommon/bridge"
	"libcommon/bridgev2"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"util/file"
	"util/logger"
	"util/timeutil"
)

var (
	ConfigFile string
	Trackers string
	LogLevel string
	LogRotationInterval string
	LogEnable = true
	Secret string
	UploadFileList list.List
	Group string
	CustomFileName string
	SetConfig string
)


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



}


func initClientFlags() {

	appFlag := cli.NewApp()
	appFlag.Version = app.APP_VERSION
	appFlag.Name = "godfs client cli"
	appFlag.Usage = ""

	// config file location
	appFlag.Flags = []cli.Flag {
		cli.StringFlag{
			Name:        "trackers, t",
			Value:       "127.0.0.1:1022",
			Usage:       "tracker servers `TRACKERS`",
			Destination: &Trackers,
		},
		cli.StringFlag{
			Name:        "log_level, l",
			Value:       "info",
			Usage:       "log level `LOG_LEVEL` (trace, debug, info, warm, error, fatal)",
			Destination: &LogLevel,
		},
		cli.StringFlag{
			Name:        "log_rotation_interval, li",
			Value:       "d",
			Usage:       "log rotation interval `LOG_ROTATION_INTERVAL` h(hour),d(day),m(month),y(year)",
			Destination: &LogRotationInterval,
		},
		cli.BoolFlag{
			Name:        "log_enable, le",
			Usage:       "whether enable log `LOG_ENABLE` (true, false)",
			Destination: &LogEnable,
		},
		cli.StringFlag{
			Name:        "secret, s",
			Value:       "",
			Usage:       "secret of trackers `LOG_LEVEL` (trace, debug, info, warm, error, fatal)",
			Destination: &Secret,
		},
	}

	// sub command 'upload'
	appFlag.Commands = []cli.Command{
		{
			Name:    "upload",
			Usage:   "upload local files",
			Action:  func(c *cli.Context) error {
				if c.Args().First() == "*" {
					abs, _ := filepath.Abs(os.Args[0])
					fmt.Println("upload all files of", abs)
				} else {
					var files = new(list.List)
					for i := range c.Args() {
						f := c.Args().Get(i)
						logger.Debug("checking file:", f)
						if file.Exists(f) && file.IsFile(f) {
							files.PushBack(f)
						} else {
							logger.Warn("file", f, "not exists or not a file, skip.")
						}
					}
					uploadFiles := strings.Split(paths, ",")
					var pickList list.List
					for i := range uploadFiles {
						uploadFiles[i] = strings.TrimSpace(uploadFiles[i])
						if file.Exists(uploadFiles[i]) && file.IsFile(uploadFiles[i]) {
							pickList.PushBack(uploadFiles[i])
						} else {
							logger.Warn("file", uploadFiles[i], "not exists or not a file, skip.")
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
					Destination: &Group,
				},
			},
		},
		{
			Name:    "download",
			Usage:   "download a file",
			Action:  func(c *cli.Context) error {
				fmt.Println("download file is: ", c.Args().First())
				return nil
			},
			Flags: []cli.Flag {
				cli.StringFlag{
					Name:        "name, n",
					Value:       "",
					Usage:       "filename for download file, if not specified, use md5 as filename",
					Destination: &CustomFileName,
				},
			},
		},
		{
			Name:    "inspect",
			Usage:   "inspect files information by md5",
			Action:  func(c *cli.Context) error {
				fmt.Println("get files is: ")
				for i := range c.Args() {
					fmt.Println(c.Args().Get(i))
				}
				return nil
			},
		},
		{ // this sub command is only used by client cli
			Name:    "config",
			Usage:   "client cli configuration settings operation",
			Action:  func(c *cli.Context) error {
				fmt.Println("get files is: ")
				for i := range c.Args() {
					fmt.Println(c.Args().Get(i))
				}
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:  "set",
					Usage: "set client cli configuration in 'key=value' form (available keys: log, trackers, secret, log_enable, log_rotation_interval)",
					Action: func(c *cli.Context) error {
						fmt.Println("upload file is: ", c.Args().First())
						return nil
					},
				},
				{
					Name:  "list",
					Usage: "list client cli configurations",
					Action: func(c *cli.Context) error {
						return nil
					},
				},
			},
		},
	}

	err := appFlag.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}

}



func prepareClient() *app.ClientConfig {
	user, e := user.Current()
	if e != nil {
		logger.Fatal(e)
	}
	basdir := user.HomeDir + string(os.PathSeparator) + ".godfs"
	app.BASE_PATH = basdir
	if !file.Exists(basdir) {
		if e1 := file.CreateDir(basdir); e != nil {
			logger.Fatal("cannot create directory:", e1)
		}
	}
	confFilePath := basdir + string(os.PathSeparator) + "client.conf"
	if !file.Exists(confFilePath) {
		config := &app.ClientConfig{
			Trackers: []string{"127.0.0.1:1022"},
			LogEnable: true,
			LogLevel: "info",
			LogRotationInterval: "m",
			Secret: "",
		}
		if e2 := writeConf(config); e2 != nil {
			logger.Fatal("error write config file:", e2)
		}
	}

	logFilePath := basdir + string(os.PathSeparator) + "logs"
	if !file.Exists(logFilePath) {
		if e1 := file.CreateDir(logFilePath); e != nil {
			logger.Fatal("cannot create directory:", e1)
		}
	}

	config, e3 := readConf()
	if e != nil {
		logger.Fatal("error read config file:", e3.Error())
	}

	app.BASE_PATH = basdir
	app.TRACKERS = strings.Join(config.Trackers, ",")
	app.SECRET = strings.TrimSpace(config.Secret)

	//enable log config
	app.LOG_ENABLE = config.LogEnable

	// check log_level
	logLevel := strings.ToLower(strings.TrimSpace(config.LogLevel))
	if logLevel != "trace" && logLevel != "debug" && logLevel != "info" && logLevel != "warn" &&
		logLevel != "error" && logLevel != "fatal" {
		logLevel = "info"
	}
	config.LogLevel = logLevel

	// check log_rotation_interval
	logRotationInterval := strings.ToLower(strings.TrimSpace(config.LogRotationInterval))
	if logRotationInterval != "h" && logRotationInterval != "d" &&
		logRotationInterval != "m" && logRotationInterval != "y" {
		logRotationInterval = "d"
	}
	config.LogRotationInterval = logRotationInterval
	app.LOG_INTERVAL = config.LogRotationInterval

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

	trackerList := libcommon.ParseTrackers(app.TRACKERS)
	trackerMap := make(map[string]string)
	if trackerList != nil {
		for ele := trackerList.Front(); ele != nil; ele = ele.Next() {
			trackerMap[ele.Value.(string)] = app.SECRET
		}
	}
	maintainer.Maintain(trackerMap)
	logger.Info("synchronize with trackers...")
	for i := 0; i < trackerList.Len(); i++ {
		<-checkChan
	}
	return client
}

func clientMonitorCollector(tracker *libclient.TrackerInstance) {
	logger.Debug("create sync task for tracker:", tracker.ConnStr)
	task := &bridgev2.Task{
		TaskType: app.TASK_SYNC_ALL_STORAGES,
		Callback: func(task *bridgev2.Task, e error) {
			logger.Debug("finish a tracker:", tracker.ConnStr)
			checkChan <- 1
		},
	}
	libclient.AddTask(task, tracker)
}


// write client config to file.
func writeConf(clientConfig *app.ClientConfig) error {
	fi, e := file.CreateFile(app.BASE_PATH + string(os.PathSeparator) + "client.conf")
	if e != nil {
		return e
	}
	defer fi.Close()
	bs, e1 := json.MarshalIndent(clientConfig, "", "  ")
	if e1 != nil {
		return e1
	}
	fi.Write(bs)
	return nil
}

// read client config
func readConf() (*app.ClientConfig, error) {
	configFile, e := file.GetFile(app.BASE_PATH + string(os.PathSeparator) + "client.conf")
	if e != nil {
		return nil, e
	}
	defer configFile.Close()

	fi, e1 := configFile.Stat()
	if e1 != nil {
		return nil, e1
	}

	buffer, e2 := bridgev2.MakeBytes(fi.Size(), true, 10240, true)
	if e2 != nil {
		return nil, e2
	}
	_, e3 := configFile.Read(buffer)
	if e3 != nil {
		return nil, e3
	}
	var config = &app.ClientConfig{}
	e4 := json.Unmarshal(buffer, config)
	if e4 != nil {
		return nil, e4
	}
	return config, nil
}



// upload files
// paths: file path to be upload
// group: file upload group, if not set, use random group
// skipCheck: whether check md5 before upload
func upload(paths []string, group string, skipCheck bool) error {
	uploadFiles := strings.Split(paths, ",")
	var pickList list.List
	for i := range uploadFiles {
		uploadFiles[i] = strings.TrimSpace(uploadFiles[i])
		if file.Exists(uploadFiles[i]) && file.IsFile(uploadFiles[i]) {
			pickList.PushBack(uploadFiles[i])
		} else {
			logger.Warn("file", uploadFiles[i], "not exists or not a file, skip.")
		}
	}
	for ele := pickList.Front(); ele != nil; ele = ele.Next() {
		var startTime = time.Now()
		fid, e := client.Upload(ele.Value.(string), group, startTime, skipCheck)
		if e != nil {
			logger.Error(e)
		} else {
			now := time.Now()
			fmt.Println("[==========] 100% [" + timeutil.GetHumanReadableDuration(startTime, now) + "]\nupload success, file id:")
			fmt.Println("+-------------------------------------------+")
			fmt.Println(fid)
			fmt.Println("+-------------------------------------------+")
		}
	}
	return nil
}

func download(path string, customDownloadFileName string) error {
	filePath := ""
	var startTime time.Time
	e := client.DownloadFile(path, 0, -1, func(manager *bridgev2.ConnectionManager, frame *bridgev2.Frame, resMeta *bridgev2.DownloadFileResponseMeta) (b bool, e error) {
		path = strings.TrimSpace(path)
		if strings.Index(path, "/") != 0 {
			path = "/" + path
		}
		var fi *os.File
		if customDownloadFileName == "" {
			md5 := regexp.MustCompile(app.PATH_REGEX).ReplaceAllString(path, "${4}")
			customDownloadFileName = md5
			f, e1 := file.CreateFile(customDownloadFileName)
			if e1 != nil {
				return true, e1
			}
			fi = f
		} else {
			f, e1 := file.CreateFile(customDownloadFileName)
			if e1 != nil {
				return true, e1
			}
			fi = f
		}
		defer fi.Close()
		filePath, _ = filepath.Abs(fi.Name())
		startTime = time.Now()
		return true, writeOut(manager.Conn, frame.BodyLength, fi, startTime)
	})
	/*e := client.DownloadFile(path, 0, -1, func(realPath string, fileLen uint64, reader io.Reader) error {
		var fi *os.File
		if customDownloadFileName == "" {
			md5 := regexp.MustCompile(app.PATH_REGEX).ReplaceAllString(realPath, "${4}")
			customDownloadFileName = md5
			f, e1 := file.CreateFile(customDownloadFileName)
			if e1 != nil {
				return e1
			}
			fi = f
		} else {
			f, e1 := file.CreateFile(customDownloadFileName)
			if e1 != nil {
				return e1
			}
			fi = f
		}
		defer fi.Close()
		filePath, _ = filepath.Abs(fi.Name())
		startTime = time.Now()
		return writeOut(reader, int64(fileLen), fi, startTime)
	})*/
	if e != nil {
		logger.Error("download failed:", e)
		return e
	} else {
		now := time.Now()
		fmt.Println("[==========] 100% [" + timeutil.GetHumanReadableDuration(startTime, now) + "]\ndownload success, file save as:")
		fmt.Println("+-------------------------------------------+")
		fmt.Println(filePath)
		fmt.Println("+-------------------------------------------+")
	}
	return nil
}


func writeOut(in io.Reader, offset int64, out io.Writer, startTime time.Time) error {
	buffer, _ := bridge.MakeBytes(app.BUFF_SIZE, false, 0, false)
	defer bridge.RecycleBytes(buffer)
	var finish, total int64
	var stopFlag = false
	defer func() { stopFlag = true }()
	total = offset
	finish = 0
	go libcommon.ShowPercent(&total, &finish, &stopFlag, startTime)

	// total read bytes
	var readBodySize int64 = 0
	// next time bytes to read
	var nextReadSize int
	for {
		// left bytes is more than a buffer
		if (offset-readBodySize)/int64(len(buffer)) >= 1 {
			nextReadSize = len(buffer)
		} else { // left bytes less than a buffer
			nextReadSize = int(offset - readBodySize)
		}
		if nextReadSize == 0 {
			break
		}
		len, e2 := in.Read(buffer[0:nextReadSize])
		if e2 == nil {
			wl, e5 := out.Write(buffer[0:len])
			if e5 != nil || wl != len {
				return errors.New("error write bytes")
			}
			finish += int64(len)
			readBodySize += int64(len)
		} else {
			if e2 == io.EOF {
				return nil
			}
			return e2
		}
	}
	return nil
}


