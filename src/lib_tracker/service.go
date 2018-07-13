package lib_tracker

import (
    "util/logger"
    "regexp"
    "strconv"
    "common"
    "net"
    "util/pool"
    "io"
    "common/header"
    "bytes"
    "encoding/binary"
    "encoding/json"
    "strings"
    "errors"
    "util/file"
    "lib_common"
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
    pt := parsePort(port)
    if pt > 0 {
        tryTimes := 0
        for {
            lib_common.Try(func() {
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
                                registerHandler(conn)
                            })
                        } else {
                            logger.Info("accept new conn error", e1)
                        }
                    }
                }
            }, func(i interface{}) {
                logger.Error("["+ strconv.Itoa(tryTimes) +"] error shutdown service duo to:", i)
            })
        }
    }
}



// storage首次连接注册的处理器
func registerHandler(conn net.Conn) {
    var headerSize = 18
    var bodyBuffSize = 1024*30
    defer conn.Close()
    lib_common.Try(func() {
        bs := make([]byte, headerSize)  // meta header size
        bodybuff := make([]byte, bodyBuffSize)     // body buff
        for {
            // read header meta data
            len, e3 := readBytes(bs, headerSize, conn)
            var operation int
            var metaSize uint64
            var bodySize uint64
            if e3 == nil && len == headerSize {
                op := bs[0:2]
                if bytes.Compare(op, header.COM_REG_FILE) == 0 {// 注册storage
                    operation = 1
                } else if bytes.Compare(op, header.COM_REG_FILE) == 0 {// 注册文件
                    operation = 2
                } else {
                    logger.Info("operation not support yet")
                    // 终止循环
                    break
                }
                b_metaSize := bs[2:10]
                b_bodySize := bs[10:18]
                metaSize = binary.BigEndian.Uint64(b_metaSize)
                bodySize = binary.BigEndian.Uint64(b_bodySize)

                //TODO limit meta size
                meta := readMetaBytes(metaSize, conn)
                if operation == 1 {
                    code := handleRegisterStorage(meta, conn)
                    if code != 0 {
                        conn.Close()
                        // 终止循环
                        break
                    }
                    continue
                }
                fi, _ := file.CreateFile("D:\\godfs\\upload.zip")
                e4 := readBodyBytes(bodySize, bodyBuffSize, bodybuff, fi, conn)
                if e4 != nil {
                    fi.Close()
                    file.Delete(fi.Name())
                    break
                }
            } else {
                if e3 == io.EOF {
                    logger.Error("error EOF from client(1)")
                } else {
                    logger.Error(e3)
                }
                // 终止循环
                break
            }
        }


    }, func(i interface{}) {
        logger.Error("read from client occurs errors:", i)
    })
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

// 通用字节读取函数，如果读取结束/失败自动关闭连接
func readBytes(buff []byte, len int, conn net.Conn) (int, error) {
    len, e := conn.Read(buff[0:len])
    if len == 0 || e == io.EOF {
        defer conn.Close()
    }
    return len, e
}

// 读取meta字节信息
func readMetaBytes(metaSize uint64, conn net.Conn) string {
    tmp := make([]byte, metaSize)
    len, e := conn.Read(tmp)
    if e == nil || e == io.EOF {
        return string(tmp[0:len])
    }
    return ""
}

// 读取body
func readBodyBytes(bodySize uint64, bodyBuffSize int, bodybuff []byte, out io.Writer, conn net.Conn) error {
    var readBodySize uint64 = 0
    var nextReadSize int
    for {
        //read finish
        if readBodySize == bodySize {
            return nil
        }
        if (bodySize - readBodySize) / uint64(bodyBuffSize) >= 1 {
            nextReadSize = int(bodyBuffSize)
        } else {
            nextReadSize = int(bodySize - readBodySize)
        }
        len, e3 := readBytes(bodybuff, nextReadSize, conn)
        if e3 != nil && len > 0 {
            readBodySize += uint64(nextReadSize)
            out.Write(bodybuff[0:len])
        } else {
            conn.Close()
            if e3 == io.EOF {
                logger.Info("read all bytes from client")
            } else {
                logger.Error(e3)
            }
            // 终止循环
            return errors.New("error read from client(1)")
        }
    }
}





// 处理注册storage
func handleRegisterStorage(meta string, conn net.Conn) int {
    headerMeta := &header.CommunicationRegisterStorageRequestMeta{}
    e2 := json.Unmarshal([]byte(meta), &headerMeta)
    if e2 == nil {
        if headerMeta.Secret == secret {
            addr := strings.TrimSpace(headerMeta.BindAddr)
            if addr == "" {
                addr = conn.RemoteAddr().String()
            }
            //headerMeta.Port
            AddStorageServer(addr, headerMeta.Port, headerMeta.Group)
            // response the storage server
            respMeta := &header.CommunicationRegisterStorageResponseMeta{}
            respMeta.Status = 0
            respMeta.LookBackAddr = conn.RemoteAddr().String()
            respMeta.GroupMembers = GetGroupMembers(addr, headerMeta.Port, headerMeta.Group)
            resp, _ := json.Marshal(respMeta)
            conn.Write(resp)
            return 0 // success
        } else {
            respMeta := &header.CommunicationRegisterStorageResponseMeta{}
            respMeta.Status = 1
            resp, _ := json.Marshal(respMeta)
            conn.Write(resp)
            return 1 // bad secret
        }
    } else {
        return 2 // parse meta error
    }
}

