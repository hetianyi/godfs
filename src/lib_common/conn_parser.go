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

func ReadConnBody(bodySize uint64, buffer []byte, conn net.Conn, out io.WriteCloser, md hash.Hash) (string, error) {
    defer func() {
        logger.Debug("close out writer")
        out.Close()
        md.Reset()
    }()
    // total read bytes
    var readBodySize uint64 = 0
    // next time bytes to read
    var nextReadSize int
    for {
        //read finish
        if readBodySize == bodySize {
            cipherStr := md.Sum(nil)
            md5 := hex.EncodeToString(cipherStr)
            logger.Info("上传结束，读取字节：", readBodySize, " MD5= " , md5)
            return md5, nil
        }
        // left bytes is more than a buffer
        if (bodySize - readBodySize) / uint64(BodyBuffSize) >= 1 {
            nextReadSize = int(BodyBuffSize)
        } else {// left bytes less than a buffer
            nextReadSize = int(bodySize - readBodySize)
        }
        logger.Debug("read next bytes:", nextReadSize, "total is:", bodySize)
        len1, e3 := ReadBytes(buffer, nextReadSize, conn)
        if e3 == nil && len1 == nextReadSize {
            readBodySize += uint64(len1)
            len2, e1 := out.Write(buffer[0:len1])
            len3, e2 := md.Write(buffer[0:len1])
            // write error
            if e1 != nil || e2 != nil || len2 != len1 || len3 != len1 {
                logger.Error("write out error:", e1, "|", e2)
                Close(conn)
                return "", errors.New("write out error(0)")
            }
        } else {
            logger.Error("error read body:", e3)
            Close(conn)
            // 终止循环
            return "", e3
        }
    }
}


// 解析连接传输数据Body(用于客户端下载文件读取文件body)
// operation : 请求操作，0：不支持的操作，1：注册storage，2：注册文件，3：上传文件
// meta      : 请求头信息
// err       : 如果发生错误，返回值为operation=-1, meta="", e

func ReadConnDownloadBody(bodySize uint64, buffer []byte, conn net.Conn, out io.WriteCloser) error {
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
            logger.Info("下载结束，读取字节：", readBodySize)
            return nil
        }
        // left bytes is more than a buffer
        if (bodySize - readBodySize) / uint64(BodyBuffSize) >= 1 {
            nextReadSize = int(BodyBuffSize)
        } else {// left bytes less than a buffer
            nextReadSize = int(bodySize - readBodySize)
        }
        logger.Debug("read next bytes:", nextReadSize, "total is:", bodySize)
        len1, e3 := ReadBytes(buffer, nextReadSize, conn)
        if e3 == nil && len1 == nextReadSize {
            readBodySize += uint64(len1)
            len2, e1 := out.Write(buffer[0:len1])
            // write error
            if e1 != nil || len2 != len1 {
                logger.Error("write out error:", e1)
                Close(conn)
                return errors.New("write out error(0)")
            }
        } else {
            logger.Error("error read body:", e3)
            Close(conn)
            // 终止循环
            return e3
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
func WriteResponse(operation int, conn net.Conn, response interface{}) error {
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