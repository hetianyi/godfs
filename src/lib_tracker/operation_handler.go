package lib_tracker

import (
    "lib_common/bridge"
    "encoding/json"
    "app"
    "regexp"
    "validate"
    "util/logger"
    "strings"
    "errors"
    "net"
    "lib_service"
)

// validate client handler
func validateClientHandler(request *bridge.Meta, connBridge *bridge.Bridge) error {
    var head = &bridge.OperationValidationRequest{}
    e1 := json.Unmarshal(request.MetaBody, head)
    var response = &bridge.OperationValidationResponse{}
    if e1 == nil {
        if head.Secret == app.SECRET {
            response.Status = bridge.STATUS_OK
            response.UUID = app.UUID
            if lib_service.QueryExistsStorageClient(head.UUID) {
                response.IsNew = false
            } else {
                response.IsNew = true
            }
            e1 = lib_service.RegisterStorageClient(head.UUID)
        } else {
            response.Status = bridge.STATUS_BAD_SECRET
        }
    } else {
        response.Status = bridge.STATUS_INTERNAL_SERVER_ERROR
    }
    logger.Info("register storage client:", head.UUID)
    e3 := connBridge.SendResponse(response, 0, nil)
    if e1 != nil {
        return e1
    }
    if e3 != nil {
        return e3
    }
    return nil
}


func registerStorageClientHandler(request *bridge.Meta, conn net.Conn,connBridge *bridge.Bridge) (*bridge.OperationRegisterStorageClientRequest, error) {
    valid := true
    var meta = &bridge.OperationRegisterStorageClientRequest{}
    logger.Debug(string(request.MetaBody))
    e1 := json.Unmarshal(request.MetaBody, meta)
    if e1 != nil {
        return nil, e1
    }
    var response = &bridge.OperationRegisterStorageClientResponse{}
    //check meta fields
    if mat, _ := regexp.Match(validate.GroupInstancePattern, []byte(meta.Group)); !mat {
        logger.Error("register failed: group or instance_id is invalid")
        valid = false
    }
    if meta.Port < 1 || meta.Port > 65535 || meta.InstanceId == "" {
        logger.Error("register failed: error parameter")
        valid = false
    }
    remoteAddr := strings.Split(conn.RemoteAddr().String(), ":")[0]
    if meta.BindAddr == "" {
        logger.Warn("storage server not send bind address, using", remoteAddr)
        meta.BindAddr = remoteAddr
    }
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
    if e2 != nil{
        return nil, e2
    }
    return meta, nil
}


func registerFileHandler(request *bridge.Meta, connBridge *bridge.Bridge) error {
    var meta = &bridge.OperationRegisterFileRequest{}
    e1 := json.Unmarshal(request.MetaBody, meta)
    if e1 != nil {
        return e1
    }
    var response = &bridge.OperationRegisterFileResponse{}
    // validate success
    e2 := lib_service.TrackerAddFile(meta)
    if e2 != nil {
        response.Status = bridge.STATUS_INTERNAL_SERVER_ERROR
    } else {
        response.Status = bridge.STATUS_OK
    }
    return connBridge.SendResponse(response, 0, nil)
}


// for client
func syncStorageServerHandler(request *bridge.Meta, connBridge *bridge.Bridge) error {
    var meta = &bridge.OperationGetStorageServerRequest{}
    logger.Debug(string(request.MetaBody))
    e1 := json.Unmarshal(request.MetaBody, meta)
    if e1 != nil {
        return e1
    }
    var response = &bridge.OperationGetStorageServerResponse{}
    response.Status = bridge.STATUS_OK
    response.GroupMembers = GetAllStorages()
    e2 := connBridge.SendResponse(response, 0, nil)
    if e2 != nil{
        return e2
    }
    return nil
}



func pullNewFile(request *bridge.Meta, connBridge *bridge.Bridge) error {
    var queryMeta = &bridge.OperationPullFileRequest{}
    e1 := json.Unmarshal(request.MetaBody, queryMeta)
    var response = &bridge.OperationPullFileResponse{}
    if e1 != nil {
        response.Status = bridge.STATUS_INTERNAL_SERVER_ERROR
        // ignore if it write success
        connBridge.SendResponse(response, 0, nil)
        return e1
    }
    ret, e2 := lib_service.GetFilesBasedOnId(queryMeta.BaseId)
    if e2 != nil {
        response.Status = bridge.STATUS_INTERNAL_SERVER_ERROR
        // ignore if it write success
        connBridge.SendResponse(response, 0, nil)
        return nil
    }
    response.Status = bridge.STATUS_OK
    files := make([]bridge.File, ret.Len())
    i := 0
    for ele := ret.Front(); ele != nil; ele = ele.Next() {
        files[i] = *ele.Value.(*bridge.File)
        i++
    }
    response.Files = files
    s, e3 := json.Marshal(response)
    if e3 != nil {
        return e3
    } else {
        logger.Debug(string(s))
    }
    return connBridge.SendResponse(response, 0, nil)
}

