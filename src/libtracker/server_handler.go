package libtracker

import (
	"app"
	"errors"
	"libcommon"
	"libcommon/bridgev2"
	"libservicev2"
	"regexp"
	"strings"
	json "github.com/json-iterator/go"
	"util/logger"
	"validate"
)

func init() {
	registerOperationHandlers()
}

// register handlers as a server side.
func registerOperationHandlers() {
	logger.Debug("register server handlers")
	bridgev2.RegisterOperationHandler(&bridgev2.OperationHandler{bridgev2.FRAME_OPERATION_VALIDATE, libcommon.ValidateConnectionHandler})
	bridgev2.RegisterOperationHandler(&bridgev2.OperationHandler{bridgev2.FRAME_OPERATION_SYNC_STORAGE_MEMBERS, SyncStorageMembersHandler})
	bridgev2.RegisterOperationHandler(&bridgev2.OperationHandler{bridgev2.FRAME_OPERATION_REGISTER_FILES, RegisterFilesHandler})
	bridgev2.RegisterOperationHandler(&bridgev2.OperationHandler{bridgev2.FRAME_OPERATION_PULL_NEW_FILES, PullNewFilesHandlers})
	bridgev2.RegisterOperationHandler(&bridgev2.OperationHandler{bridgev2.FRAME_OPERATION_SYNC_ALL_STORAGE_SEVERS, SyncAllStorageMembersHandler})
	bridgev2.RegisterOperationHandler(&bridgev2.OperationHandler{bridgev2.FRAME_OPERATION_SYNC_STATISTIC, SyncStatisticHandlers})
	bridgev2.RegisterOperationHandler(&bridgev2.OperationHandler{bridgev2.FRAME_OPERATION_QUERY_FILE, QueryFileHandler})
}


