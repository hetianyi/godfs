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
	"os"
	"os/user"
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
// echo \"$(ls -m)\" |xargs /e/godfs-storage/client/bin/go_build_client_go -u
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

	// the file to be upload
	var setConfig = flag.String("set", "", "set global configuration with pattern \"name=value\".")
	// the file to be upload
	var uploadFile = flag.String("u", "", "the file to be upload,\nif you want upload many file once,\nquote file paths using \"\"\" and split with \",\""+
		"\nexample:\nclient -u \"/home/foo/bar1.tar.gz, /home/foo/bar1.tar.gz\"")
	// the file to download
	var downFile = flag.String("d", "", "the file to be download")
	// the download file name
	var customDownloadFileName = flag.String("n", "", "custom download file name")
	// custom override log level
	var logLevel = flag.String("l", "", "custom logging level: trace, debug, info, warning, error, and fatal")
	// custom upload group
	var customGroup = flag.String("g", "", "custom upload group, use with command parameter '-u'")
	// config file path
	m := prepare()
	// whether check file md5 before upload
	var skipCheck = false //*flag.Bool("skip-check", true, "whether check file md5 before upload, true|false")
	flag.Parse()

	if *setConfig != "" {
		ec := resetConfig(*setConfig, m)
		if ec != nil {
			logger.Fatal("error set config:", ec)
		} else {
			logger.Info("success")
		}
		return
	}

	*logLevel = strings.ToLower(strings.TrimSpace(*logLevel))
	if *logLevel != "trace" && *logLevel != "debug" && *logLevel != "info" && *logLevel != "warning" && *logLevel != "error" && *logLevel != "fatal" {
		*logLevel = ""
	}

	if *logLevel != "" {
		if *logLevel != "trace" && *logLevel != "debug" && *logLevel != "info" && *logLevel != "warn" &&
			*logLevel != "error" && *logLevel != "fatal" {
			*logLevel = "info"
		}
		m["log_level"] = *logLevel
	}
	validate.SetSystemLogLevel(m["log_level"])
	logger.SetEnable(app.LOG_ENABLE)

	if *uploadFile != "" || *downFile != "" {
		client = Init()
	}
	if *uploadFile != "" {
		upload(*uploadFile, *customGroup, skipCheck)
		return
	}
	if *downFile != "" {
		download(*downFile, strings.TrimSpace(*customDownloadFileName))
		return
	}
	if *uploadFile == "" && *downFile == "" {
		fmt.Println("godfs client usage:")
		fmt.Println("\t-u string \n\t    the file to be upload,\n\t    if you want upload many file once,\n\t    quote file paths using \"\"\" and split with \",\"" +
			"\n\t    example:\n\t\tclient -u \"/home/foo/bar1.tar.gz, /home/foo/bar1.tar.gz\"")
		fmt.Println("\t-d string \n\t    the file to be download")
		fmt.Println("\t-l string \n\t    custom logging level: trace, debug, info, warning, error, and fatal")
		fmt.Println("\t-n string \n\t    custom download file name")
		fmt.Println("\t-g string \n\t    custom upload group, use with command parameter '-u'")
		fmt.Println("\t--set string \n\t    set client config, for example: \n\t" +
			"    client --set \"tracker=127.0.0.1:1022\"\n\t" +
			"    client --set \"log_level=info\"")
	}
}

// upload files
// paths: file path to be upload
// group: file upload group, if not set, use random group
// skipCheck: whether check md5 before upload
func upload(paths string, group string, skipCheck bool) error {
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
	e := client.DownloadFile(path, 0, -1, func(realPath string, fileLen uint64, reader io.Reader) error {
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
	})
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

func Init() *libclient.Client {
	client := libclient.NewClient(10)
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
	task := &bridge.Task{
		TaskType: app.TASK_SYNC_ALL_STORAGES,
		Callback: func(task *bridge.Task, e error) {
			checkChan <- 1
		},
	}
	libclient.AddTask(task, tracker)
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
				return errors.New("error write out")
			}
			finish += int64(len)
			readBodySize += int64(len)
			logger.Trace("write:", readBodySize)
		} else {
			if e2 == io.EOF {
				return nil
			}
			return e2
		}
	}
	return nil
}

