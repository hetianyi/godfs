package lib_storage

import (
    "strings"
    "util/logger"
    "container/list"
    "regexp"
    "strconv"
    "common"
    "net"
    "util/pool"
    "fmt"
    "io"
)

var p, _ = pool.NewPool(1000, 100000)


// Start service and listen
// 1. Start task for upload listen
// 2. Start task for communication with tracker
func StartService(config map[string] string) {
    trackers := config["trackers"]
    port := config["port"]
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
                logger.Error("["+ strconv.Itoa(tryTimes) +"] error start service duo to:", i)
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
    logger.Info("start trakcer conn with tracker server:", tracker)

}



func uploadHandler(conn net.Conn) {
    bs := make([]byte, 1024 * 30)
    common.Try(func() {
        for {
            len, e3 := conn.Read(bs)
            if e3 == nil {
                fmt.Println(string(bs[0:len]))
            } else {
                if e3 == io.EOF {
                    logger.Info("read all bytes from client")
                } else {
                    logger.Error(e3)
                }
                break
            }
        }
    }, func(i interface{}) {
        logger.Error("read from client occurs errors:", i)
    })
    conn.Close()
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