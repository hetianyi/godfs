package lib_storage

import (
    "strings"
    "util/logger"
    "container/list"
    "strconv"
    "net"
    "util/pool"
    "util/file"
    "encoding/json"
    "time"
    "crypto/md5"
    "lib_common"
    "lib_common/header"
    "util/common"
    "util/timeutil"
    "hash"
    "app"
    "regexp"
    "io"
    "errors"
    "net/http"
)


// max client connection set to 1000
var p, _ = pool.NewPool(1000, 100000)

var secret string

// Start service and listen
// 1. Start task for upload listen
// 2. Start task for communication with tracker
func StartService(config map[string] string) {
    trackers := config["trackers"]
    port := config["port"]
    secret = config["secret"]
    startDownloadService()
    startUploadService(port)
    startConnTracker(trackers)
}


// upload listen
func startUploadService(port string) {
    pt := lib_common.ParsePort(port)
    if pt > 0 {
        tryTimes := 0
        for {
            common.Try(func() {
                listener, e := net.Listen("tcp", ":" + strconv.Itoa(pt))
                logger.Info("service listening on port:", pt)
                if e != nil {
                    panic(e)
                } else {
                    // keep accept connections.
                    for {
                        conn, e1 := listener.Accept()
                        if e1 == nil {
                            ee := p.Exec(func() {
                                uploadHandler(conn)
                            })
                            // maybe the poll is full
                            if ee != nil {
                                lib_common.Close(conn)
                            }
                        } else {
                            logger.Info("accept new conn error", e1)
                            if conn != nil {
                                lib_common.Close(conn)
                            }
                        }
                    }
                }
            }, func(i interface{}) {
                logger.Error("["+ strconv.Itoa(tryTimes) +"] error shutdown service duo to:", i)
                time.Sleep(time.Second * 10)
            })
        }
    }
}

func startDownloadService() {

    if !app.HTTP_ENABLE {
        logger.Info("http server disabled")
        return
    }

    http.HandleFunc("/download/", DownloadHandler)

    s := &http.Server{
        Addr:           ":" + strconv.Itoa(app.HTTP_PORT),
        ReadTimeout:    10 * time.Second,
        WriteTimeout:   100 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }
    logger.Info("http server listen on port:", app.HTTP_PORT)
    go s.ListenAndServe()
}

// communication with tracker
func startConnTracker(trackers string) {
    ls := parseTrackers(trackers)
    if ls.Len() == 0 {
        logger.Warn("no trackers set, the storage server will run in stand-alone mode.")
        return
    }

    for e := ls.Front(); e != nil; e = e.Next() {
        go onceConnTracker(e.Value.(string))
    }
}

func onceConnTracker(tracker string) {
    logger.Info("start tracker conn with tracker server:", tracker)
}


// accept a new connection for file upload
// the connection will keep till it is broken
func uploadHandler(conn net.Conn) {

    defer func() {
        logger.Debug("close connection from server")
        conn.Close()
    }()

    common.Try(func() {
        // body buff
        bodyBuff := make([]byte, lib_common.BodyBuffSize)
        // calculate md5
        md := md5.New()
        for {
            // read meta
            operation, meta, bodySize, err := lib_common.ReadConnMeta(conn)
            if operation == -1 || meta == "" || err != nil {
                // otherwise mark as broken connection
                if err != nil {
                    logger.Error(err)
                }
                break
            }

            // check secret
            checkStatus, _ := checkUploadMeta(meta, conn)
            // if secret validate failed or meta parse error
            if checkStatus != 0 {
                break
            }

            // client wants to upload file
            if operation == 2 {
                ee := operationUpload(meta, bodySize, bodyBuff, md, conn)
                if ee != nil {
                    logger.Error("error read upload file:", ee)
                    break
                }
            } else if operation == 5 {// client wants to query file
                ee := operationQueryFile(meta, conn)
                if ee != nil {
                    logger.Error("error query file:", ee)
                    break
                }
            } else if operation == 6 {// client wants to query file
                ee := operationDownloadFile(meta, bodyBuff, conn)
                if ee != nil {
                    logger.Error("error query file:", ee)
                    break
                }
            }
        }
    }, func(i interface{}) {
        logger.Error("connection error:", i)
    })
}


