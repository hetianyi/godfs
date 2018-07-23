package bridge

import (
    "errors"
    "encoding/json"
    "encoding/binary"
    "net"
    "io"
    "bytes"
    "strconv"
    "app"
    "util/logger"
)

// operation codes const.
const (
    O_CONNECT = 1
    O_RESPONSE = 2

    O_UPLOAD = 2
    O_QUERY_FILE = 3
    O_DOWNLOAD_FILE = 4
    O_REG_STORAGE = 5

)

const HeaderSize = 18
// store operation code mapped with 2 first head bytes.
var operationHeadMap = make(map[int][]byte)

// error indicate that the operation is not support.
var OPERATION_NOT_SUPPORT_ERROR = errors.New("operation not support")
var SEND_HEAD_BYTES_ERROR = errors.New("error send head bytes")
var SEND_BODY_BYTES_ERROR = errors.New("error send head bytes")
var READ_ERROR = errors.New("error read bytes")

// init operations
func init() {
    operationHeadMap[O_CONNECT] = []byte{1,1}
    operationHeadMap[O_RESPONSE] = []byte{1,2}
    operationHeadMap[O_UPLOAD] = []byte{1,3}
    operationHeadMap[O_QUERY_FILE] = []byte{1,4}
    operationHeadMap[O_DOWNLOAD_FILE] = []byte{1,5}
    operationHeadMap[O_REG_STORAGE] = []byte{1,6}
}

// SendReceiveCloser
type SendReceiver interface {
    // client send request.
    SendRequest(operation int, meta interface{}, bodyLen uint64, bodyWriterHandler func(out io.WriteCloser) error) error
    // server receive request, it will block till get a new request from client.
    ReceiveRequest(requestHandler func(request *Meta, in io.ReadCloser) error) error
    // server send response.
    SendResponse(meta interface{}, bodyLen uint64, bodyWriterHandler func(out io.WriteCloser) error) error
    // client receive response from server.
    ReceiveResponse(responseHandler func(response *Meta, in io.ReadCloser) error) error
    // send validate request
    ValidateConnection() error
    GetConn() net.Conn
}

// include a tcp interact request meta data:
// operation bytes, meta length bytes, body length bytes and meta body bytes.
type Meta struct {
    Operation int
    metaLength uint64
    BodyLength uint64
    MetaBody []byte
    Err error
}


type Bridge struct {
    connection net.Conn
}

func (bridge *Bridge) Close() {
    Close(bridge.connection)
}

func (bridge *Bridge) GetConn() net.Conn {
    return bridge.connection
}


func (bridge *Bridge) SendRequest(operation int, meta interface{}, bodyLen uint64, bodyWriterHandler func(out io.WriteCloser) error) error {
    request, e2 := CreateMeta(operation, meta, bodyLen)
    if e2 != nil {
        return e2
    }
    metaLenBytes := convertLen2Bytes(request.metaLength)
    bodyLenBytes := convertLen2Bytes(request.BodyLength)

    var headerBuff bytes.Buffer
    headerBuff.Write(operationHeadMap[request.Operation])
    headerBuff.Write(metaLenBytes)
    headerBuff.Write(bodyLenBytes)
    headerBuff.Write(request.MetaBody)
    len1, e1 := bridge.connection.Write(headerBuff.Bytes())
    if e1 != nil {
        Close(bridge.connection)
        return e1
    }
    if len1 != headerBuff.Len() {
        Close(bridge.connection)
        return SEND_HEAD_BYTES_ERROR
    }
    if request.BodyLength > 0 {
        // write request body bytes using custom writer handler.
        err := bodyWriterHandler(bridge.connection)
        if err != nil {
            Close(bridge.connection)
            return err
        }
    }
    return nil
}

func (bridge *Bridge) ReceiveResponse(responseHandler func(response *Meta, in io.Reader) error) error {
    operation, metaSize, bodySize, metaBodyBytes, e1 := readHeadBytes(bridge.connection)
    if e1 != nil {
        return e1
    }
    response := &Meta{
        Operation: operation,
        metaLength: metaSize,
        BodyLength: bodySize,
        MetaBody: metaBodyBytes,
    }
    return responseHandler(response, bridge.connection)
}

func (bridge *Bridge) ReceiveRequest(requestHandler func(request *Meta, in io.ReadCloser) error) error {
    operation, _, bodySize, metaBodyBytes, e1 := readHeadBytes(bridge.connection)
    if e1 != nil {
        return e1
    }
    request := &Meta{
        Operation: operation,
        BodyLength: bodySize,
        MetaBody: metaBodyBytes,
    }
    return requestHandler(request, bridge.connection)
}

