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
    "errors"
    "reflect"
    "lib_service"
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

    // anyway defer close conn
    defer func() {
        logger.Debug("close connection from server")
        lib_common.Close(conn)
    }()

    common.Try(func() {
        // 连接的时候客户端要一次性表明身份信息
        // 首次连接的operation必须是注册行为
        operation, meta, ferr := checkOnceOnConnect(conn)
        if operation == -1 {
            logger.Error(ferr)
            return
        }
        if operation == 0 {
            var tMeta  = meta.(header.CommunicationRegisterStorageRequestMeta)
            defer FutureExpireStorageServer(&tMeta)
        }

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
            // register file from client
            if operation == 1 {
                ee := handleRegisterFile(meta, conn)
                if ee != nil {
                    logger.Error(ee)
                    break
                }
            } else {
                logger.Error("unSupport operation")
                lib_common.Close(conn)
                break
            }
        }
    }, func(i interface{}) {
        logger.Error("connection error:", i)
    })
}




// 处理注册storage
func checkMetaSecret(meta string, metaType interface{}, conn net.Conn) (int, interface{}) {
    e2 := json.Unmarshal([]byte(meta), &metaType)
    s := reflect.ValueOf(&metaType).Elem()
    f := s.FieldByName("Secret")
    metaSecret := f.Interface().(string)
    if e2 == nil {
        if metaSecret == secret {
            return 0, metaType // success
        } else {
            var response = &header.CommunicationRegisterStorageResponseMeta {
                Status: 1,
            }
            // write response close conn, and not check if success
            lib_common.WriteResponse(4, conn, response)
            //close conn
            lib_common.Close(conn)
            logger.Error("error check secret")
            return 1, nil // bad secret
        }
    } else {
        //close conn
        lib_common.Close(conn)
        return 2, nil // parse meta error
    }
}

// 首次的时候检查客户端
// return operation, error
func checkOnceOnConnect(conn net.Conn) (int, interface{}, error) {
    // read meta
    operation, meta, _, err := lib_common.ReadConnMeta(conn)
    // TODO maybe add one more operation for upload client
    if meta == "" || err != nil {
        // otherwise mark as broken connection
        lib_common.Close(conn)
        if err != nil {
            return -1, nil, err
        }
        return -1, nil, errors.New("meta check failed")
    }

    var checkStatus int
    var tcommuMeta interface{}
    // check secret
    checkStatus, tcommuMeta = checkMetaSecret(meta, header.CommunicationRegisterStorageRequestMeta{}, conn)
    // if secret validate failed or meta parse error
    if checkStatus != 0 {
        return -1, nil, errors.New("secret check failed")
    }
    var commuMeta = tcommuMeta.(header.CommunicationRegisterStorageRequestMeta)

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
        if !IsInstanceIdUnique(&commuMeta) {
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
            return -1, nil, errors.New("invalid meta data")
        }
        // validate success
        AddStorageServer(&commuMeta)
        var response = &header.CommunicationRegisterStorageResponseMeta {
            Status: 0,
            LookBackAddr: remoteAddr,
            GroupMembers: GetGroupMembers(&commuMeta),
        }
        // write response close conn, and not check if success
        e1 := lib_common.WriteResponse(4, conn, response)
        if e1 != nil{
            FutureExpireStorageServer(&commuMeta)
            lib_common.Close(conn)
            return -1, nil, e1
        }
        return operation, commuMeta, nil
    } else {
        lib_common.Close(conn)
        return -1, nil, errors.New("operation not support")
    }
}


func handleRegisterFile(meta string, conn net.Conn) error {
    var metaEntity = &header.CommunicationRegisterFileRequestMeta{}
    e1 := json.Unmarshal([]byte(meta), metaEntity)
    if e1 != nil {
        lib_common.Close(conn)
        return e1
    }
    e2 := lib_service.TrackerAddFile(metaEntity)
    if e2 != nil {
        lib_common.Close(conn)
        return e2
    }
    return nil
}
