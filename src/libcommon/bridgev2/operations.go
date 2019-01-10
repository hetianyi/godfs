package bridgev2

import "io"

const (
    SUCCESS              byte = 1
    INTERNAL_ERROR       byte = 2
    BAD_SECRET           byte = 3
    CONNECTION_POOL_FULL byte = 4

    FRAME_OPERATION_NONE          byte = 0 // none operation, means no operation specified
    FRAME_OPERATION_VALIDATE      byte = 2 // operation for connect/validate
    FRAME_OPERATION_TEST_SEND_MSG byte = 3 // operation : test send msg
    FRAME_OPERATION_SYNC_STORAGE_MEMBERS byte = 4 // operation : test send msg
)

// operation handler mapping
var operationHandlerMap = make(map[byte]*OperationHandler)
// operation response message mapping
var responseCodeMap = make(map[byte]string)

func init() {
    responseCodeMap[SUCCESS] = "operation success"
    responseCodeMap[INTERNAL_ERROR] = "internal server error"
    responseCodeMap[BAD_SECRET] = "bad secret"
    responseCodeMap[CONNECTION_POOL_FULL] = "connection pool is full"


    operationHandlerMap[FRAME_OPERATION_VALIDATE] = &OperationHandler{FRAME_OPERATION_VALIDATE, ValidateConnectionHandler, nil, nil}
}

func TranslateResponseMsg(code byte) string {
    return responseCodeMap[code]
}


type OperationHandler struct {
    OperationCode     byte
    MetaHandler func(manager *ConnectionManager, frame *Frame) error
    BodyReaderHandler func(frame *Frame, reader io.Reader) error
    BodyWriterHandler func(frame *Frame, writer io.Writer) error
}


func GetOperationHandler(operation byte) *OperationHandler {
    return operationHandlerMap[operation]
}

