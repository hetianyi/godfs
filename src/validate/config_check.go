package validate

import (
    "strconv"
    "util/logger"
    "strings"
    "path/filepath"
    "os"
    "util/file"
    "bytes"
    "regexp"
    "lib_common"
)

// check configuration file parameter.
// if check failed, system will shutdown.
// runWith:
//        1: storage server
//        2: tracker server
func Check(m map[string] string, runWith int) {
    // check: bind_address
    //bind_address := m["bind_address"]

    // check port
    port, e := strconv.Atoi(m["port"])
    if e == nil {
        if port <= 0 || port > 65535 {
            logger.Fatal("invalid port range:", m["port"])
        }
    } else {
        logger.Fatal("invalid port ", m["port"], ":", e)
    }

    // check base_path
    basePath := strings.TrimSpace(m["base_path"])
    if basePath == "" {
        abs,_ := filepath.Abs(os.Args[0])
        parent, _ := filepath.Split(abs)
        finalPath := parent + "godfs"
        logger.Info("base_path not set, use", finalPath)
        m["base_path"] = finalPath
    } else {
        m["base_path"] = file.FixPath(basePath)
    }
    lib_common.BASE_PATH = m["base_path"]
    prepareDirs(m["base_path"])

    // check secret
    m["secret"] = strings.TrimSpace(m["secret"])
    // check log_level
    logLevel := strings.ToLower(strings.TrimSpace(m["log_level"]))
    if logLevel != "debug" && logLevel != "info" && logLevel != "warn" &&
        logLevel != "error" && logLevel != "fatal" {
        logLevel = "info"
    }   
    m["log_level"] = logLevel
    setSystemLogLevel(logLevel)

    // check log_rotation_interval
    log_rotation_interval := strings.ToLower(strings.TrimSpace(m["log_rotation_interval"]))
    if log_rotation_interval != "h" && log_rotation_interval != "d" &&
        log_rotation_interval != "m" && log_rotation_interval != "y" {
        log_rotation_interval = "d"
    }
    m["log_rotation_interval"] = log_rotation_interval
    lib_common.LOG_INTERVAL = log_rotation_interval

    //enable log config
    logger.SetEnable(true)


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
        logger.Fatal("error assign_disk_space:", value + unit)
    }
    var _unit = GetUnitVal(unit)
    lib_common.ASSIGN_DISK_SPACE = int64(_val * float64(_unit))
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
        logger.Fatal("error slice_size:", value1 + unit1)
    }
    var _unit1 = GetUnitVal(unit1)
    lib_common.SLICE_SIZE = int64(_val1 * float64(_unit1))
    m["slice_size"] = value1 + unit1


    if runWith == 1 {
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
        //--
    }

    if runWith == 2 {

    }
}


func createDirs(basePath string) {
    dataDir := basePath + string(os.PathSeparator) + "data"
    logsDir := basePath + string(os.PathSeparator) + "logs"
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
    if file.Exists(dataDir) && file.IsFile(dataDir) {
        logger.Fatal("cannot create data directory:", dataDir)
    }
    if file.Exists(logsDir) && file.IsFile(logsDir) {
        logger.Fatal("cannot create data directory:", logsDir)
    }
}

func FixStorageSize(input string, defaultUnit string) (value string, unit string) {
    input = strings.ToLower(input)
    if mat,e2 := regexp.Match("^([1-9][0-9]*)([kmgtp]?[b]?)$", []byte(input)); mat && e2 == nil {
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
        _unit = 1024*1024
    } else if unit == "GB" {
        _unit = 1024*1024*1024
    } else if unit == "PB" {
        _unit = 1024*1024*1024*1024*1024
    } else {
        _unit = 0
    }
    return _unit
}


func setSystemLogLevel(logLevel string) {
    logger.Info("log level set to", logLevel)
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