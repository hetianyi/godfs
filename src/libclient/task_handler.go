package libclient

import (
	"app"
	"errors"
	"libcommon/bridgev2"
	"libservicev2"
	"util/logger"
	"util/timeutil"
	"time"
	"util/common"
	"libservice"
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


// register file to tracker
func TaskRegisterFileHandler(tracker *TrackerInstance) (bool, error) {
	client := *tracker.client

	files, e1 := libservicev2.GetReadyPushFiles(tracker.trackerUUID)

	if e1 != nil {
		return false, e1
	}
	if files == nil || files.Len() == 0 {
		return false, nil
	}
	fs := make([]app.FileVO, files.Len())
	i := 0
	var maxId int64 = 0
	for ele := files.Front(); ele != nil; ele = ele.Next() {
		fs[i] = *new(app.FileVO).From(ele.Value.(*app.FileDO))
		if maxId < fs[i].Id {
			maxId = fs[i].Id
		}
		i++
	}
	// register storage client to tracker server
	regFileMeta := &bridgev2.RegisterFileMeta{
		Files: fs,
	}
	logger.Info("register", files.Len(), "files to tracker server")

	responseMeta, e2 := client.RegisterFiles(regFileMeta)
	if e2 != nil {
		return true, e2
	}

	e3 := libservicev2.UpdateTrackerWithMap(tracker.trackerUUID,
		map[string]interface{}{"tracker_sync_id": responseMeta.LastInsertId, "local_push_id": maxId}, nil)
	if e3 != nil {
		return false, e2
	}
	return false, nil
}



// register file to tracker
func TaskPullFileHandler(tracker *TrackerInstance) (bool, error) {
	client := *tracker.client

	config, e1 := libservicev2.GetTracker(tracker.trackerUUID)
	// config, e1 := libservice.GetTrackerConfig(tracker.connBridge.UUID)
	if e1 != nil {
		return false, e1
	}
	if config == nil {
		h, p := common.ParseHostPortFromConnStr(tracker.ConnStr)
		config = &app.TrackerDO{
			Uuid: tracker.trackerUUID,
			TrackerSyncId: 0,
			LastRegTime: timeutil.GetTimestamp(time.Now()),
			LocalPushId: 0,
			Host: h,
			Port: p,
			Status: app.STATUS_ENABLED,
			Secret: app.SECRET,
			TotalFiles: 0,
			Remark: "",
			AddTime: timeutil.GetTimestamp(time.Now()),
		}
		if e2 := libservicev2.SaveTracker(config); e2 != nil {
			return false, e2
		}
	}
	// register storage client to tracker server
	pullMeta := &bridgev2.PullFileMeta{
		BaseId: config.TrackerSyncId,
		Group:  app.GROUP,
	}

	responseMeta, e2 := client.PullFiles(pullMeta)
	if e2 != nil {
		return true, e2
	}
	files := responseMeta.Files
	if files == nil || len(files) > 0 {
		logger.Info("pull", len(files), "files from tracker server:", tracker.ConnStr)
		return false, nil
	} else {
		return false, libservicev2.InsertTrackerFile(tracker.trackerUUID, responseMeta.Files, nil)
	}
}


func TaskDownloadFileHandler(tracker *TrackerInstance) (bool, error) {
	client := *tracker.client
	logger.Debug("trying download file from other storage server...")
	if increaseActiveDownload(0) >= ParallelDownload {
		logger.Debug("ParallelDownload reached")
		// AddTask(task, tracker)
		return false, nil
	}
	fi, e1 := libservice.GetFullFileByFid(task.FileId, 0)
	if e1 != nil {
		return false, e1
	}
	if fi == nil || len(fi.Parts) == 0 {
		return false, nil
	}
	addDownloadingFile(fi.Id, false)
	go downloadFile(fi)
	return false, nil
}