// parse trackers into a list
func parseTrackers(tracker string) *list.List {
    sp := strings.Split(tracker, ",")
    ls := list.New()
    for i := range sp {
        trimS := strings.TrimSpace(sp[i])
        if len(trimS) > 0 {
            ls.PushBack(trimS)
        }

    }
    return ls
}


// 处理注册storage
func checkUploadMeta(meta string, conn net.Conn) (int, *header.UploadRequestMeta) {
    headerMeta := &header.UploadRequestMeta{}
    e2 := json.Unmarshal([]byte(meta), &headerMeta)
    if e2 == nil {
        if headerMeta.Secret == secret {
            return 0, headerMeta // success
        } else {
            var response = &header.UploadResponseMeta{
                Status: 1,
                Path: "",
                Exist: false,
            }
            // write response close conn, and not check if success
            lib_common.WriteResponse(4, conn, response)
            //close conn
            lib_common.Close(conn)
            logger.Error("error check secret")
            return 1, headerMeta // bad secret
        }
    } else {
        return 2, nil // parse meta error
    }
}

// 处理文件上传请求
func operationUpload(meta string, bodySize uint64, bodyBuff []byte, md hash.Hash, conn net.Conn) error {

    logger.Info("begin read file body, file len is ", bodySize/1024, "KB")
    // begin upload file
    tmpFileName := timeutil.GetUUID()
    // using tmp ext and rename after upload success
    tmpPath := file.FixPath(app.BASE_PATH + "/data/tmp/" + tmpFileName)
    fi, e8 := file.CreateFile(tmpPath)
    if e8 != nil {
        lib_common.Close(conn)
        logger.Error("error create file")
        return e8
    }

    // read file bytes
    md5, e4 := lib_common.ReadConnBody(bodySize, bodyBuff, conn, fi, md)
    if e4 != nil {
        file.Delete(fi.Name())
        return e4
    }

    dig1 := strings.ToUpper(md5[0:2])
    dig2 := strings.ToUpper(md5[2:4])
    finalPath := app.BASE_PATH + "/data/" + dig1 + "/" + dig2
    if !file.Exists(finalPath) {
        e := file.CreateAllDir(finalPath)
        if e != nil {
            return e4
        }
    }
    if !file.Exists(finalPath + "/" + md5) {
        eee := file.MoveFile(tmpPath, finalPath + "/" + md5)
        if eee != nil {
            logger.Error("error move tmp file from", tmpPath, "to", finalPath)
            // upload success
            var response = &header.UploadResponseMeta{
                Status: 3,
                Path: "",
                Exist: false,
            }
            e9 := lib_common.WriteResponse(4, conn, response)
            if e9 != nil {
                lib_common.Close(conn)
                return e9
            }
        }
    } else {
        s := file.Delete(tmpPath)
        if !s {
            logger.Error("error clean tmp file:", tmpPath)
        }
    }

    // try log file md5, but not
    /*common.Try(func() {
        db.AddFile(md5)
    }, func(i interface{}) {
        logger.Error(i)
    })*/


    // upload success
    var response = &header.UploadResponseMeta{
        Status: 0,
        Path: app.GROUP + "/" + app.INSTANCE_ID + "/" + md5,
        Exist: true,
    }
    e5 := lib_common.WriteResponse(4, conn, response)
    if e5 != nil {
        lib_common.Close(conn)
        return e5
    }
    return nil
}