func prepare() map[string]string {
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
		if e2 := writeConf("", "", "true", "info", "d"); e2 != nil {
			logger.Fatal("error write config file:", e2)
		}
	}
	logFilePath := basdir + string(os.PathSeparator) + "logs"
	if !file.Exists(logFilePath) {
		if e1 := file.CreateDir(logFilePath); e != nil {
			logger.Fatal("cannot create directory:", e1)
		}
	}
	m, e3 := file.ReadPropFile(confFilePath)
	if e3 != nil {
		logger.Fatal("error read config file:", e3)
	}

	m["base_path"] = confFilePath
	app.TRACKERS = m["trackers"]
	m["secret"] = strings.TrimSpace(m["secret"])
	app.SECRET = m["secret"]

	//enable log config
	logEnable := strings.ToLower(strings.TrimSpace(m["log_enable"]))
	if logEnable == "" || (logEnable != "true" && logEnable != "false") {
		logEnable = "true"
	}
	if logEnable == "true" {
		app.LOG_ENABLE = true
	} else {
		app.LOG_ENABLE = false
	}
	m["log_enable"] = logEnable

	// check log_level
	logLevel := strings.ToLower(strings.TrimSpace(m["log_level"]))
	if logLevel != "trace" && logLevel != "debug" && logLevel != "info" && logLevel != "warn" &&
		logLevel != "error" && logLevel != "fatal" {
		logLevel = "info"
	}
	m["log_level"] = logLevel
	// check log_rotation_interval
	log_rotation_interval := strings.ToLower(strings.TrimSpace(m["log_rotation_interval"]))
	if log_rotation_interval != "h" && log_rotation_interval != "d" &&
		log_rotation_interval != "m" && log_rotation_interval != "y" {
		log_rotation_interval = "d"
	}
	m["log_rotation_interval"] = log_rotation_interval
	app.LOG_INTERVAL = log_rotation_interval
	return m
}

func resetConfig(setConfig string, m map[string]string) error {
	kv := make([]string, 2)
	firstEQ := strings.Index(setConfig, "=")
	if firstEQ == -1 {
		kv[0] = setConfig
		kv[1] = ""
	} else {
		kv[0] = setConfig[0:firstEQ]
		kv[1] = setConfig[firstEQ+1:]
	}
	var k, v string
	if len(kv) == 1 {
		k = strings.TrimSpace(kv[0])
	} else if len(kv) > 1 {
		k = strings.TrimSpace(kv[0])
		for i := range kv {
			if i == 0 {
				continue
			}
			v += strings.TrimSpace(kv[i])
		}
	}

	if k == "secret" {
		m[k] = v
	} else if k == "trackers" {
		m[k] = v
	} else if k == "log_enable" {
		if v == "" || (v != "true" && v != "false") {
			v = "true"
		}
		m[k] = v
	} else if k == "log_level" {
		if v != "trace" && v != "debug" && v != "info" && v != "warn" &&
			v != "error" && v != "fatal" {
			v = "info"
		}
		m[k] = v
	} else if k == "log_rotation_interval" {
		if v != "h" && v != "d" &&
			v != "m" && v != "y" {
			v = "d"
		}
		m[k] = v
	} else {
		return errors.New("unknown parameter: \"" + k + "\"")
	}
	logger.Info("set", k, "to", "\""+v+"\"")
	return writeConf(m["trackers"], m["secret"], m["log_enable"], m["log_level"], m["log_rotation_interval"])
}

func writeConf(trackers string, secret string, log_enable string, log_level string, log_rotation_interval string) error {
	fi, e := file.CreateFile(app.BASE_PATH + string(os.PathSeparator) + "client.conf")
	if e != nil {
		return e
	}
	fi.WriteString("trackers=" + trackers)
	fi.WriteString("\n")
	fi.WriteString("secret=" + secret)
	fi.WriteString("\n")
	fi.WriteString("log_enable=" + log_enable)
	fi.WriteString("\n")
	fi.WriteString("log_level=" + log_level)
	fi.WriteString("\n")
	fi.WriteString("log_rotation_interval=" + log_rotation_interval)
	fi.WriteString("\n")
	fi.Close()
	return nil
}
