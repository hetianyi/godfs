package validate

import (
	"app"
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"util/file"
	"util/logger"
)

var (
	az                   = []rune{'A', 'B', 'C', 'D', 'E', 'F'}
	i09                  = []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
	GroupInstancePattern = "^[0-9a-zA-Z_]{1,10}$" // group name and instance id pattern
)

// check configuration file parameter.
// if check failed, system will shutdown.
// runWith:
//        1: storage server
//        2: tracker server
//        3: client
func Check(m map[string]string, runWith int) {
	replaceParams(m)
	// check: bind_address
	bind_address := strings.TrimSpace(m["bind_address"])
	app.BIND_ADDRESS = bind_address

	// check base_path
	basePath := strings.TrimSpace(m["base_path"])
	if basePath == "" {
		abs, _ := filepath.Abs(os.Args[0])
		parent, _ := filepath.Split(abs)
		finalPath := parent + "godfs"
		logger.Info("base_path not set, use", finalPath)
		m["base_path"] = finalPath
	} else {
		m["base_path"] = file.FixPath(basePath)
	}
	app.BASE_PATH = m["base_path"]
	cleanTmpdir()
	prepareDirs(m["base_path"])

	// check secret
	m["secret"] = strings.TrimSpace(m["secret"])
	app.SECRET = m["secret"]

	// check log_level
	logLevel := strings.ToLower(strings.TrimSpace(m["log_level"]))
	if logLevel != "trace" && logLevel != "debug" && logLevel != "info" && logLevel != "warn" &&
		logLevel != "error" && logLevel != "fatal" {
		logLevel = "info"
	}
	m["log_level"] = logLevel
	SetSystemLogLevel(logLevel)

	// check log_rotation_interval
	log_rotation_interval := strings.ToLower(strings.TrimSpace(m["log_rotation_interval"]))
	if log_rotation_interval != "h" && log_rotation_interval != "d" &&
		log_rotation_interval != "m" && log_rotation_interval != "y" {
		log_rotation_interval = "d"
	}
	m["log_rotation_interval"] = log_rotation_interval
	app.LOG_INTERVAL = log_rotation_interval

	//enable log config
	logEnable := strings.ToLower(strings.TrimSpace(m["log_enable"]))
	if logEnable == "" || (logEnable != "true" && logEnable != "false") {
		logEnable = "true"
	}
	if logEnable == "true" {
		app.LOG_ENABLE = true
		logger.SetEnable(true)
	} else {
		app.LOG_ENABLE = false
		logger.SetEnable(false)
	}

	if runWith == 1 {
		// check GROUP
		m["group"] = strings.TrimSpace(m["group"])
		if mat, _ := regexp.Match(GroupInstancePattern, []byte(m["group"])); !mat {
			logger.Fatal("error parameter 'group'")
		}
		app.GROUP = m["group"]

		// check http auth
		m["http_auth"] = strings.TrimSpace(m["http_auth"])
		app.HTTP_AUTH = m["http_auth"]

		// check instance id
		m["instance_id"] = strings.TrimSpace(m["instance_id"])
		if mat, _ := regexp.Match(GroupInstancePattern, []byte(m["instance_id"])); !mat {
			logger.Fatal("error parameter 'instance_id'")
		}
		app.INSTANCE_ID = m["instance_id"]

		// check http_port
		http_port := strings.ToLower(strings.TrimSpace(m["http_port"]))

		httpPort, ehp := strconv.Atoi(http_port)
		if ehp != nil || httpPort <= 0 || httpPort > 65535 {
			logger.Fatal("error http_port:", http_port)
		}
		m["http_port"] = http_port
		app.HTTP_PORT = httpPort

		// check assign_disk_space
		assign_disk_space := strings.ToLower(strings.TrimSpace(m["assign_disk_space"]))
		value, unit := FixStorageSize(assign_disk_space, "MB")
		if value == "" {
			value = "50"
		}
		if unit == "" {
			unit = "MB"
		}
		_val, e3 := strconv.ParseFloat(value, 64)
		if e3 != nil {
			logger.Fatal("error assign_disk_space:", value+unit)
		}
		var _unit = GetUnitVal(unit)
		app.ASSIGN_DISK_SPACE = int64(_val * float64(_unit))
		m["assign_disk_space"] = value + unit

		// check slice_size
		slice_size := strings.ToLower(strings.TrimSpace(m["slice_size"]))
		value1, unit1 := FixStorageSize(slice_size, "MB")
		if value1 == "" {
			value1 = "50"
		}
		if unit1 == "" {
			unit1 = "MB"
		}
		_val1, e4 := strconv.ParseFloat(value1, 64)
		if e4 != nil {
			logger.Fatal("error slice_size:", value1+unit1)
		}
		var _unit1 = GetUnitVal(unit1)
		app.SLICE_SIZE = int64(_val1 * float64(_unit1))
		m["slice_size"] = value1 + unit1
		logger.Debug("slice_size:", app.SLICE_SIZE)

		// check http_enable
		http_enable := strings.ToLower(strings.TrimSpace(m["http_enable"]))
		if http_enable != "true" && http_enable != "false" {
			http_enable = "false"
		}
		m["http_enable"] = http_enable
		app.HTTP_ENABLE = http_enable == "true"

		// check upload_enable
		upload_enable := strings.ToLower(strings.TrimSpace(m["upload_enable"]))
		if upload_enable != "true" && upload_enable != "false" {
			upload_enable = "true"
		}
		m["upload_enable"] = upload_enable
		app.UPLOAD_ENABLE = upload_enable == "true"

		// check enable_mime_types
		enable_mime_types := strings.ToLower(strings.TrimSpace(m["enable_mime_types"]))
		if enable_mime_types != "true" && enable_mime_types != "false" {
			enable_mime_types = "true"
		}
		m["enable_mime_types"] = enable_mime_types
		app.MIME_TYPES_ENABLE = (enable_mime_types == "true") && app.HTTP_ENABLE
		if app.MIME_TYPES_ENABLE {
			app.SetMimeTypesEnable()
		}

		// check web_content_mime_types
		m["web_content_mime_types"] = strings.TrimSpace(m["web_content_mime_types"])
		wcmt := strings.Split(m["web_content_mime_types"], ",")
		for i := range wcmt {
			strS := strings.TrimSpace(wcmt[i])
			if strS == "" {
				continue
			}
			app.AddWebMimeType(strS)
		}

		//--
	} else if runWith == 2 {

	}

	if runWith == 1 || runWith == 2 {
		// check port
		port, e := strconv.Atoi(m["port"])
		if e == nil {
			if port <= 0 || port > 65535 {
				logger.Fatal("invalid port range:", m["port"])
			}
			app.PORT = port
		} else {
			logger.Fatal("invalid port ", m["port"], ":", e)
		}
	}
	if runWith == 1 || runWith == 3 {

		// check trackers
		trackers := strings.TrimSpace(m["trackers"])
		_ts := strings.Split(trackers, ",")
		var bytebuff bytes.Buffer
		for i := range _ts {
			strS := strings.TrimSpace(_ts[i])
			if strS == "" {
				continue
			}
			bytebuff.WriteString(strS)
			if i < len(_ts)-1 {
				bytebuff.WriteString(",")
			}
		}
		m["trackers"] = string(bytebuff.Bytes())
		app.TRACKERS = m["trackers"]
	}
}

