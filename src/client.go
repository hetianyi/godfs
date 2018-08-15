package main

import (
    "flag"
    "lib_client"
    "os"
    "path/filepath"
    "util/file"
    "util/logger"
    "validate"
    "app"
    "lib_common"
    "time"
    "container/list"
    "lib_common/bridge"
    "io"
    "strings"
    "regexp"
    "fmt"
    "errors"
    "util/timeutil"
)


var client *lib_client.Client
var trackerList *list.List
var checkChan chan int

// 对于客户端，只提供类似于mysql的客户端，每个client与所有的tracker建立单个连接进行数据同步
// client和每个storage server最多建立一个连接
// 三方客户端可以开发成为一个连接池

func main() {
    checkChan = make(chan int)
    abs, _ := filepath.Abs(os.Args[0])
    s, _ := filepath.Split(abs)
    s = file.FixPath(s) // client executor parent path

    // set client type
    app.CLIENT_TYPE = 2
    //for test
    //a := "D:/nginx-1.8.1.zip"

    // the file to be upload
    var uploadFile = flag.String("u", "", "the file to be upload")
    // the file to download
    var downFile = flag.String("d", "", "the file to be download")
    // the download file name
    var customDownloadFileName = flag.String("n", "", "custom download file name")
    // the download file name
    var logLevel = flag.String("l", "", "custom logging level: trace, debug, info, warning, error, and fatal")
    // config file path
    var confPath = flag.String("c", s + string(filepath.Separator) + ".." + string(filepath.Separator) + "conf" + string(filepath.Separator) + "client.conf", "custom config file")
    flag.Parse()

    *logLevel = strings.ToLower(strings.TrimSpace(*logLevel))
    if *logLevel != "trace" && *logLevel != "debug" && *logLevel != "info" && *logLevel != "warning" && *logLevel != "error" && *logLevel != "fatal" {
        *logLevel = ""
    }
    validate.SetSystemLogLevel(*logLevel)

    logger.Info("using config file:", *confPath)
    m, e := file.ReadPropFile(*confPath)
    if e == nil {
        if m["log_level"] == "" {
            m["log_level"] = *logLevel
        }
        app.RUN_WITH = 3
        validate.Check(m, 3)
        if *uploadFile != "" || *downFile != "" {
            client = Init()
        }
        if *uploadFile != "" {
            upload(*uploadFile)
        }
        if *downFile != "" {
            download(*downFile, strings.TrimSpace(*customDownloadFileName))
        }
        if *uploadFile == "" && *downFile == "" {
            fmt.Println("godfs usage:")
            fmt.Println("\t-u \n\t\tthe file to be upload")
            fmt.Println("\t-d \n\t\tthe file to be download")
            fmt.Println("\t-l \n\t\tcustom logging level: trace, debug, info, warning, error, and fatal")
            fmt.Println("\t-n \n\t\tcustom download file name")
        }
    } else {
        logger.Fatal("error read file:", e)
    }
}

func upload(path string) error {
    var startTime time.Time
    fid, e := client.Upload(path, "")
    if e != nil {
        logger.Error(e)
    } else {
        now := time.Now()
        fmt.Println("[==========] 100% ["+ timeutil.GetHumanReadableDuration(startTime, now) +"]\nupload success, file id:")
        fmt.Println("+--------------------------------------------+")
        fmt.Println("| "+ fid +" |")
        fmt.Println("+--------------------------------------------+")
    }
    return nil
}



func download(path string, customDownloadFileName string) error {
    filePath := ""
    var startTime time.Time
    e := client.DownloadFile(path, 0, -1, func(fileLen uint64, reader io.Reader) error {
        var fi *os.File
        if customDownloadFileName == "" {
            md5 := regexp.MustCompile(app.PATH_REGEX).ReplaceAllString(path, "${4}")
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
        buffer := make([]byte, app.BUFF_SIZE)
        filePath = fi.Name()
        startTime = time.Now()
        return writeOut(reader, int64(fileLen), buffer, fi, startTime)
    })
    if e != nil {
        logger.Error("download failed:", e)
        return e
    } else {
        now := time.Now()
        fmt.Println("[==========] 100% ["+ timeutil.GetHumanReadableDuration(startTime, now) +"]\ndownload success, file save as:")
        fmt.Println("+--------------------------------------------+")
        fmt.Println("| "+ filePath)
        fmt.Println("+--------------------------------------------+")
    }
    return nil
}

func Init() *lib_client.Client {
    client:= lib_client.NewClient(10)
    collector := lib_client.TaskCollector {
        Interval: time.Millisecond * 30,
        FirstDelay: 0,
        Name: "::: synchronize storage server instances :::",
        Job: clientMonitorCollector,
    }
    collectors := *new(list.List)
    collectors.PushBack(&collector)
    maintainer := &lib_client.TrackerMaintainer{Collectors: collectors}
    maintainer.Maintain(app.TRACKERS)
    trackerList = lib_common.ParseTrackers(app.TRACKERS)
    logger.Info("synchronize with trackers...")
    for i := 0; i < trackerList.Len(); i++ {
        <- checkChan
    }
    return client
}

func clientMonitorCollector(tracker *lib_client.TrackerInstance) {
    task := &bridge.Task{
        TaskType: app.TASK_SYNC_ALL_STORAGES,
        Callback: func(task *bridge.Task, e error) {
            checkChan <- 1
        },
    }
    lib_client.AddTask(task, tracker)
}



func writeOut(in io.Reader, offset int64, buffer []byte, out io.Writer, startTime time.Time) error {
    var finish, total int64
    var stopFlag = false
    defer func() {stopFlag = true}()
    total = offset
    finish = 0
    go lib_common.ShowPercent(&total, &finish, &stopFlag, startTime)

    // total read bytes
    var readBodySize int64 = 0
    // next time bytes to read
    var nextReadSize int
    for {
        // left bytes is more than a buffer
        if (offset - readBodySize) / int64(len(buffer)) >= 1 {
            nextReadSize = len(buffer)
        } else {// left bytes less than a buffer
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
