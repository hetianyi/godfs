package bridgev2

import (
    "errors"
    "util/json"
    "app"
    "libservice"
)

var NULL_FRAME_ERR = errors.New("frame is null")

// <---------------------------------------------------- handlers ---------------------------------------------------->

func ValidateClientHandler(manager *ConnectionManager, frame *Frame) error {
    if frame == nil {
        return NULL_FRAME_ERR
    }

    var meta = &ConnectMeta{}
    e1 := json.Unmarshal(frame.frameMeta, meta)
    if e1 != nil {
        return e1
    }

    response := &ConnectResponseMeta{
        UUID: app.UUID,
        New4Tracker: false,
    }

    responseFrame := &Frame{}

    if meta.Secret == app.SECRET {
        responseFrame.SetStatus(SUCCESS)
        responseFrame.SetMeta(response)
        exist, e2 := libservice.QueryExistsStorageClient(meta.UUID)
        if e2 != nil {
            responseFrame.SetStatus(INTERNAL_ERROR)
        } else {
            if exist {
                response.New4Tracker = false
            } else {
                response.New4Tracker = true
            }
        }
        // only valid client uuid (means storage client) will log into db.
        if meta.UUID != "" && len(meta.UUID) == 30 {
            e1 = libservice.RegisterStorageClient(meta.UUID)
            if e1 != nil {
                responseFrame.SetStatus(INTERNAL_ERROR)
            }
        }
    } else {
        responseFrame.SetStatus(INTERNAL_ERROR)
    }
    return writeFrame(manager, responseFrame)
}


