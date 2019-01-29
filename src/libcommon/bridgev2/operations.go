package bridgev2

import (
	"util/logger"
)

const (
	StatusSuccess              byte = 1
	StatusInternalErr       byte = 2
	StatusBadSecret           byte = 3
	StatusFullConnectionPool byte = 4
	StatusInstanceIdExist   byte = 5

	FrameOperationNone                    byte = 0
	FrameOperationValidate                byte = 2
	FrameOperationSyncStorageMember    byte = 4
	FrameOperationRegisterFiles          byte = 5
	FrameOperationSyncAllStorageServers byte = 6
	FrameOperationPullNewFiles          byte = 7
	FrameOperationSyncStatistic          byte = 8
	FrameOperationQueryFile              byte = 9
	FrameOperationUploadFile             byte = 10
	FrameOperationDownloadFile           byte = 11
)

// operation handler mapping
var operationHandlerMap = make(map[byte]*OperationHandler)
// operation response message mapping
var responseCodeMap = make(map[byte]string)

func init() {
	responseCodeMap[StatusSuccess] = "operation success"
	responseCodeMap[StatusInternalErr] = "internal server error"
	responseCodeMap[StatusBadSecret] = "bad secret"
	responseCodeMap[StatusFullConnectionPool] = "connection pool is full"
	responseCodeMap[StatusInstanceIdExist] = "instance id is not unique"
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
