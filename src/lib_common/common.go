package lib_common

import (
    "net"
    "io"
    "bytes"
    "lib_common/header"
    "encoding/binary"
    "errors"
    "util/logger"
    "encoding/hex"
    "hash"
    "regexp"
    "strconv"
    "encoding/json"
    "crypto/md5"
    "util/timeutil"
    "util/file"
    "app"
    "os"
    "lib_service"
    "container/list"
    "strings"
)

const HeaderSize = 18
const BodyBuffSize = 1024*30

// 解析连接传输数据Meta
// operation : 请求操作，1：注册storage，2：注册文件，3：上传文件
// meta      : 请求头信息
// err       : 如果发生错误，返回值为operation=-1, meta="", e
func ReadConnMeta(conn net.Conn) (operation int, meta string, bodySize uint64, err error) {
    operation = -1
    headerBytes := make([]byte, HeaderSize)  // meta header size
    // read header meta data
    len, e := ReadBytes(headerBytes, HeaderSize, conn)
    if e == nil && len == HeaderSize {
        op := headerBytes[0:2]
        if bytes.Compare(op, header.OperationHeadByteMap[0]) == 0 {
            operation = 0
        } else if bytes.Compare(op, header.OperationHeadByteMap[1]) == 0 {
            operation = 1
        } else if bytes.Compare(op, header.OperationHeadByteMap[2]) == 0 {
            operation = 2
        } else if bytes.Compare(op, header.OperationHeadByteMap[3]) == 0 {
            operation = 3
        } else if bytes.Compare(op, header.OperationHeadByteMap[4]) == 0 {
            operation = 4
        } else if bytes.Compare(op, header.OperationHeadByteMap[5]) == 0 {
            operation = 5
        } else if bytes.Compare(op, header.OperationHeadByteMap[6]) == 0 {
            operation = 6
        } else if bytes.Compare(op, header.OperationHeadByteMap[8]) == 0 {
            operation = 8
        } else {
            logger.Debug("operation not support")
            var response = &header.UploadResponseMeta{
                Status: 2,
                Path: "",
            }
            // write response close conn, and not check if success
            WriteResponse(4, conn, response)
            // otherwise mark as broken connection
            Close(conn)
            return -1, "", 0, nil
        }

        // read meta and body size
        bMetaSize := headerBytes[2:10]
        bBodySize := headerBytes[10:18]
        metaSize := binary.BigEndian.Uint64(bMetaSize)
        bodySize := binary.BigEndian.Uint64(bBodySize)

        // read meta bytes
        metaStr, e1 := readMetaBytes(int(metaSize), conn)
        if e1 != nil {
            // otherwise mark as broken connection
            Close(conn)
            logger.Debug("operation not support")
            return -1, "", 0, e
        }
        return operation, metaStr, bodySize, nil
        // body read in the future
    } else {
        // otherwise mark as broken connection
        Close(conn)
        logger.Debug("read meta bytes failed")
        return -1, "", 0, e
    }
}