// 处理文件上传请求
func operationQueryFile(meta string, conn net.Conn) error {
    var queryMeta = &header.QueryFileRequestMeta{}
    e := json.Unmarshal([]byte(meta), queryMeta)
    if e != nil {
        var response = &header.UploadResponseMeta {
            Status: 3,
            Path: "",
            Exist: false,
        }
        lib_common.WriteResponse(4, conn, response)
        lib_common.Close(conn)
        return e
    }

    dig1 := queryMeta.Md5[0:2]
    dig2 := queryMeta.Md5[2:4]
    finalPath := app.BASE_PATH + "/data/" + dig1 + "/" + dig2 + "/" + queryMeta.Md5

    var response = &header.UploadResponseMeta{}
    if file.Exists(finalPath) {
        response = &header.UploadResponseMeta {
            Status: 0,
            Path: "",
            Exist: true,
        }
    } else {
        response = &header.UploadResponseMeta {
            Status: 0,
            Path: "",
            Exist: false,
        }
    }
    e1 := lib_common.WriteResponse(4, conn, response)
    if e1 != nil {
        lib_common.Close(conn)
        return e1
    }
    return nil
}

// 处理文件下载请求
func operationDownloadFile(meta string, buff []byte, conn net.Conn) error {
    var queryMeta = &header.DownloadFileRequestMeta{}
    e := json.Unmarshal([]byte(meta), queryMeta)
    if e != nil {
        var response = &header.UploadResponseMeta {
            Status: 3,
            Path: "",
            Exist: false,
        }
        lib_common.WriteResponse(4, conn, response)
        lib_common.Close(conn)
        return e
    }

    pathRegex := "^/([0-9a-zA-Z_]+)/([0-9a-zA-Z_]+)/([0-9a-fA-F]{32})$"
    if mat, _ := regexp.Match(pathRegex, []byte(queryMeta.Path)); !mat {
        var response = &header.UploadResponseMeta {
            Status: 0,
            Path: "",
            Exist: false,
        }
        e1 := lib_common.WriteResponse(4, conn, response)
        if e1 != nil {
            lib_common.Close(conn)
            return e1
        }
        return nil
    }
    //initialServer := regexp.MustCompile(pathRegex).ReplaceAllString("/x_/_123/432597de0e65eedbc867620e744a35ad", "${2}")
    md5 := regexp.MustCompile(pathRegex).ReplaceAllString(queryMeta.Path, "${3}")

    finalPath := GetFilePathByMd5(md5)

    var response = &header.UploadResponseMeta{}
    if file.Exists(finalPath) {
        response = &header.UploadResponseMeta {
            Status: 0,
            Path: "",
            Exist: true,
        }
        downFile, e1 := file.GetFile(finalPath)
        if e1 != nil {
            response.Status = 3
            response.Exist = false
            e1 := lib_common.WriteResponse(4, conn, response)
            if e1 != nil {
                lib_common.Close(conn)
                return e1
            }
            return nil
        }
        fInfo, _ := downFile.Stat()
        response.Status = 0
        response.Exist = true
        response.FileSize = fInfo.Size()
        e4 := lib_common.WriteResponse(4, conn, response)
        if e4 != nil {
            lib_common.Close(conn)
            return e4
        }
        for {
            len, e2 := downFile.Read(buff)
            if e2 == nil || e2 == io.EOF {
                wl, e5 := conn.Write(buff[0:len])
                if e2 == io.EOF {
                    downFile.Close()
                    return nil
                }
                if e5 != nil || wl != len {
                    downFile.Close()
                    lib_common.Close(conn)
                    return errors.New("error handle download file")
                }
            } else {
                downFile.Close()
                return e2
            }
        }
    } else {
        response = &header.UploadResponseMeta {
            Status: 0,
            Path: "",
            Exist: false,
        }
        e1 := lib_common.WriteResponse(4, conn, response)
        if e1 != nil {
            lib_common.Close(conn)
            return e1
        }
    }
    return nil
}

// check if file exists on the filesystem
func CheckIfFileExistByMd5(md5 string) bool {
    if file.Exists(GetFilePathByMd5(md5)) {
        return true
    }
    return false
}

// return file path using md5
func GetFilePathByMd5(md5 string) string {
    dig1 := strings.ToUpper(md5[0:2])
    dig2 := strings.ToUpper(md5[2:4])
    return app.BASE_PATH + "/data/" + dig1 + "/" + dig2 + "/" + md5
}

