package bridgev2

import (
    "errors"
    "util/json"
    "app"
    "libservicev2"
    "util/logger"
    "strings"
    "libcommon/bridge"
    "regexp"
    "validate"
)

var NULL_FRAME_ERR = errors.New("frame is null")


// validate connection
func ValidateConnectionHandler(manager *ConnectionManager, frame *Frame) error {
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

// validate connection
func SyncStorageMembersHandler(manager *ConnectionManager, frame *Frame) error {
    if frame == nil {
        return NULL_FRAME_ERR
    }

    valid := true
    var meta = &SyncStorageMembersMeta{}
    // logger.Debug(string(request.MetaBody))
    e1 := json.Unmarshal(frame.frameMeta, meta)
    if e1 != nil {
        return e1
    }

    resMeta := &SyncStorageMembersResponseMeta{}
    responseFrame := &Frame{}


    // check meta fields
    if mat, _ := regexp.Match(validate.GroupInstancePattern, []byte(meta.Group)); !mat {
        logger.Error("register failed: invalid group or instance_id")
        valid = false
    }
    if meta.Port < 1 || meta.Port > 65535 || meta.InstanceId == "" {
        logger.Error("register failed: error parameter")
        valid = false
    }
    remoteAddr := string([]rune(manager.Conn.RemoteAddr().String())[0:strings.LastIndex(manager.Conn.RemoteAddr().String(), ":")])
    if meta.AdvertiseAddr == "" {
        logger.Debug("storage server has no advertise address, using", remoteAddr)
        meta.AdvertiseAddr = remoteAddr
    }
    meta.LookBackAddress = remoteAddr
    resMeta.LookBackAddr = remoteAddr
    if !IsInstanceIdUnique(meta) {
        logger.Error("register failed: instance_id is not unique")
        valid = false
    }
    if !valid {
        response.Status = bridge.STATUS_INTERNAL_SERVER_ERROR
        connBridge.SendResponse(response, 0, nil)
        return nil, errors.New("invalid meta data")
    }
    // validate success
    AddStorageServer(meta)
    response.Status = bridge.STATUS_OK
    response.LookBackAddr = remoteAddr
    response.GroupMembers = GetGroupMembers(meta)
    e2 := connBridge.SendResponse(response, 0, nil)
    if e2 != nil {
        return nil, e2
    }
    return meta, nil

}