func (bridge *Bridge) SendResponse(meta interface{}, bodyLen uint64, bodyWriterHandler func(out io.WriteCloser) error) error {
    response, e2 := CreateMeta(O_RESPONSE, meta, bodyLen)
    if e2 != nil {
        return e2
    }
    logger.Debug(string(response.MetaBody))
    metaLenBytes := convertLen2Bytes(response.metaLength)
    bodyLenBytes := convertLen2Bytes(response.BodyLength)
    var headerBuff bytes.Buffer
    headerBuff.Write(operationHeadMap[response.Operation])
    headerBuff.Write(metaLenBytes)
    headerBuff.Write(bodyLenBytes)
    headerBuff.Write(response.MetaBody)
    len1, e1 := bridge.connection.Write(headerBuff.Bytes())
    if e1 != nil {
        Close(bridge.connection)
        return e1
    }
    if len1 != headerBuff.Len() {
        Close(bridge.connection)
        return SEND_HEAD_BYTES_ERROR
    }
    if response.BodyLength > 0 {
        // write request body bytes using custom writer handler.
        err := bodyWriterHandler(bridge.connection)
        if err != nil {
            Close(bridge.connection)
            return err
        }
    }
    return nil
}

func (bridge *Bridge) ValidateConnection(secret string) error {
    var sec = app.SECRET
    if secret != "" {
        sec = secret
    }
    validateMeta := &OperationValidationRequest {
        Secret: sec,
    }
    // send validate request
    e1 := bridge.SendRequest(O_CONNECT, validateMeta, 0, nil)
    if e1 != nil {
        return e1
    }
    e4 := bridge.ReceiveResponse(func(response *Meta, in io.Reader) error {
        if response.Err != nil {
            return response.Err
        }
        var validateResp = &OperationValidationResponse{}
        logger.Debug(string(response.MetaBody))
        e3 := json.Unmarshal(response.MetaBody, validateResp)
        if e3 != nil {
            return e3
        }
        if validateResp.Status != 0 {
            return errors.New("error connect to server, server response status:" + strconv.Itoa(validateResp.Status))
        }
        // connect success
        return nil
    })
    if e4 != nil {
        return e4
    }
    return nil
}




func NewBridge(conn net.Conn) *Bridge {
    return &Bridge{connection: conn}
}



func Close(closer io.Closer) error {
    if closer != nil {
        return closer.Close()
    }
    return nil
}

func convertLen2Bytes(len uint64) []byte {
    bodyLenBytes := make([]byte, 8)
    binary.BigEndian.PutUint64(bodyLenBytes, uint64(len))
    return bodyLenBytes
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
            return 0, READ_ERROR
        }
        if l <= len {
            read += l
            continue
        }
    }
    return len, nil
}

// read 18 head bytes.
func readHeadBytes(reader io.ReadCloser) (int, uint64, uint64, []byte, error) {
    headerBytes := make([]byte, HeaderSize)  // meta header size
    // read header meta data
    len, e := ReadBytes(headerBytes, HeaderSize, reader)
    if e == nil && len == HeaderSize {
        operation := retrieveOperation(headerBytes[0:2])
        // read meta and body size
        bMetaSize := headerBytes[2:10]
        bBodySize := headerBytes[10:18]
        metaSize := binary.BigEndian.Uint64(bMetaSize)
        bodySize := binary.BigEndian.Uint64(bBodySize)
        metaBodyBytes, e1 := readMetaBytes(int(metaSize), reader)
        if e1 != nil {
            return 0, 0, 0, nil, e1
        }
        return operation, metaSize, bodySize, metaBodyBytes, nil
    }
    return 0, 0, 0, nil, READ_ERROR
}


// 读取meta字节信息
func readMetaBytes(metaSize int, reader io.ReadCloser) ([]byte, error) {
    tmp := make([]byte, metaSize)
    len, e := ReadBytes(tmp, metaSize, reader)
    if e != nil && e != io.EOF {
        return nil, e
    }
    if (e == nil || e == io.EOF) && len == metaSize {
        return tmp, nil
    }
    //should never happen, mark as broken connection
    return nil, READ_ERROR
}

// retrieve operation code from operation head bytes.
// return 0 if no operation code matches.
func retrieveOperation(op []byte) int {
    for k, v := range operationHeadMap {
        if bytes.Compare(op, v) == 0 {
            return k
        }
    }
    return 0
}



// create a new tcp request using given data
// operation: operation code, such as 'O_CONNECT'
// meta     : meta object
// bodyLen  : request body length
// if create success, it returns a *Request, or else returns with error
func CreateMeta(operation int, meta interface{}, bodyLen uint64) (*Meta, error) {
    // operation bytes not found
    if operationHeadMap[operation] == nil {
        return nil, OPERATION_NOT_SUPPORT_ERROR
    }

    metaBodyBytes, e := json.Marshal(meta)
    if e != nil {
        return nil, e
    }

    metaLen := uint64(len(metaBodyBytes))

    return &Meta{
        operation,
        metaLen,
        bodyLen,
        metaBodyBytes,
        nil,
    }, nil
}