// 解析连接传输数据Body
// operation : 请求操作，0：不支持的操作，1：注册storage，2：注册文件，3：上传文件
// meta      : 请求头信息
// err       : 如果发生错误，返回值为operation=-1, meta="", e
func ReadConnBody(bodySize uint64, buffer []byte, conn io.ReadCloser, md hash.Hash) error {
    defer func() {
        md.Reset()
    }()
    out, oe := createTmpFile()
    if oe != nil {
        return oe
    }
    // total read bytes
    var readBodySize uint64 = 0
    // next time bytes to read
    var nextReadSize int
    var sliceReadSize int64 = 0
    var sliceMd5 = md5.New()
    var sliceIds list.List

    for {
        //read finish
        if readBodySize == bodySize {
            totalCipherStr := md.Sum(nil)
            sliceCipherStr := sliceMd5.Sum(nil)
            md5  := hex.EncodeToString(totalCipherStr)
            sMd5 := hex.EncodeToString(sliceCipherStr)
            out.Close()
            md.Reset()
            sliceMd5.Reset()

            e10 := moveTmpFileTo(sMd5, out)
            if e10 != nil {
                return e10
            }
            // save slice info to db
            pid, e8 := lib_service.AddPart(sMd5, sliceReadSize)
            if e8 != nil {
                Close(conn)
                return e8
            }
            sliceIds.PushBack(pid)
            logger.Info("上传结束，读取字节：", readBodySize, " MD5= " , md5)

            stoe := lib_service.StorageAddFile(md5, &sliceIds)
            if stoe != nil {
                Close(conn)
                return stoe
            }
            // mark the file is multi part or single part
            var path string
            if sliceIds.Len() > 1 {
                path = app.GROUP + "/" + app.INSTANCE_ID + "/M/" + md5
            } else {
                path = app.GROUP + "/" + app.INSTANCE_ID + "/S/" + md5
            }
            // upload success
            var response = &header.UploadResponseMeta{
                Status: 0,
                Path: path,
                Exist: true,
            }
            e5 := WriteResponse(4, conn, response)
            if e5 != nil {
                Close(conn)
                return e5
            }
            return nil
        }
        // left bytes is more than a buffer
        if (bodySize - readBodySize) / uint64(BodyBuffSize) >= 1 {
            nextReadSize = int(BodyBuffSize)
        } else {// left bytes less than a buffer
            nextReadSize = int(bodySize - readBodySize)
        }
        logger.Trace("read next bytes:", nextReadSize, "total is:", bodySize)
        len1, e3 := ReadBytes(buffer, nextReadSize, conn)
        if e3 == nil && len1 == nextReadSize {
            // if sliceReadSize > sliceSize then create a new slice file
            if sliceReadSize + int64(len1) > app.SLICE_SIZE {
                // write bytes to file
                leftN := app.SLICE_SIZE - sliceReadSize
                rightN := int64(len1) - (app.SLICE_SIZE - sliceReadSize)
                len2, e1 := out.Write(buffer[0:leftN])
                len4, e11 := sliceMd5.Write(buffer[0:leftN])
                if e1 != nil || e11 != nil || int64(len2) != leftN || int64(len4) != leftN {
                    logger.Error("write out error:", e1, "|", e11)
                    closeAndDeleteTmpFile(out)
                    Close(conn)
                    return errors.New("write out error(0)")
                }

                // close slice file and create a new slice file （承上启下）
                out.Close()
                sliceCipherStr := sliceMd5.Sum(nil)
                sMd5 := hex.EncodeToString(sliceCipherStr)
                sliceMd5.Reset()
                e10 := moveTmpFileTo(sMd5, out)
                if e10 != nil {
                    return e10
                }
                // save slice info to db
                pid, e8 := lib_service.AddPart(sMd5, app.SLICE_SIZE)
                if e8 != nil {
                    Close(conn)
                    return e8
                }
                sliceIds.PushBack(pid)

                out12, e12 := createTmpFile()
                if e12 != nil {
                    return e12
                }
                out = out12
                len6, e2 := out.Write(buffer[leftN:len1])
                len7, e12 := sliceMd5.Write(buffer[leftN:len1])
                if e2 != nil || e12 != nil || int64(len6) != rightN || int64(len7) != rightN {
                    logger.Error("write out error:", e2, "|", e12)
                    closeAndDeleteTmpFile(out)
                    Close(conn)
                    return errors.New("write out error(1)")
                }
                sliceReadSize = rightN
            } else {
                // write bytes to file
                len2, e1 := out.Write(buffer[0:len1])
                len3, e2 := md.Write(buffer[0:len1])
                len4, e3 := sliceMd5.Write(buffer[0:len1])
                // write error
                if e1 != nil || e2 != nil || e3 != nil || len2 != len1 || len3 != len1 || len4 != len1 {
                    logger.Error("write out error:", e1, "|", e2)
                    closeAndDeleteTmpFile(out)
                    Close(conn)
                    return errors.New("write out error(0)")
                }
                sliceReadSize += int64(len1)
            }
            readBodySize += uint64(len1)
        } else {
            logger.Error("error read body:", e3)
            closeAndDeleteTmpFile(out)
            Close(conn)
            // 终止循环
            return e3
        }
    }
}