func createDirs(basePath string) {
	dataDir := file.FixPath(basePath + string(os.PathSeparator) + "data")
	logsDir := file.FixPath(basePath + string(os.PathSeparator) + "logs")
	tmpDir := file.FixPath(basePath + string(os.PathSeparator) + "data/tmp")
	if !file.Exists(dataDir) {
		e := file.CreateAllDir(dataDir)
		if e != nil {
			logger.Fatal("cannot create data directory:", dataDir)
		}
	}
	if !file.Exists(logsDir) {
		e := file.CreateAllDir(logsDir)
		if e != nil {
			logger.Fatal("cannot create data directory:", logsDir)
		}
	}
	if !file.Exists(tmpDir) {
		e := file.CreateAllDir(tmpDir)
		if e != nil {
			logger.Fatal("cannot create data directory:", tmpDir)
		}
	}
	if file.Exists(dataDir) && file.IsFile(dataDir) {
		logger.Fatal("cannot create data directory:", dataDir)
	}
	if file.Exists(logsDir) && file.IsFile(logsDir) {
		logger.Fatal("cannot create data directory:", logsDir)
	}
	if file.Exists(tmpDir) && file.IsFile(tmpDir) {
		logger.Fatal("cannot create data directory:", tmpDir)
	}
	//create dir now disabled
	/*
	   crossCreateDir(dataDir, az, az, true)
	   crossCreateDir(dataDir, i09, i09, true)
	   crossCreateDir(dataDir, az, i09, true)
	*/
}

