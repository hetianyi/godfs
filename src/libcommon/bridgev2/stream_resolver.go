package bridgev2

import (
    "io"
    "app"
    "errors"
    "encoding/binary"
    "util/pool"
    "bytes"
    "util/common"
    "util/logger"
    "strconv"
)

var bytesPool *pool.BytesPool

var OPERATION_NOT_SUPPORT_ERROR = errors.New("operation not support")
var SEND_HEAD_BYTES_ERROR = errors.New("error send head bytes")
var SEND_BODY_BYTES_ERROR = errors.New("error send head bytes")
var READ_ERROR = errors.New("error read bytes")
var FILE_NOT_FOUND_ERROR = errors.New("file not found")
var INTERNAL_SERVER_ERROR = errors.New("internal server error")
var DOWNLOAD_FILE_ERROR = errors.New("error download file")

const FRAME_HEAD_SIZE = 19

func init() {
    bytesPool = pool.NewBytesPool(100)
}

// common bytes reader, if error occurs, it will close automatically
func ReadBytes(buff []byte, len int, manager *ConnectionManager) (int, error) {
    var read = 0
    for {
        if read >= len {
            break
        }
        l, e := manager.Conn.Read(buff[read:len])
        if l == 0 || e == io.EOF {
            manager.Close()
            return 0, READ_ERROR
        }
        if l <= len {
            read += l
            continue
        }
    }
    if manager.Md != nil {
        _, e1 := manager.Md.Write(buff[0:len])
        if e1 != nil {
            return 0, e1
        }
    }
    app.UpdateIOIN(int64(len))
    return len, nil
}



// read head bytes.
func readFrame(manager *ConnectionManager) (*Frame, error) {
    headerBytes, _ := MakeBytes(FRAME_HEAD_SIZE, false, 0, false) // meta header size
    defer RecycleBytes(headerBytes)
    // read header meta data
    len, e := ReadBytes(headerBytes, FRAME_HEAD_SIZE, manager)
    if e == nil && len == FRAME_HEAD_SIZE {
        // read meta and body size
        bMetaSize := headerBytes[2:10]
        bBodySize := headerBytes[10:18]
        metaLength := binary.BigEndian.Uint64(bMetaSize)
        bodyLength := binary.BigEndian.Uint64(bBodySize)
        metaBodyBytes, e1 := readFrameMeta(int(metaLength), manager)
        if e1 != nil {
            return nil, e1
        }
        frame := &Frame{
            frameHead: headerBytes[0:2],
            frameStatus: headerBytes[3],
            metaLength: int(metaLength),
            bodyLength: int64(bodyLength),
            frameMeta: metaBodyBytes,
        }

        if frame.frameStatus != SUCCESS && manager.Side == CLIENT_SIDE {
            logger.Debug("server response error code " + strconv.Itoa(int(frame.frameStatus)) + " ("+ TranslateResponseMsg(frame.frameStatus) +")")
            if frame.frameStatus != SUCCESS {
                return nil, errors.New("server response error code " + strconv.Itoa(int(frame.frameStatus)) + " ("+ TranslateResponseMsg(frame.frameStatus) +")")
            }
        }
        bodyReaderHandler := GetOperationHandler(frame.GetOperation())
        if bodyLength > 0 && bodyReaderHandler != nil && bodyReaderHandler.BodyReaderHandler != nil {
            return frame, bodyReaderHandler.BodyReaderHandler(frame, manager.Conn)
        }
        return frame, nil
    }
    return nil, READ_ERROR
}



func writeFrame(manager *ConnectionManager, frame *Frame) error {
    // prepare frame meta
    tmpBuf, _ := MakeBytes(8, false, 0, false)
    defer RecycleBytes(tmpBuf)
    var headerBuff bytes.Buffer
    headerBuff.Write(frame.frameHead)
    headerBuff.WriteByte(frame.frameStatus)
    metaLenBytes := common.ConvertLen2Bytes(int64(frame.metaLength), &tmpBuf)
    headerBuff.Write(*metaLenBytes)
    bodyLenBytes := common.ConvertLen2Bytes(frame.bodyLength, &tmpBuf)
    headerBuff.Write(*bodyLenBytes)
    headerBuff.Write(frame.frameMeta)
    // write frame meta
    len1, e2 := manager.Conn.Write(headerBuff.Bytes())
    if e2 != nil {
        manager.Close()
        return e2
    }
    if len1 != headerBuff.Len() {
        manager.Close()
        return SEND_HEAD_BYTES_ERROR
    }
    app.UpdateIOOUT(int64(headerBuff.Len()))
    bodyWriterHandler := GetOperationHandler(frame.GetOperation())
    if frame.bodyLength > 0 && bodyWriterHandler != nil && bodyWriterHandler.BodyReaderHandler != nil {
        return bodyWriterHandler.BodyReaderHandler(frame, manager.Conn)
    }
    return nil
}

// 读取meta字节信息
func readFrameMeta(metaSize int, manager *ConnectionManager) ([]byte, error) {
    tmp, me := MakeBytes(uint64(metaSize), true, 5242880, true) // 5MB
    if me != nil {
        return nil, me
    }
    len, e := ReadBytes(tmp, metaSize, manager)
    if e != nil && e != io.EOF {
        return nil, e
    }
    if (e == nil || e == io.EOF) && len == metaSize {
        return tmp, nil
    }
    // should never happen, mark as broken connection
    return nil, READ_ERROR
}


// dynamic: whether apply size is dynamic or unknown before,
// and we don't cache those.
func MakeBytes(len uint64, dangerCheck bool, max uint64, dynamic bool) ([]byte, error) {
    if dangerCheck && len > max {
        return nil, errors.New("cannot create bytes: system protection")
    }
    if !dynamic {
        return bytesPool.Apply(int(len)), nil
    }
    return make([]byte, len), nil
}

// recycle bytes buffer.
func RecycleBytes(buffer []byte) {
    bytesPool.Recycle(buffer)
}



