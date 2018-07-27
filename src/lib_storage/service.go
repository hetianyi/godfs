package lib_storage

import (
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
    "lib_client"
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
    go startTrackerMaintainer(trackers)
    startStorageService(port)
}


func startTrackerMaintainer(trackers string) {
    collector1 := lib_client.TaskCollector {
        Interval: time.Second * 30,
        Name: "推送本地新文件到tracker",
        Single: false,
        Job: lib_client.QueryPushFileTaskCollector,//TODO 多tracker如何保证全部tracker成功？
    }
    collector2 := lib_client.TaskCollector {
        Interval: time.Second * 30,
        Name: "推送本地新文件到tracker",
        Single: true,
        Job: lib_client.QueryDownloadFileTaskCollector,
    }
    collector3 := lib_client.TaskCollector {
        Interval: time.Second * 30,
        Name: "拉取tracker新文件",
        Single: false,
        Job: lib_client.QueryNewFileTaskCollector,
    }
    collector4 := lib_client.TaskCollector {
        Interval: time.Second * 30,
        Name: "同步storage成员",
        Single: false,
        Job: lib_client.SyncMemberTaskCollector,
    }
    collectors := *new(list.List)
    collectors.PushBack(&collector1)
    collectors.PushBack(&collector2)
    collectors.PushBack(&collector3)
    collectors.PushBack(&collector4)

    maintainer := &lib_client.TrackerMaintainer{Collectors: collectors}
    maintainer.Maintain(trackers)
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
                    return QueryFileHandler(request, connBridge, 1)
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

