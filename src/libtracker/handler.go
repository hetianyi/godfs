package libtracker

import (
	"app"
	"errors"
	"libcommon"
	"libcommon/bridgev2"
	"regexp"
	"strings"
	"util/json"
	"util/logger"
	"validate"
)

// validate connection
func SyncStorageMembersHandler(manager *bridgev2.ConnectionManager, frame *bridgev2.Frame) error {
	if frame == nil {
		return bridgev2.NULL_FRAME_ERR
	}

	valid := true
	var meta = &app.StorageDO{}
	// logger.Debug(string(request.MetaBody))
	e1 := json.Unmarshal(frame.FrameMeta, meta)
	if e1 != nil {
		return e1
	}

	resMeta := &bridgev2.SyncStorageMembersResponseMeta{}
	responseFrame := &bridgev2.Frame{}

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
	meta.Host = remoteAddr
	resMeta.LookBackAddr = remoteAddr
	if !libcommon.IsInstanceIdUnique(meta) {
		logger.Error("register failed: instance_id is not unique")
		valid = false
	}
	if !valid {
		responseFrame.SetStatus(bridgev2.STATUS_INTERNAL_ERROR)
		if e2 := bridgev2.WriteFrame(manager, responseFrame); e2 != nil {
			return e2
		}
		return errors.New("invalid meta data")
	}
	// validate success
	if e2 := libcommon.CacheStorageServer(meta); e2 != nil {
		logger.Error("cannot cache/persist storage server info")
		return e2
	}
	responseFrame.SetStatus(bridgev2.STATUS_SUCCESS)
	resMeta.GroupMembers = libcommon.GetGroupMembers(meta)
	if e2 := bridgev2.WriteFrame(manager, responseFrame); e2 != nil {
		return e2
	}
	return nil
}
