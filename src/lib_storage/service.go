package lib_storage

import (
    "strings"
    "util/logger"
    "container/list"
    "regexp"
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
    pt := parsePort(port)
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
                    for  {
                        conn, e1 := listener.Accept()
                        if e1 == nil {
                            p.Exec(func() {
                                uploadHandler(conn)
                            })
                        } else {
                            logger.Info("accept new conn error", e1)
                        }
                    }
                }
            }, func(i interface{}) {
                logger.Error("["+ strconv.Itoa(tryTimes) +"] error shutdown service duo to:", i)
                time.Sleep(time.Second * 5)
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



func uploadHandler(conn net.Conn) {
    defer func() {
        logger.Debug("close connection from server")
        conn.Close()
    }()
    common.Try(func() {
        bodyBuff := make([]byte, lib_common.BodyBuffSize)     // body buff
        index := 0 //test
        md := md5.New()
        for {
            operation, meta, bodySize, err := lib_common.ParseConnRequestMeta(conn)
            //respond unSupport operation
            if operation == 0 {
                //TODO write response
                //close conn
                break
            }

            if operation == -1 || meta == "" || err != nil {
                // otherwise mark as broken connection
                if err != nil {
                    logger.Error(err)
                }
                break
            }

            checkStatus, _ := checkUploadMeta(meta,conn)
            // if secret validate failed or meta parse error
            if checkStatus != 0 {
                lib_common.Close(conn)
                break
            }
            index++
            // begin upload file
            logger.Info("开始上传文件，文件大小：", bodySize/1024, "KB")
            fi, _ := file.CreateFile("D:\\godfs\\nginx-1.8.1("+ strconv.Itoa(index) +").zip")

            e4 := lib_common.ParseConnRequestBody(bodySize, bodyBuff, conn, fi, md)
            if e4 != nil {
                logger.Error(e4, "delete file")
                file.Delete(fi.Name())
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

func parsePort(port string) int {
    if len(port) < 2 {
        logger.Fatal("parameter 'port' not set yet, server will not exit now!")
    }
    if b, _ := regexp.Match("^[1-9][0-9]{1,6}$", []byte(port)); b {
        p, e := strconv.Atoi(port)
        if e != nil || p > 65535 {
            logger.Fatal("parameter 'port' must be a valid port number!")
            return 0
        }
        return p
    }
    return 0
}


// 处理注册storage
func checkUploadMeta(meta string, conn net.Conn) (int, *header.UploadRequestMeta) {
    headerMeta := &header.UploadRequestMeta{}
    e2 := json.Unmarshal([]byte(meta), &headerMeta)
    if e2 == nil {
        if headerMeta.Secret == secret {
            return 0, headerMeta // success
        } else {
            //TODO write response
            //close conn
            return 1, headerMeta // bad secret
        }
    } else {
        return 2, nil // parse meta error
    }
}
