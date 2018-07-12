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
    "io"
    "util/file"
    "bytes"
    "common/header"
    "encoding/binary"
    "encoding/json"
    "time"
    "crypto/md5"
    "encoding/hex"
    "hash"
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
    var headerSize = 18
    var bodyBuffSize = 1024*30
    defer func() {
        conn.Close()
        logger.Info("Close connection/////")
    }()
    common.Try(func() {
        bs := make([]byte, headerSize)  // meta header size
        bodybuff := make([]byte, bodyBuffSize)     // body buff
        index := 0 //test
        md := md5.New()
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
                } else if bytes.Compare(op, header.COM_UPLOAD_FILE) == 0 {// 山川文件
                    operation = 3
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
                logger.Info("upload meta: ", meta)
                if operation == 1 {
                    code := handleUploadFile(meta, conn)
                    if code != 0 {
                        conn.Close()
                        // 终止循环
                        break
                    }
                    continue
                }
                index++
                // begin upload file
                logger.Info("开始上传文件，文件大小：", bodySize/1024, "KB")
                fi, _ := file.CreateFile("D:\\godfs\\nginx-1.8.1("+ strconv.Itoa(index) +").zip")
                println(&bodybuff)
                e4 := readBodyBytes(bodySize, bodyBuffSize, bodybuff, fi, conn, md)
                fi.Close()
                if e4 != nil {
                    logger.Error("delete file")
                    logger.Error(e4)
                    file.Delete(fi.Name())
                    break
                }
            } else {
                logger.Error("read header failed")
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
func readBodyBytes(bodySize uint64, bodyBuffSize int, bodybuff []byte, out io.Writer, conn net.Conn, md hash.Hash) error {
    println(&bodybuff)
    var readBodySize uint64 = 0
    var nextReadSize int
    for {
        //read finish
        if readBodySize == bodySize {
            cipherStr := md.Sum(nil)
            logger.Info("上传结束，读取字节：", readBodySize, " MD5= " , hex.EncodeToString(cipherStr))
            md.Reset()

            /*resp := &header.UploadResponseMeta{}
            resp.Status = 0
            resp.Path = ""
            conn.Write()*/

            return nil
        }
        if (bodySize - readBodySize) / uint64(bodyBuffSize) >= 1 {
            nextReadSize = int(bodyBuffSize)
        } else {
            nextReadSize = int(bodySize - readBodySize)
        }
        len, e3 := readBytes(bodybuff, nextReadSize, conn)
        logger.Trace("read bytes...")
        if e3 == nil && len > 0 {
            readBodySize += uint64(len)
            lenn, e1 := out.Write(bodybuff[0:len])
            md.Write(bodybuff[0:len])
            logger.Trace("everything is ok, write", lenn)
            if e1 != nil {
                logger.Info(lenn)
                conn.Close()
                return e1
            }
        } else {
            conn.Close()
            logger.Error(e3)
            // 终止循环
            return e3
        }
    }
}



// 处理注册storage
func handleUploadFile(meta string, conn net.Conn) int {
    headerMeta := &header.UploadRequestMeta{}
    e2 := json.Unmarshal([]byte(meta), &headerMeta)
    if e2 == nil {
        if headerMeta.Secret == secret {
            return 0 // success
        } else {
            respMeta := &header.UploadResponseMeta{}
            respMeta.Status = 1
            resp, _ := json.Marshal(respMeta)
            conn.Write(resp)
            return 1 // bad secret
        }
    } else {
        return 2 // parse meta error
    }
}

