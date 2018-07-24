package lib_storage

import (
    "strings"
    "util/logger"
    "container/list"
    "strconv"
    "net"
    "util/pool"
    "time"
    "crypto/md5"
    "lib_common"
    "util/common"
    "app"
    "io"
    "net/http"
    "util/db"
    "lib_common/bridge"
)


// max client connection set to 1000
var p, _ = pool.NewPool(1000, 100000)
// sys secret
var secret string
// sys config
var cfg map[string] string
// tasks put in this list
var fileRegisterList list.List

// Start service and listen
// 1. Start task for upload listen
// 2. Start task for communication with tracker
func StartService(config map[string] string) {
    cfg = config
    trackers := config["trackers"]
    port := config["port"]
    secret = config["secret"]

    // 连接数据库
    db.InitDB()

    startHttpDownloadService()
    go startConnTracker(trackers)
    // start task collector
    go startTaskCollector()
    startStorageService(port)
}


// upload listen
func startStorageService(port string) {
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
                                clientHandler(conn)
                            })
                            // maybe the poll is full
                            if ee != nil {
                                bridge.Close(conn)
                            }
                        } else {
                            logger.Info("accept new conn error", e1)
                            if conn != nil {
                                bridge.Close(conn)
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

// http download service
func startHttpDownloadService() {
    if !app.HTTP_ENABLE {
        logger.Info("http server disabled")
        return
    }

    http.HandleFunc("/download/", DownloadHandler)

    s := &http.Server{
        Addr:           ":" + strconv.Itoa(app.HTTP_PORT),
        ReadTimeout:    10 * time.Second,
        WriteTimeout:   0,
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

// connect to each tracker
func onceConnTracker(tracker string) {
    logger.Info("start tracker conn with tracker server:", tracker)
    retry := 0
    for {//keep trying to connect to tracker server.
        conn, e := net.Dial("tcp", tracker)
        if e == nil {
            // validate client
            connBridge, e1 := connectAndValidate(conn)
            if e1 != nil {
                bridge.Close(conn)
                logger.Error(e1)
            } else {
                retry = 0
                logger.Debug("connect to tracker server success.")
                for { // keep sending client statistic info to tracker server.
                    task := GetTask()
                    if task == nil {
                        time.Sleep(time.Second * 1)
                        continue
                    }
                    forceClosed, e2 := ExecTask(task, connBridge)
                    if e2 != nil {
                        logger.Debug("task exec error:", e2)
                        FailReturnTask(task)
                    } else {
                        logger.Debug("task exec success:", task.TaskType)
                    }
                    if forceClosed {
                        logger.Debug("connection force closed by client")
                        bridge.Close(conn)
                        break
                    }
                }
            }
        } else {
            logger.Error("(" + strconv.Itoa(retry) + ")error connect to tracker server:", tracker)
        }
        retry++
        time.Sleep(time.Second * 1)
    }
}

// connect to tracker server and register client to it.
func connectAndValidate(conn net.Conn) (*bridge.Bridge, error) {
    // create bridge
    connBridge := bridge.NewBridge(conn)
    // send validate request
    e1 := connBridge.ValidateConnection("")
    if e1 != nil {
        return nil, e1
    }
    return connBridge, nil
}



// accept a new connection for file upload
// the connection will keep till it is broken
// 文件同步策略：
// 文件上传成功将任务写到本地文件storage_task.data作为备份
// 将任务通知到tracker服务器，通知成功，tracker服务进行广播
// 其他storage定时取任务，将任务
func clientHandler(conn net.Conn) {
    defer func() {
        logger.Debug("close connection from server")
        bridge.Close(conn)
    }()
    common.Try(func() {
        // body buff
        bodyBuff := make([]byte, app.BUFF_SIZE)
        // calculate md5
        md := md5.New()
        connBridge := bridge.NewBridge(conn)
        for {
            error := connBridge.ReceiveRequest(func(request *bridge.Meta, in io.ReadCloser) error {
                //return requestRouter(request, &bodyBuff, md, connBridge, conn)
                if request.Err != nil {
                    return request.Err
                }
                // route
                if request.Operation == bridge.O_CONNECT {
                    return validateClientHandler(request, connBridge)
                } else if request.Operation == bridge.O_UPLOAD {
                    return uploadHandler(request, bodyBuff, md, conn, connBridge)
                } else if request.Operation == bridge.O_QUERY_FILE {
                    return queryFileHandler(request, connBridge)
                } else if request.Operation == bridge.O_DOWNLOAD_FILE {
                    return downloadFileHandler(request, bodyBuff, connBridge)
                } else {
                    return bridge.OPERATION_NOT_SUPPORT_ERROR
                }
                return nil
            })
            if error != nil {
                logger.Error(error)
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
