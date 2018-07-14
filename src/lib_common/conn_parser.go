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
)

const HeaderSize = 18
const BodyBuffSize = 1024*30

// 解析连接传输数据Meta
// operation : 请求操作，0：不支持的操作，1：注册storage，2：注册文件，3：上传文件
// meta      : 请求头信息
// err       : 如果发生错误，返回值为operation=-1, meta="", e
func ParseConnRequestMeta(conn net.Conn) (operation int, meta string, bodySize uint64, err error) {
    operation = -1
    headerBytes := make([]byte, HeaderSize)  // meta header size
    // read header meta data
    len, e := ReadBytes(headerBytes, HeaderSize, conn)
    if e == nil && len == HeaderSize {
        op := headerBytes[0:2]
        if bytes.Compare(op, header.COM_REG_FILE) == 0 {// 注册storage
            operation = 1
        } else if bytes.Compare(op, header.COM_REG_FILE) == 0 {// 注册文件
            operation = 2
        } else if bytes.Compare(op, header.COM_UPLOAD_FILE) == 0 {// 上传文件
            operation = 3
        } else {
            // otherwise mark as broken connection
            return 0, "", 0, nil
        }

        // read meta and body size
        bMetaSize := headerBytes[2:10]
        bBodySize := headerBytes[10:18]
        metaSize := binary.BigEndian.Uint64(bMetaSize)
        bodySize := binary.BigEndian.Uint64(bBodySize)

        // read meta bytes
        metaStr, e1 := readMetaBytes(metaSize, conn)
        if e1 != nil {
            // otherwise mark as broken connection
            Close(conn)
            return -1, "", 0, e
        }
        return operation, metaStr, bodySize, nil
        // body read in the future
    } else {
        // otherwise mark as broken connection
        Close(conn)
        return -1, "", 0, e
    }
}

// 解析连接传输数据Body
// operation : 请求操作，0：不支持的操作，1：注册storage，2：注册文件，3：上传文件
// meta      : 请求头信息
// err       : 如果发生错误，返回值为operation=-1, meta="", e

func ParseConnRequestBody(bodySize uint64, buffer []byte, conn net.Conn, out io.WriteCloser, md hash.Hash) error {
    defer func() {
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
            logger.Info("上传结束，读取字节：", readBodySize, " MD5= " , hex.EncodeToString(cipherStr))
            return nil
        }
        // left bytes is more than a buffer
        if (bodySize - readBodySize) / uint64(BodyBuffSize) >= 1 {
            nextReadSize = int(BodyBuffSize)
        } else {// left bytes less than a buffer
            nextReadSize = int(bodySize - readBodySize)
        }
        len1, e3 := readBytes(buffer, nextReadSize, conn)
        if e3 == nil && len1 == nextReadSize {
            readBodySize += uint64(len1)
            len2, e1 := out.Write(buffer[0:len1])
            len3, e2 := md.Write(buffer[0:len1])
            // write error
            if e1 != nil || e2 != nil || len2 != len1 || len3 != len1 {
                logger.Error("write out error:", e1, "|", e2)
                Close(conn)
                return errors.New("write out error(0)")
            }
        } else {
            logger.Error("error read body")
            Close(conn)
            // 终止循环
            return e3
        }
    }
}


// 通用字节读取函数，如果读取结束/失败自动关闭连接
func readBytes(buff []byte, len int, conn net.Conn) (int, error) {
    len, e := conn.Read(buff[0:len])
    if len == 0 || e == io.EOF {
        Close(conn)
    }
    return len, e
}

// close connection
func Close(conn net.Conn) {
    conn.Close()
}

// 通用字节读取函数，如果读取结束/失败自动关闭连接
func ReadBytes(buff []byte, len int, conn net.Conn) (int, error) {
    len, e := conn.Read(buff[0:len])
    if len == 0 || e == io.EOF {
        defer conn.Close()
    }
    return len, e
}

// 读取meta字节信息
func readMetaBytes(metaSize uint64, conn net.Conn) (string, error) {
    tmp := make([]byte, metaSize)
    len, e := conn.Read(tmp)
    if e != nil && e != io.EOF {
        return "", e
    }
    if (e == nil || e == io.EOF) && uint64(len) == metaSize {
        return string(tmp[:]), nil
    }
    //should never happen, mark as broken connection
    return "", errors.New("error read meta bytes")
}