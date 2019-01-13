package libclient

import (
	"libcommon/bridgev2"
	"io"
)

func init() {
	registerOperationHandlers()
}

// register
func registerOperationHandlers() {
	bridgev2.RegisterOperationHandler(&bridgev2.OperationHandler{bridgev2.FRAME_OPERATION_VALIDATE, bridgev2.ValidateConnectionHandler, nil, nil})
}

// validate connection
func UploadFileHandler(manager *bridgev2.ConnectionManager, frame *bridgev2.Frame) error {
	if frame == nil {
		return bridgev2.NULL_FRAME_ERR
	}
}
