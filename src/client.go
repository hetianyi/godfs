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
// echo \"$(ls -m /f/Software)\" |xargs /e/godfs-storage/client/bin/go_build_client_go -u
func main() {


    fmt.Println(os.Args[0])

    checkChan = make(chan int)
    abs, _ := filepath.Abs(os.Args[0])
    s, _ := filepath.Split(abs)
    s = file.FixPath(s) // client executor parent path

    // set client type
    app.CLIENT_TYPE = 2

    // the file to be upload
    var uploadFile = flag.String("u", "", "the file to be upload, if you want upload many file once, quote file paths using \"\"\" and split with \",\"")
    // the file to download
    var downFile = flag.String("d", "", "the file to be download")
    // the download file name
    var customDownloadFileName = flag.String("n", "", "custom download file name")
    // custom override log level
    var logLevel = flag.String("l", "", "custom logging level: trace, debug, info, warning, error, and fatal")
    // config file path
    var confPath = flag.String("c", s + string(filepath.Separator) + ".." + string(filepath.Separator) + "conf" + string(filepath.Separator) + "client.conf", "custom config file")
    // whether check file md5 before upload
    var beforeCheck = false//flag.Bool("-skip-check", true, "whether check file md5 before upload, true|false")
    flag.Parse()

    *logLevel = strings.ToLower(strings.TrimSpace(*logLevel))
    if *logLevel != "trace" && *logLevel != "debug" && *logLevel != "info" && *logLevel != "warning" && *logLevel != "error" && *logLevel != "fatal" {
        *logLevel = ""
    }

    logger.Info("Usage of godfs client:", *confPath)
    m, e := file.ReadPropFile(*confPath)
    if e == nil {
        if *logLevel != "" {
            m["log_level"] = *logLevel
        }
        logger.Debug("uploadFile=" + *uploadFile)
        app.RUN_WITH = 3
        validate.Check(m, 3)
        for k, v := range m {
            logger.Debug(k, "=", v)
        }
        if *uploadFile != "" || *downFile != "" {
            client = Init()
        }
        if *uploadFile != "" {
            upload(*uploadFile, beforeCheck)
        }
        if *downFile != "" {
            download(*downFile, strings.TrimSpace(*customDownloadFileName))
        }
        if *uploadFile == "" && *downFile == "" {
            fmt.Println("godfs client usage:")
            fmt.Println("\t-u string \n\t    the file to be upload, if you want upload many file once, quote file paths using \"\"\" and split with \",\"" +
                "\n\t    example:\n\t\tclient -u \"/home/foo/bar1.tar.gz, /home/foo/bar1.tar.gz\"")
            fmt.Println("\t-d string \n\t    the file to be download")
            fmt.Println("\t-l string \n\t    custom logging level: trace, debug, info, warning, error, and fatal")
            fmt.Println("\t-n string \n\t    custom download file name")
            fmt.Println("\t--skip-check bool \n\t    whether check file md5 before upload, true|false")
        }
    } else {
        logger.Fatal("error read file:", e)
    }
}

// upload files
//TODO support md5 check before upload
func upload(paths string, beforeCheck bool) error {
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
        fid, e := client.Upload(ele.Value.(string), "", startTime, beforeCheck)
        if e != nil {
            logger.Error(e)
        } else {
            now := time.Now()
            fmt.Println("[==========] 100% ["+ timeutil.GetHumanReadableDuration(startTime, now) +"]\nupload success, file id:")
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
        buffer := make([]byte, app.BUFF_SIZE)
        filePath, _ = filepath.Abs(fi.Name())
        startTime = time.Now()
        return writeOut(reader, int64(fileLen), buffer, fi, startTime)
    })
    if e != nil {
        logger.Error("download failed:", e)
        return e
    } else {
        now := time.Now()
        fmt.Println("[==========] 100% ["+ timeutil.GetHumanReadableDuration(startTime, now) +"]\ndownload success, file save as:")
        fmt.Println("+-------------------------------------------+")
        fmt.Println(filePath)
        fmt.Println("+-------------------------------------------+")
    }
    return nil
}

func Init() *lib_client.Client {
    client:= lib_client.NewClient(10)
    collector := lib_client.TaskCollector {
        Interval: time.Millisecond * 30,
        FirstDelay: 0,
        ExecTimes: 1,
        Name: "::: synchronize storage server instances :::",
        Job: clientMonitorCollector,
    }
    collectors := new(list.List)
    collectors.PushBack(&collector)
    maintainer := &lib_client.TrackerMaintainer{Collectors: *collectors}
    client.TrackerMaintainer = maintainer
    trackerList = maintainer.Maintain(app.TRACKERS)
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



