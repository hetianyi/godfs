package bridgev2

import (
    "errors"
    "util/json"
    "app"
    "libservicev2"
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
        exist, e2 :=libservicev2.ExistsStorage(meta.UUID)
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
            storage := &app.StorageDO{
                Uuid: meta.UUID,
                Host: "",
                Port: 0,
                Status: app.STATUS_ENABLED,
                TotalFiles: 0,
                Group: "",
                InstanceId: "",
                HttpPort: 0,
                HttpEnable: false,
                StartTime: 0,
                Download: 0,
                Upload: 0,
                Disk: 0,
                ReadOnly: false,
                Finish: 0,
                IOin: 0,
                IOout: 0,
            }
            e3 := libservicev2.SaveStorage(storage)
            if e3 != nil {
                responseFrame.SetStatus(INTERNAL_ERROR)
            }
        }
        responseFrame.SetMeta(response)
    } else {
        responseFrame.SetStatus(INTERNAL_ERROR)
    }
    return writeFrame(manager, responseFrame)
}


