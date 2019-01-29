package validate

import (
	"app"
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"util/common"
	"util/file"
	"util/logger"
)

var (
	az                   = []rune{'A', 'B', 'C', 'D', 'E', 'F'}
	i09                  = []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
	GroupInstancePattern = "^[0-9a-zA-Z_]{1,30}$" // group name and instance id pattern
)

// check configuration file parameter.
// if check failed, system will shutdown.
// runWith:
//        1: storage server
//        2: tracker server
//        3: client
//        4: dashboard
func Check(m map[string]string, runWith int) {
	replaceParams(m)

	// client doesn't has base path
	if runWith != 3 {
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
		app.BasePath = m["base_path"]
		cleanTmpDir()
		prepareDirs(m["base_path"])
	}

	// check secret
	m["secret"] = strings.TrimSpace(m["secret"])
	app.Secret = m["secret"]

	// check log_level
	logLevel := strings.ToLower(strings.TrimSpace(m["log_level"]))
	if logLevel != "trace" && logLevel != "debug" && logLevel != "info" && logLevel != "warn" &&
		logLevel != "error" && logLevel != "fatal" {
		logLevel = "info"
	}
	m["log_level"] = logLevel
	SetSystemLogLevel(logLevel)

	// check log_rotation_interval
	logRotationInterval := strings.ToLower(strings.TrimSpace(m["log_rotation_interval"]))
	if logRotationInterval != "h" && logRotationInterval != "d" &&
		logRotationInterval != "m" && logRotationInterval != "y" {
		logRotationInterval = "d"
	}
	m["log_rotation_interval"] = logRotationInterval
	app.LogInterval = logRotationInterval

	// enable log config
	logEnable := strings.ToLower(strings.TrimSpace(m["log_enable"]))
	if logEnable == "" || (logEnable != "true" && logEnable != "false") {
		logEnable = "true"
	}
	if logEnable == "true" {
		app.LogEnable = true
		logger.SetEnable(true)
	} else {
		app.LogEnable = false
		logger.SetEnable(false)
	}

	if runWith == 1 || runWith == 2 {
		// check port
		port, e := strconv.Atoi(m["port"])
		if e == nil {
			if port <= 0 || port > 65535 {
				logger.Fatal("invalid port range:", m["port"])
			}
			app.Port = port
		} else {
			logger.Fatal("invalid port ", m["port"], ":", e)
		}
	}

	if runWith == 1 {
		// check: advertise_addr
		advertiseAddr := strings.TrimSpace(m["advertise_addr"])
		app.AdvertiseAddress = advertiseAddr

		if strings.TrimSpace(m["advertise_port"]) == "" {
			app.AdvertisePort = app.Port
		} else {
			advertisePort, e := strconv.Atoi(strings.TrimSpace(m["advertise_port"]))
			if e == nil {
				if advertisePort <= 0 || advertisePort > 65535 {
					logger.Warn("invalid advertise_port range:", m["advertise_port"]+", use default port", app.Port)
					app.AdvertisePort = app.Port
				} else {
					app.AdvertisePort = advertisePort
				}
			} else {
				logger.Fatal("invalid advertise_port ", m["advertise_port"], ":", e, ", use default port", app.Port)
				app.AdvertisePort = app.Port
			}
		}

		// check GROUP
		m["group"] = strings.TrimSpace(m["group"])
		if mat, _ := regexp.Match(GroupInstancePattern, []byte(m["group"])); !mat {
			logger.Fatal("error parameter 'group'")
		}
		app.Group = m["group"]

		// check instance id
		m["instance_id"] = strings.TrimSpace(m["instance_id"])
		if m["instance_id"] != "" {
			if mat, _ := regexp.Match(GroupInstancePattern, []byte(m["instance_id"])); !mat {
				logger.Fatal("error parameter 'instance_id'")
			}
		}
		app.InstanceId = m["instance_id"]

		// check assign_disk_space
		assignDiskSpace := strings.ToLower(strings.TrimSpace(m["assign_disk_space"]))
		value, unit := FixStorageSize(assignDiskSpace, "MB")
		if value == "" {
			value = "50"
		}
		if unit == "" {
			unit = "MB"
		}
		val, e3 := strconv.ParseFloat(value, 64)
		if e3 != nil {
			logger.Fatal("error assign_disk_space:", value+unit)
		}
		var sizeUnit1 = GetUnitVal(unit)
		app.AssignDiskSpace = int64(val * float64(sizeUnit1))
		m["assign_disk_space"] = value + unit

		// check slice_size
		sliceSize := strings.ToLower(strings.TrimSpace(m["slice_size"]))
		value1, unit1 := FixStorageSize(sliceSize, "MB")
		if value1 == "" {
			value1 = "50"
		}
		if unit1 == "" {
			unit1 = "MB"
		}
		val1, e4 := strconv.ParseFloat(value1, 64)
		if e4 != nil {
			logger.Fatal("error slice_size:", value1+unit1)
		}
		var sizeUnit2 = GetUnitVal(unit1)
		app.SliceSize = int64(val1 * float64(sizeUnit2))
		m["slice_size"] = value1 + unit1
		logger.Debug("slice_size:", app.SliceSize)

		// check upload_enable
		uploadEnable := strings.ToLower(strings.TrimSpace(m["upload_enable"]))
		if uploadEnable != "true" && uploadEnable != "false" {
			uploadEnable = "true"
		}
		m["upload_enable"] = uploadEnable
		app.UploadEnable = uploadEnable == "true"

		// check enable_mime_types
		enableMimeTypes := strings.ToLower(strings.TrimSpace(m["enable_mime_types"]))
		if enableMimeTypes != "true" && enableMimeTypes != "false" {
			enableMimeTypes = "true"
		}
		m["enable_mime_types"] = enableMimeTypes
		app.MimeTypesEnable = enableMimeTypes == "true"
		if app.MimeTypesEnable {
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

		// check web_content_mime_types
		m["access_control_allow_origin"] = strings.TrimSpace(m["access_control_allow_origin"])
		acao := strings.Split(m["access_control_allow_origin"], ",")
		for i := range acao {
			strS := strings.TrimSpace(acao[i])
			if strS == "" {
				continue
			}
			app.AddAccessAllowOrigin(strS)
		}

		// check: preferred_networks
		preferredNetworks := strings.TrimSpace(m["preferred_networks"])
		if len(preferredNetworks) > 0 {
			splitNetworkNames := strings.Split(preferredNetworks, ",")
			for i := range splitNetworkNames {
				name := strings.TrimSpace(splitNetworkNames[i])
				if name != "" {
					app.PreferredNetworks.PushBack(name)
				}
			}
		}
		if app.PreferredNetworks.Len() > 0 {
			tmpStr := "("
			common.WalkList(&app.PreferredNetworks, func(item interface{}) bool {
				tmpStr += item.(string) + "|"
				return false
			})
			tmpStr = tmpStr[0:len(tmpStr) - 1]
			tmpStr += ")"
			logger.Info("preferred network interfaces:", tmpStr)
		}


		// check: preferred_networks
		preferredIPPrefix := strings.TrimSpace(m["preferred_ip_prefix"])
		app.PreferredIPPrefix = preferredIPPrefix
	}

	if runWith == 1 || runWith == 2 || runWith == 4 {

		// check http_enable
		httpEnable := strings.ToLower(strings.TrimSpace(m["http_enable"]))
		if httpEnable != "true" && httpEnable != "false" {
			httpEnable = "false"
		}
		m["http_enable"] = httpEnable
		app.HttpEnable = httpEnable == "true"

		// check http_port
		httpPortStr := strings.ToLower(strings.TrimSpace(m["http_port"]))

		httpPort, ehp := strconv.Atoi(httpPortStr)
		if runWith == 4 || ((runWith == 1 || runWith == 2) && app.HttpEnable) {
			if ehp != nil || httpPort <= 0 || httpPort > 65535 {
				logger.Fatal("error http_port:", httpPortStr)
			} else {
				m["http_port"] = httpPortStr
				app.HttpPort = httpPort
			}
		}

		// check http auth
		m["http_auth"] = strings.TrimSpace(m["http_auth"])
		app.HttpAuth = m["http_auth"]
	}

	if runWith != 2 {
		// check trackers
		trackers := strings.TrimSpace(m["trackers"])
		ts := strings.Split(trackers, ",")
		var byteBuff bytes.Buffer
		for i := range ts {
			strS := strings.TrimSpace(ts[i])
			if strS == "" {
				continue
			}
			byteBuff.WriteString(strS)
			if i < len(ts)-1 {
				byteBuff.WriteString(",")
			}
		}
		m["trackers"] = string(byteBuff.Bytes())
		app.Trackers = m["trackers"]
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
	// create dir now disabled
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
	var sizeUnit int64
	if unit == "BB" {
		sizeUnit = 1
	} else if unit == "KB" {
		sizeUnit = 1024
	} else if unit == "MB" {
		sizeUnit = 1024 * 1024
	} else if unit == "GB" {
		sizeUnit = 1024 * 1024 * 1024
	} else if unit == "PB" {
		sizeUnit = 1024 * 1024 * 1024 * 1024 * 1024
	} else {
		sizeUnit = 0
	}
	return sizeUnit
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
	// if base path file exists and it is a file.
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

// clean tmp dir before boot
func cleanTmpDir() {
	logger.Debug("clean tmp path:" + app.BasePath + "/data/tmp")
	file.DeleteAll(app.BasePath + "/data/tmp")
}
