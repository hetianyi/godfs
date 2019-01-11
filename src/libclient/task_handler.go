package libclient

import (
	"app"
	"errors"
	"io"
	"libcommon/bridge"
	"libcommon/bridgev2"
	"libservice"
	"libservicev2"
	"strconv"
	"util/logger"
)


func TaskSyncMemberHandler(tracker *TrackerInstance) (bool, error) {
	client := *tracker.client
	// storage client to tracker server
	storageInfo := &app.StorageDO{
		Uuid:          app.UUID,
		AdvertiseAddr: app.ADVERTISE_ADDRESS,
		Group:         app.GROUP,
		InstanceId:    app.INSTANCE_ID,
		Port:          app.PORT,
		AdvertisePort: app.ADVERTISE_PORT,
		HttpPort:      app.HTTP_PORT,
		HttpEnable:    app.HTTP_ENABLE,
		TotalFiles:    app.FILE_TOTAL,
		Finish:        app.FILE_FINISH,
		IOin:          app.IOIN,
		IOout:         app.IOOUT,
		Disk:          app.DISK_USAGE,
		Download:      app.DOWNLOADS,
		Upload:        app.UPLOADS,
		StartTime:     app.START_TIME,
		Memory:        app.MEMORY,
		ReadOnly:      !app.UPLOAD_ENABLE,

		StageIOin:      app.STAGE_IOIN,
		StageIOout:     app.STAGE_IOOUT,
		StageDownloads: app.STAGE_DOWNLOADS,
		StageUploads:   app.STAGE_UPLOADS,
	}

	frame := &bridgev2.Frame{}
	frame.SetOperation(bridgev2.FRAME_OPERATION_SYNC_STORAGE_MEMBERS)
	frame.SetMeta(storageInfo)
	frame.SetMetaBodyLength(0)
	response, err := client.SyncStorageMembers(storageInfo)
	if err != nil {
		return true, err
	}
	if response != nil {
		storageMembers(response.GroupMembers)
	} else {
		return true, errors.New("receive empty response from server")
	}

	app.STAGE_DOWNLOADS = 0
	app.STAGE_UPLOADS = 0
	app.STAGE_IOIN = 0
	app.STAGE_IOOUT = 0

	return false, nil
}


func TaskPushFileHandler(tracker *TrackerInstance) (bool, error) {
	client := *tracker.client

	files, e1 := libservicev2.GetReadyPushFiles(tracker.trackerUUID)

	if e1 != nil {
		return false, e1
	}
	if files == nil || files.Len() == 0 {
		return false, nil
	}
	fs := make([]bridge.File, files.Len())
	i := 0
	maxId := 0
	for ele := files.Front(); ele != nil; ele = ele.Next() {
		fs[i] = *ele.Value.(*bridge.File)
		if maxId < fs[i].Id {
			maxId = fs[i].Id
		}
		i++
	}
	// register storage client to tracker server
	regFileMeta := &bridge.OperationRegisterFileRequest{
		Files: fs,
	}
	logger.Info("register", files.Len(), "files to tracker server")
	// reg client
	e2 := connBridge.SendRequest(bridge.O_REG_FILE, regFileMeta, 0, nil)
	if e2 != nil {
		return true, e2
	}
	e5 := connBridge.ReceiveResponse(func(response *bridge.Meta, in io.Reader) error {
		if response.Err != nil {
			return response.Err
		}
		var regResp = &bridge.OperationRegisterFileResponse{}
		e3 := json.Unmarshal(response.MetaBody, regResp)
		if e3 != nil {
			return e3
		}
		if regResp.Status != bridge.STATUS_OK {
			return errors.New("error register file " + strconv.Itoa(task.FileId) + " to tracker server, server response status:" + strconv.Itoa(regResp.Status))
		}
		// update table trackers and set local_push_fid to new id
		e7 := libservice.FinishLocalFilePushTask(maxId, tracker.connBridge.UUID)
		if e7 != nil {
			return e7
		}
		return nil
	})
	if e5 != nil {
		return true, e5
	}
	return false, nil
}