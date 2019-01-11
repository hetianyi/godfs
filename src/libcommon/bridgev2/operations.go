package bridgev2

import (
    "io"
    "util/logger"
)

const (
    STATUS_SUCCESS              byte = 1
    STATUS_INTERNAL_ERROR       byte = 2
    STATUS_BAD_SECRET           byte = 3
    STATUS_CONNECTION_POOL_FULL byte = 4

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
    responseCodeMap[STATUS_SUCCESS] = "operation success"
    responseCodeMap[STATUS_INTERNAL_ERROR] = "internal server error"
    responseCodeMap[STATUS_BAD_SECRET] = "bad secret"
    responseCodeMap[STATUS_CONNECTION_POOL_FULL] = "connection pool is full"
}

func TranslateResponseMsg(code byte) string {
    return responseCodeMap[code]
}

// register handler dynamically from high level.
// usually register all handlers at entry file, such as tracker.go, storage.go...
func RegisterOperationHandler(handler *OperationHandler) {
    if handler == nil {
        return
    }
    if operationHandlerMap[handler.OperationCode] != nil {
        logger.Warn("handler already registered:", handler.OperationCode)
        return
    }
    logger.Debug("register operation handler:", handler.OperationCode)
    operationHandlerMap[handler.OperationCode] = handler
}


type OperationHandler struct {
    OperationCode     byte
    MetaHandler func(manager *ConnectionManager, frame *Frame) error // handler is used by server
    BodyReaderHandler func(frame *Frame, reader io.Reader) error // handler is used by server
    BodyWriterHandler func(frame *Frame, writer io.Writer) error // handler is used by client
}


func GetOperationHandler(operation byte) *OperationHandler {
    return operationHandlerMap[operation]
}

