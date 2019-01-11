package libcommon

import (
	"app"
	"container/list"
	"libcommon/bridgev2"
)

func InitRegistry() {

	handlers := list.New()

	if app.RUN_WITH == 1 {
		handlers.PushBack(&bridgev2.OperationHandler{bridgev2.FRAME_OPERATION_VALIDATE, bridgev2.ValidateConnectionHandler, nil, nil})
		handlers.PushBack(&bridgev2.OperationHandler{bridgev2.FRAME_OPERATION_SYNC_STORAGE_MEMBERS, bridgev2.ValidateConnectionHandler, nil, nil})
	} else if app.RUN_WITH == 2 {

	} else if app.RUN_WITH == 3 {

	} else if app.RUN_WITH == 4 {

	}

	for ele := handlers.Front(); ele != nil; ele = ele.Next() {
		bridgev2.RegisterOperationHandler(ele.Value.(*bridgev2.OperationHandler))
	}
}