func crossCreateDir(dataDir string, arr1 []rune, arr2 []rune, deeper bool) {
	for i := range arr1 {
		for k := range arr2 {
			d1 := dataDir + string(os.PathSeparator) + string(arr1[i]) + string(arr2[k])
			d2 := dataDir + string(os.PathSeparator) + string(arr2[k]) + string(arr1[i])
			if file.Exists(d1) {
				if file.IsFile(d1) {
					logger.Fatal("error create dir:", d1)
				}
			} else {
				e := file.CreateDir(d1)
				if e != nil {
					logger.Fatal("error create dir:", d1)
				}
				logger.Debug("create data dir:", d1)
			}
			if file.Exists(d2) {
				if file.IsFile(d2) {
					logger.Fatal("error create dir:", d2)
				}
			} else {
				e := file.CreateDir(d2)
				if e != nil {
					logger.Fatal("error create dir:", d2)
					logger.Debug("create data dir:", d2)
				}
				logger.Debug("create data dir:", d2)
			}
			if deeper {
				cd1 := string(arr2[k]) + string(arr1[i])
				crossCreateDir(dataDir+"/"+cd1, az, az, false)
				crossCreateDir(dataDir+"/"+cd1, i09, i09, false)
				crossCreateDir(dataDir+"/"+cd1, az, i09, false)
				cd2 := string(arr1[i]) + string(arr2[k])
				crossCreateDir(dataDir+"/"+cd2, az, az, false)
				crossCreateDir(dataDir+"/"+cd2, i09, i09, false)
				crossCreateDir(dataDir+"/"+cd2, az, i09, false)
			}
		}
	}
}

func FixStorageSize(input string, defaultUnit string) (value string, unit string) {
	input = strings.ToLower(input)
	if mat, e2 := regexp.Match("^([1-9][0-9]*)([kmgtp]?[b]?)$", []byte(input)); mat && e2 == nil {
		value := regexp.MustCompile("^([1-9][0-9]*)([kmgtp]?[b]?)$").ReplaceAllString(input, "${1}")
		unit := regexp.MustCompile("^([1-9][0-9]*)([kmgtp]?[b]?)$").ReplaceAllString(input, "${2}")
		if len(unit) == 0 {
			unit = strings.ToUpper(unit + defaultUnit)
		}
		if len(unit) == 1 {
			unit = strings.ToUpper(unit + "b")
		}
		return value, unit
	}
	return "", ""
}

func GetUnitVal(unit string) int64 {
	var _unit int64
	if unit == "BB" {
		_unit = 1
	} else if unit == "KB" {
		_unit = 1024
	} else if unit == "MB" {
		_unit = 1024 * 1024
	} else if unit == "GB" {
		_unit = 1024 * 1024 * 1024
	} else if unit == "PB" {
		_unit = 1024 * 1024 * 1024 * 1024 * 1024
	} else {
		_unit = 0
	}
	return _unit
}

func SetSystemLogLevel(logLevel string) {
	logger.Debug("log level set to", logLevel)
	if logLevel == "debug" {
		logger.SetLogLevel(1)
	} else if logLevel == "info" {
		logger.SetLogLevel(2)
	} else if logLevel == "warn" {
		logger.SetLogLevel(3)
	} else if logLevel == "error" {
		logger.SetLogLevel(4)
	} else if logLevel == "fatal" {
		logger.SetLogLevel(5)
	} else if logLevel == "trace" {
		logger.SetLogLevel(0)
	}
}

func prepareDirs(finalPath string) {
	// if basepath file exists and it is a file.
	if file.Exists(finalPath) && file.IsFile(finalPath) {
		logger.Fatal("could not create base path:", finalPath)
	}

	if !file.Exists(finalPath) {
		e := file.CreateDir(finalPath)
		if e != nil {
			logger.Fatal("could not create base path:", finalPath)
		}
	}
	createDirs(finalPath)
}

// 每次启动前尝试清理tmp目录
func cleanTmpdir() {
	logger.Debug("clean tmp path:" + app.BASE_PATH + "/data/tmp")
	file.DeleteAll(app.BASE_PATH + "/data/tmp")
}
