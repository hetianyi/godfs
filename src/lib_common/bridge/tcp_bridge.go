package bridge

import (
    "errors"
    "encoding/json"
    "encoding/binary"
    "net"
    "io"
)

// operation codes const.
const (
    O_CONNECT = 1
    O_RESPONSE = 2
)

// store operation code mapped with 2 first head bytes.
var operationHeadMap = make(map[int][]byte)

// error indicate that the operation is not support.
var OPERATION_NOT_SUPPORT = errors.New("operation not support")

// init operations
func init() {
    operationHeadMap[O_CONNECT] = []byte{1,1}
    operationHeadMap[O_RESPONSE] = []byte{1,2}
}

// SendReceiveCloser
type SendReceiveCloser interface {
    // send request with no body and get response from server.
    Send(conn net.Conn) (*Response, error)
    // send request with none empty body and get response from server.
    SendWithBody(conn net.Conn, bodyReader io.ReadCloser) (*Response, error)
}

// include a tcp interact request meta data:
// operation bytes, meta length bytes, body length bytes and meta body bytes.
type Request struct {
    Operation int
    HeadBytes []byte
    MetaLength []byte
    BodyLength []byte
    MetaBody []byte
}
// response from server.
type Response struct {
    Operation int
    MetaLength uint64
    BodyLength uint64
    MetaBody string
}


//perform request send job with empty body and receive response.
func (req *Request) Send(conn net.Conn) (*Response, error) {

    return nil, nil
}

//perform request send job with a body and receive response.
func (req *Request) SendWithBody(conn net.Conn, bodyReader io.ReadCloser) (*Response, error) {

    return nil, nil
}




// create a new tcp request using given data
// operation: operation code, such as 'O_CONNECT'
// meta     : meta object
// bodyLen  : request body length
// if create success, it returns a *Request, or else returns with error
func NewRequest(operation int, meta *interface{}, bodyLen uint64) (*Request, error) {
    // operation bytes not found
    if operationHeadMap[operation] == nil {
        return nil, OPERATION_NOT_SUPPORT
    }

    metaBodyBytes, e := json.Marshal(*meta)
    if e != nil {
        return nil, e
    }

    metaSize := len(metaBodyBytes)
    metaLenBytes := make([]byte, 8)
    bodyLenBytes := make([]byte, 8)

    binary.BigEndian.PutUint64(metaLenBytes, uint64(metaSize))
    binary.BigEndian.PutUint64(bodyLenBytes, uint64(bodyLen))

    return &Request{
        operation,
        operationHeadMap[operation],
        metaLenBytes,
        bodyLenBytes,
        metaBodyBytes,
    }, nil
}







