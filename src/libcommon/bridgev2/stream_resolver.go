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
    "hash"
)

var bytesPool *pool.BytesPool

var ErrOperationNotSupport = errors.New("operation not support")
var ErrSendHeadBytes = errors.New("error send head bytes")
var ErrSendBodyBytes = errors.New("error send head bytes")
var ErrRead = errors.New("error read bytes")
var ErrFileNotFound = errors.New("file not found")
var ErrInternalServer = errors.New("internal server error")
var ErrDownloadFile = errors.New("error download file")

const FrameHeadSize = 19

func init() {
    bytesPool = pool.NewBytesPool(100)
}

// common bytes reader, if error occurs, it will close automatically
func ReadBytes(buff []byte, len int, manager *ConnectionManager, md hash.Hash) (int, error) {
    var read = 0
    for {
        if read >= len {
            break
        }
        l, e := manager.Conn.Read(buff[read:len])
        if l == 0 || e == io.EOF {
            manager.Destroy()
            return 0, ErrRead
        }
        if l <= len {
            read += l
            continue
        }
    }
    if md != nil {
        _, e1 := md.Write(buff[0:len])
        if e1 != nil {
            return 0, e1
        }
    }
    app.UpdateIOIN(int64(len))
    return len, nil
}



// read frame from server/client.
// if frame containers body, then should read it later in custom handlers.
func readFrame(manager *ConnectionManager) (*Frame, error) {
    headerBytes, _ := MakeBytes(FrameHeadSize, false, 0, false) // meta header size
    defer RecycleBytes(headerBytes)
    // read header meta data
    _, e := ReadBytes(headerBytes, FrameHeadSize, manager, manager.Md)
    if e == nil {
        // read meta and body size
        bMetaSize := headerBytes[3:11]
        bBodySize := headerBytes[11:19]
        metaLength := binary.BigEndian.Uint64(bMetaSize)
        bodyLength := binary.BigEndian.Uint64(bBodySize)
        metaBodyBytes, e1 := readFrameMeta(int(metaLength), manager)
        if e1 != nil {
            return nil, e1
        }
        frame := &Frame{
            FrameHead: headerBytes[0:2],
            FrameStatus: headerBytes[2],
            MetaLength: int(metaLength),
            BodyLength: int64(bodyLength),
            FrameMeta: metaBodyBytes,
        }
        if frame.FrameStatus != StatusSuccess && manager.Side == ClientSide {
            logger.Debug("server response error code " + strconv.Itoa(int(frame.FrameStatus)) + " ("+ TranslateResponseMsg(frame.FrameStatus) +")")
            if frame.FrameStatus != StatusSuccess {
                return nil, errors.New("server response error code " + strconv.Itoa(int(frame.FrameStatus)) + " ("+ TranslateResponseMsg(frame.FrameStatus) +")")
            }
        }
        // server socket need validated connection state before action
        if manager.Side == ServerSide && frame.GetOperation() != FrameOperationValidate {
            manager.RequireStatus(StateValidated)
        }
        return frame, nil
    }
    return nil, e
}


// write frame to server/client.
func writeFrame(manager *ConnectionManager, frame *Frame) error {
    // prepare frame meta
    tmpBuf, _ := MakeBytes(8, false, 0, false)
    defer RecycleBytes(tmpBuf)
    var headerBuff bytes.Buffer
    if frame.GetOperation() == FrameOperationNone {
        frame.SetOperation(FrameOperationNone)
    }
    headerBuff.Write(frame.FrameHead)
    headerBuff.WriteByte(frame.FrameStatus)
    metaLenBytes := common.ConvertLen2Bytes(int64(frame.MetaLength), &tmpBuf)
    headerBuff.Write(*metaLenBytes)
    bodyLenBytes := common.ConvertLen2Bytes(frame.BodyLength, &tmpBuf)
    headerBuff.Write(*bodyLenBytes)
    headerBuff.Write(frame.FrameMeta)

    bs := headerBuff.Bytes()
    // write frame meta
    len1, e2 := manager.Conn.Write(bs)
    if e2 != nil {
        manager.Destroy()
        return e2
    }
    if len1 != headerBuff.Len() {
        manager.Destroy()
        return ErrSendHeadBytes
    }
    app.UpdateIOOUT(int64(headerBuff.Len()))
    bodyWriterHandler := frame.BodyWriterHandler
    if frame.BodyLength > 0 && bodyWriterHandler != nil {
        return bodyWriterHandler(manager, frame)
    }
    return nil
}

// 读取meta字节信息
func readFrameMeta(metaSize int, manager *ConnectionManager) ([]byte, error) {
    tmp, me := MakeBytes(int64(metaSize), true, 5242880, true) // 5MB
    if me != nil {
        return nil, me
    }
    len, e := ReadBytes(tmp, metaSize, manager, manager.Md)
    if e != nil && e != io.EOF {
        return nil, e
    }
    if (e == nil || e == io.EOF) && len == metaSize {
        return tmp, nil
    }
    // should never happen, mark as broken connection
    return nil, ErrRead
}


// dynamic: whether apply size is dynamic or unknown before,
// and we don't cache those.
func MakeBytes(len int64, dangerCheck bool, max int64, dynamic bool) ([]byte, error) {
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