func CreateTmpFile() (*os.File, error) {
    // begin upload file
    tmpFileName := timeutil.GetUUID()
    // using tmp ext and rename after upload success
    tmpPath := file.FixPath(app.BASE_PATH + "/data/tmp/" + tmpFileName)
    fi, e8 := file.CreateFile(tmpPath)
    if e8 != nil {
        return nil, e8
    }
    return fi, nil
}

func CloseAndDeleteTmpFile(fi *os.File) {
    fi.Close()
    file.Delete(fi.Name())
}

func MoveTmpFileTo(md5 string, fi *os.File) error {
    dig1 := strings.ToUpper(md5[0:2])
    dig2 := strings.ToUpper(md5[2:4])
    finalPath := app.BASE_PATH + "/data/" + dig1 + "/" + dig2
    if !file.Exists(finalPath) {
        e := file.CreateAllDir(finalPath)
        if e != nil {
            return e
        }
    }
    if !file.Exists(finalPath + "/" + md5) {
        eee := file.MoveFile(fi.Name(), finalPath + "/" + md5)
        if eee != nil {
            return eee
        }
    } else {
        s := file.Delete(fi.Name())
        if !s {
            logger.Error("error clean tmp file:", fi.Name())
        }
    }
    return nil
}


// 解析连接传输数据Body(用于客户端下载文件读取文件body)
// operation : 请求操作，0：不支持的操作，1：注册storage，2：注册文件，3：上传文件
// meta      : 请求头信息
// err       : 如果发生错误，返回值为operation=-1, meta="", e

func ReadConnDownloadBody(bodySize uint64, buffer []byte, in io.ReadWriteCloser, out io.WriteCloser) error {
    defer func() {
        logger.Debug("close out writer")
        out.Close()
    }()
    // total read bytes
    var readBodySize uint64 = 0
    // next time bytes to read
    var nextReadSize int
    for {
        //read finish
        if readBodySize == bodySize {
            logger.Debug("download finish, read bytes:", readBodySize)
            return nil
        }
        // left bytes is more than a buffer
        if (bodySize - readBodySize) / uint64(BodyBuffSize) >= 1 {
            nextReadSize = int(BodyBuffSize)
        } else {// left bytes less than a buffer
            nextReadSize = int(bodySize - readBodySize)
        }
        logger.Trace("read next bytes:", nextReadSize, "total is:", bodySize)
        len1, e3 := ReadBytes(buffer, nextReadSize, in)
        if e3 == nil && len1 == nextReadSize {
            readBodySize += uint64(len1)
            len2, e1 := out.Write(buffer[0:len1])
            // write error
            if e1 != nil || len2 != len1 {
                logger.Error("write out error:", e1)
                Close(in)
                out.Close()
                return errors.New("write out error(0)")
            }
        } else {
            Close(in)
            out.Close()
            // 终止循环
            return errors.New("error read body")
        }
    }
}


// close connection
func Close(conn io.ReadCloser) {
    conn.Close()
}

// 通用字节读取函数，如果读取结束/失败自动关闭连接
func ReadBytes(buff []byte, len int, conn io.ReadCloser) (int, error) {
     read := 0
    for {
        if read >= len {
            break
        }
        l, e := conn.Read(buff[read:len])
        if l == 0 || e == io.EOF {
            Close(conn)
            return 0, errors.New("error read bytes")
        }
        if l <= len {
            read += l
            continue
        }
    }
    return len, nil
}

// 读取meta字节信息
func readMetaBytes(metaSize int, conn net.Conn) (string, error) {
    tmp := make([]byte, metaSize)
    len, e := ReadBytes(tmp, metaSize, conn)
    if e != nil && e != io.EOF {
        return "", e
    }
    if (e == nil || e == io.EOF) && len == metaSize {
        return string(tmp[:]), nil
    }
    //should never happen, mark as broken connection
    return "", errors.New("error read meta bytes")
}

// 从string解析port端口，返回int类型的port端口
// 如果返回0表示port不合法
func ParsePort(port string) int {
    if len(port) < 1 {
        logger.Error("parameter 'port' not set yet, server will not exit now!")
        return 0
    }
    if b, _ := regexp.Match("^[1-9][0-9]{1,6}$", []byte(port)); b {
        p, e := strconv.Atoi(port)
        if e != nil || p > 65535 {
            logger.Error("parameter 'port' must be a valid port number!")
            return 0
        }
        return p
    }
    return 0
}