// storage server synchronized group members
func SyncStorageMembersHandler(manager *bridgev2.ConnectionManager, frame *bridgev2.Frame) error {
	if frame == nil {
		return libcommon.NULL_FRAME_ERR
	}

	valid := true
	var meta = &app.StorageDO{}
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
		meta.AdvertiseAddr = meta.Host
	}
	resMeta.LookBackAddr = remoteAddr
	if !valid {
		responseFrame.SetStatus(bridgev2.STATUS_INTERNAL_ERROR)
		if e2 := manager.Send(responseFrame); e2 != nil {
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
	responseFrame.SetMeta(resMeta)
	if e2 := manager.Send(responseFrame); e2 != nil {
		return e2
	}
	return nil
}


// storage server synchronized group members
func RegisterFilesHandler(manager *bridgev2.ConnectionManager, frame *bridgev2.Frame) error {
	var meta = &bridgev2.RegisterFileMeta{}
	e1 := json.Unmarshal(frame.FrameMeta, meta)
	if e1 != nil {
		return e1
	}

	resMeta := &bridgev2.RegisterFileResponseMeta{}
	responseFrame := &bridgev2.Frame{}

	if meta.Files == nil || len(meta.Files) == 0 {
		return errors.New("no file registered")
	}

	// check files
	for i := range meta.Files {
		file := meta.Files[i]
		if file.Parts == nil || len(file.Parts) == 0 {
			return errors.New("file has no file part")
		}
	}
	lastId, e2 := libservicev2.InsertRegisteredFiles(meta.Files)
	// e2 := libservice.TrackerAddFile(meta)
	if e2 != nil {
		resMeta.LastInsertId = 0
		responseFrame.SetStatus(bridgev2.STATUS_INTERNAL_ERROR)
		responseFrame.SetMeta(resMeta)
		responseFrame.SetMetaBodyLength(0)
		manager.Send(responseFrame) // send response frame and ignore the result
		return e2
	}
	resMeta.LastInsertId = lastId
	responseFrame.SetStatus(bridgev2.STATUS_SUCCESS)
	responseFrame.SetMeta(resMeta)
	responseFrame.SetMetaBodyLength(0)
	return manager.Send(responseFrame)
}


// client synchronized all storage servers
func SyncAllStorageMembersHandler(manager *bridgev2.ConnectionManager, frame *bridgev2.Frame) error {
	if frame == nil {
		return libcommon.NULL_FRAME_ERR
	}

	resMeta := &bridgev2.SyncAllStorageServerResponseMeta{}
	responseFrame := &bridgev2.Frame{}

	resMeta.Servers = libcommon.GetAllStorageServers()
	responseFrame.SetStatus(bridgev2.STATUS_SUCCESS)
	responseFrame.SetMeta(resMeta)
	responseFrame.SetMetaBodyLength(0)
	return manager.Send(responseFrame)
}


// storage client pull new file of group members.
func PullNewFilesHandlers(manager *bridgev2.ConnectionManager, frame *bridgev2.Frame) error {
	if frame == nil {
		return libcommon.NULL_FRAME_ERR
	}
	var meta = &bridgev2.PullNewFileMeta{}
	e1 := json.Unmarshal(frame.FrameMeta, meta)
	if e1 != nil {
		return e1
	}

	resMeta := &bridgev2.PullNewFileResponseMeta{}
	responseFrame := &bridgev2.Frame{}


	// ret, e2 := libservice.GetFilesBasedOnId(queryMeta.BaseId, false, queryMeta.Group)
	ls, e2 := libservicev2.GetFullFilesFromId(meta.BaseId, false, meta.Group, 50)
	if e2 != nil {
		responseFrame.SetStatus(bridgev2.STATUS_INTERNAL_ERROR)
		responseFrame.SetMeta(resMeta)
		responseFrame.SetMetaBodyLength(0)
		manager.Send(responseFrame) // send response frame and ignore the result
		return e2
	}

	files := make([]app.FileVO, ls.Len())
	index := 0
	for ele := ls.Front(); ele != nil; ele = ele.Next() {
		files[index] = *ele.Value.(*app.FileVO)
		index++
	}
	resMeta.Files = files
	responseFrame.SetStatus(bridgev2.STATUS_SUCCESS)
	responseFrame.SetMeta(resMeta)
	responseFrame.SetMetaBodyLength(0)
	return manager.Send(responseFrame)
}


// dashboard client synchronized statistic info of all storage servers.
func SyncStatisticHandlers(manager *bridgev2.ConnectionManager, frame *bridgev2.Frame) error {
	if frame == nil {
		return libcommon.NULL_FRAME_ERR
	}

	resMeta := &bridgev2.SyncStatisticResponseMeta{}
	responseFrame := &bridgev2.Frame{}

	resMeta.Statistic = libcommon.GetSyncStatistic()
	responseFrame.SetStatus(bridgev2.STATUS_SUCCESS)
	responseFrame.SetMeta(resMeta)
	responseFrame.SetMetaBodyLength(0)
	return manager.Send(responseFrame)

}


// storage server synchronized group members
func QueryFileHandler(manager *bridgev2.ConnectionManager, frame *bridgev2.Frame) error {
	var meta = &bridgev2.QueryFileMeta{}
	e1 := json.Unmarshal(frame.FrameMeta, meta)
	if e1 != nil {
		return e1
	}

	resMeta := &bridgev2.QueryFileResponseMeta{}
	responseFrame := &bridgev2.Frame{}


	var md5 string
	if mat1, _ := regexp.Match(app.Md5Regex, []byte(meta.PathOrMd5)); mat1 {
		md5 = meta.PathOrMd5
	} else if mat2, _ := regexp.Match(app.PathRegex, []byte(meta.PathOrMd5)); mat2 {
		md5 = regexp.MustCompile(app.PathRegex).ReplaceAllString(meta.PathOrMd5, "${4}")
	} else {
		resMeta.Exist = false
		responseFrame.SetStatus(bridgev2.STATUS_SUCCESS)
		responseFrame.SetMeta(resMeta)
		responseFrame.SetMetaBodyLength(0)
		return manager.Send(responseFrame)
	}
	fileInfo, e2 := libservicev2.GetFullFileByMd5(md5, 2)
	if e2 != nil {
		resMeta.Exist = false
		responseFrame.SetStatus(bridgev2.STATUS_INTERNAL_ERROR)
		responseFrame.SetMeta(resMeta)
		responseFrame.SetMetaBodyLength(0)
		manager.Send(responseFrame)
		return e2
	}
	if fileInfo != nil && fileInfo.Id > 0 {
		resMeta.Exist = true
		resMeta.File = *fileInfo
		responseFrame.SetStatus(bridgev2.STATUS_SUCCESS)
		responseFrame.SetMeta(resMeta)
		responseFrame.SetMetaBodyLength(0)
	} else {
		resMeta.Exist = false
		responseFrame.SetStatus(bridgev2.STATUS_SUCCESS)
		responseFrame.SetMeta(resMeta)
		responseFrame.SetMetaBodyLength(0)
	}
	return manager.Send(responseFrame)
}