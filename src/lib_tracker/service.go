package lib_tracker

import (
    "util/logger"
    "strconv"
    "net"
    "util/pool"
    "util/common"
    "lib_common"
    "time"
    "lib_common/header"
    "encoding/json"
    "regexp"
    "validate"
    "strings"
)

var p, _ = pool.NewPool(1000, 100000)


var secret string

// Start service and listen
// 1. Start task for upload listen
// 2. Start task for communication with tracker
func StartService(config map[string] string) {
    port := config["port"]
    secret = config["secret"]
    go ExpirationDetection()
    startTrackerService(port)
}

// Tracker server start listen
func startTrackerService(port string) {
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
                                registerHandler(conn)
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



// accept a new connection for file upload
// the connection will keep till it is broken
func registerHandler(conn net.Conn) {

    defer func() {
        logger.Debug("close connection from server")
        conn.Close()
    }()

    common.Try(func() {
        // body buff
        // calculate md5
        for {
            // read meta
            operation, meta, _, err := lib_common.ReadConnMeta(conn)
            if operation == -1 || meta == "" || err != nil {
                // otherwise mark as broken connection
                if err != nil {
                    logger.Error(err)
                }
                break
            }

            // check secret
            checkStatus, commuMeta := checkRegisterMeta(meta, conn)
            // if secret validate failed or meta parse error
            if checkStatus != 0 {
                break
            }

            // register storage client
            if operation == 0 {
                valid := true
                //check meta fields
                if mat, _ := regexp.Match(validate.GroupInstancePattern, []byte(commuMeta.Group)); !mat {
                    logger.Error("register failed: group or instance_id is invalid")
                    valid = false
                }
                if commuMeta.Port < 1 || commuMeta.Port > 65535 || commuMeta.InstanceId == "" {
                    logger.Error("register failed: error parameter")
                    valid = false
                }
                remoteAddr := strings.Split(conn.RemoteAddr().String(), ":")[0]
                if commuMeta.BindAddr == "" {
                    logger.Warn("storage server not send bind address, using", remoteAddr)
                    commuMeta.BindAddr = remoteAddr
                }
                if !IsInstanceIdUnique(commuMeta) {
                    logger.Error("register failed: instance_id is not unique")
                    valid = false
                }
                if !valid {
                    var response = &header.CommunicationRegisterStorageResponseMeta {
                        Status: 3,
                    }
                    // write response close conn, and not check if success
                    lib_common.WriteResponse(4, conn, response)
                    //close conn
                    lib_common.Close(conn)
                    break
                }
                // validate success
                AddStorageServer(commuMeta)
                var response = &header.CommunicationRegisterStorageResponseMeta {
                    Status: 0,
                    LookBackAddr: remoteAddr,
                    GroupMembers: GetGroupMembers(commuMeta),
                }
                // write response close conn, and not check if success
                e1 := lib_common.WriteResponse(4, conn, response)
                if e1 != nil{
                    lib_common.Close(conn)
                    break
                }
            }

        }
    }, func(i interface{}) {
        logger.Error("connection error:", i)
    })
}




// 处理注册storage
func checkRegisterMeta(meta string, conn net.Conn) (int, *header.CommunicationRegisterStorageRequestMeta) {
    headerMeta := &header.CommunicationRegisterStorageRequestMeta{}
    e2 := json.Unmarshal([]byte(meta), &headerMeta)
    if e2 == nil {
        if headerMeta.Secret == secret {
            return 0, headerMeta // success
        } else {
            var response = &header.CommunicationRegisterStorageResponseMeta {
                Status: 1,
            }
            // write response close conn, and not check if success
            lib_common.WriteResponse(4, conn, response)
            //close conn
            lib_common.Close(conn)
            logger.Error("error check secret")
            return 1, headerMeta // bad secret
        }
    } else {
        //close conn
        lib_common.Close(conn)
        return 2, nil // parse meta error
    }
}