// generate meta header data using given meta
// returns (wrapped meta length bytes, body length bytes, meta json bytes, error)
func PrepareMetaData(bodySize int64, meta interface{}) ([]byte, []byte, []byte, error) {
    metaStr, e := json.Marshal(meta)
    if e != nil {
        return nil, nil, nil, e
    }

    metaSize := len(metaStr)
    metaSizeBytes := make([]byte, 8)
    bodyBytes := make([]byte, 8)

    binary.BigEndian.PutUint64(metaSizeBytes, uint64(metaSize))
    binary.BigEndian.PutUint64(bodyBytes, uint64(bodySize))
    return metaSizeBytes, bodyBytes, metaStr, nil
}


// write response
func WriteResponse(operation int, conn io.ReadWriteCloser, response interface{}) error {
    metaSize, bodySize, metaBytes, e := PrepareMetaData(0, response)
    if e != nil {
        return e
    }
    var buff bytes.Buffer
    buff.Write(header.OperationHeadByteMap[operation])
    buff.Write(metaSize)
    buff.Write(bodySize)
    buff.Write(metaBytes)
    len, e1 := conn.Write(buff.Bytes())
    if e1 != nil || len != buff.Len() {
        return errors.New("error write response")
    }
    return nil
}

func WriteMeta(operation int, metaSize []byte, bodySize []byte, meta []byte, conn net.Conn) error {
    var headerBuff bytes.Buffer
    headerBuff.Write(header.OperationHeadByteMap[operation])
    headerBuff.Write(metaSize)
    headerBuff.Write(bodySize)
    len1, e2 := conn.Write(headerBuff.Bytes())
    if e2 != nil || len1 != headerBuff.Len() {
        Close(conn)
        return errors.New("error write head bytes")
    }
    len2, e3 := conn.Write(meta)
    if e3 != nil || len2 != len(meta) {
        Close(conn)
        return errors.New("error write meta")
    }
    return nil
}

func TranslateResponseStatus(status int, conn net.Conn) error {
    if status == 1 {
        return nil
    } else if status == 1 {
        return errors.New("register storage to tracker server failed with error: bad secret|" + conn.RemoteAddr().String())
    } else if status == 2 {
        return errors.New("register storage to tracker server failed with error: operation not support|" + conn.RemoteAddr().String())
    } else if status == 3 {
        return errors.New("register storage to tracker server failed with error: server error|" + conn.RemoteAddr().String())
    }
    return errors.New("unknown error")
}

// 首次的时候检查客户端
// return error
func CheckOnceOnConnect(conn net.Conn) error {
    // read meta
    operation, meta, _, err := ReadConnMeta(conn)
    // TODO maybe add one more operation for upload client
    if operation != 8 || meta == "" || err != nil {
        // otherwise mark as broken connection
        Close(conn)
        if err != nil {
            return err
        }
        return errors.New("meta check failed")
    }
    // check secret
    s := checkMetaSecret(meta, conn)
    // if secret validate failed or meta parse error
    if !s {
        return errors.New("secret check failed")
    }
    return nil
}



// 处理注册storage
func checkMetaSecret(meta string, conn net.Conn) bool {
    var head = &header.ConnectionHead{}
    e2 := json.Unmarshal([]byte(meta), head)
    if e2 == nil {
        if head.Secret == app.SECRET {
            var response = &header.ConnectionHeadResponse {
                Status: 0,
            }
            // write response close conn, and not check if success
            e3 := WriteResponse(4, conn, response)
            if e3 != nil {
                //close conn
                Close(conn)
                return false
            }
            return true // success
        } else {
            var response = &header.ConnectionHeadResponse {
                Status: 1,
            }
            // write response close conn, and not check if success
            WriteResponse(4, conn, response)
            //close conn
            Close(conn)
            return false // bad secret
        }
    } else {
        //close conn
        Close(conn)
        return false // parse meta error
    }
}
