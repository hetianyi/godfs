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
    "errors"
    "util/timeutil"
    "hash"
    "app"
)

var p, _ = pool.NewPool(1000, 100000)

var secret string

// Start service and listen
// 1. Start task for upload listen
// 2. Start task for communication with tracker
func StartService(config map[string] string) {
    trackers := config["trackers"]
    port := config["port"]
    secret = config["secret"]
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
                            p.Exec(func() {
                                uploadHandler(conn)
                            })
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
            ee := operationUpload(meta, bodySize, bodyBuff, md, conn)
            if ee != nil {
                logger.Error("error read upload file:", ee)
                break
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
    checkStatus, _ := checkUploadMeta(meta,conn)
    // if secret validate failed or meta parse error
    if checkStatus != 0 {
        lib_common.Close(conn)
        var response = &header.UploadResponseMeta{
            Status: 1,
            Path: "",
        }
        e5 := lib_common.WriteResponse(4, conn, response)
        if e5 != nil {
            logger.Error(e5)
        }
        return errors.New("error check meta")
    }

    // begin upload file
    tmpFileName := timeutil.GetUUID()
    logger.Info("begin read file body, file len is ", bodySize/1024, "KB")
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

    dig := strings.ToUpper(md5[0:2])
    finalPath := app.BASE_PATH + "/data/" + dig + "/" + md5
    if !file.Exists(finalPath) {
        eee := file.MoveFile(tmpPath, finalPath)
        if eee != nil {
            logger.Error("error move tmp file from", tmpPath, "to", finalPath)
            // upload success
            var response = &header.UploadResponseMeta{
                Status: 3,
                Path: "",
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

    // upload success
    var response = &header.UploadResponseMeta{
        Status: 0,
        Path: app.GROUP + "/" + app.INSTANCE_ID + "/" + dig + "/" + md5,
    }
    e5 := lib_common.WriteResponse(4, conn, response)
    if e5 != nil {
        lib_common.Close(conn)
        return e5
    }
    return nil
}






