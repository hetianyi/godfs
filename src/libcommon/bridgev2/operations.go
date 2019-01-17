package bridgev2

import (
	"util/logger"
)

const (
	STATUS_SUCCESS              byte = 1
	STATUS_INTERNAL_ERROR       byte = 2
	STATUS_BAD_SECRET           byte = 3
	STATUS_CONNECTION_POOL_FULL byte = 4
	STATUS_INSTANCE_ID_EXISTS   byte = 5

	FRAME_OPERATION_NONE                    byte = 0
	FRAME_OPERATION_VALIDATE                byte = 2
	FRAME_OPERATION_TEST_SEND_MSG           byte = 3
	FRAME_OPERATION_SYNC_STORAGE_MEMBERS    byte = 4
	FRAME_OPERATION_REGISTER_FILES          byte = 5
	FRAME_OPERATION_SYNC_ALL_STORAGE_SEVERS byte = 6
	FRAME_OPERATION_PULL_NEW_FILES          byte = 7
	FRAME_OPERATION_SYNC_STATISTIC          byte = 8
	FRAME_OPERATION_QUERY_FILE              byte = 9
	FRAME_OPERATION_UPLOAD_FILE             byte = 10
	FRAME_OPERATION_DOWNLOAD_FILE           byte = 11
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
	responseCodeMap[STATUS_INSTANCE_ID_EXISTS] = "instance id is not unique"
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
	OperationCode byte
	Handler       func(manager *ConnectionManager, frame *Frame) error // handler is used by server
}

func GetOperationHandler(operation byte) *OperationHandler {
	return operationHandlerMap[operation]
}
