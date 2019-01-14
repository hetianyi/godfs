package libstorage

import (
	"app"
	"container/list"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"libcommon"
	"libcommon/bridgev2"
	"libservicev2"
	"regexp"
	"util/file"
	"util/json"
	"util/logger"
)

func init() {
	registerOperationHandlers()
}

// register handlers as a server side.
func registerOperationHandlers() {
	if app.UPLOAD_ENABLE {
		bridgev2.RegisterOperationHandler(&bridgev2.OperationHandler{bridgev2.FRAME_OPERATION_VALIDATE, bridgev2.ValidateConnectionHandler})
		bridgev2.RegisterOperationHandler(&bridgev2.OperationHandler{bridgev2.FRAME_OPERATION_UPLOAD_FILE, UploadFileHandler})
		bridgev2.RegisterOperationHandler(&bridgev2.OperationHandler{bridgev2.FRAME_OPERATION_DOWNLOAD_FILE, DownFileHandler})
	}
}

// upload file handler.
func UploadFileHandler(manager *bridgev2.ConnectionManager, frame *bridgev2.Frame) error {
	if frame == nil {
		return bridgev2.NULL_FRAME_ERR
	}
	manager.Md.Reset()
	logger.Info("begin read file body, file len is ", frame.BodyLength/1024, "KB")
	buffer, _ := bridgev2.MakeBytes(app.BUFF_SIZE, false, 0, false)
	defer func() {
		manager.Md.Reset()
		bridgev2.RecycleBytes(buffer)
	}()
	out, oe := libcommon.CreateTmpFile()
	if oe != nil {
		return oe
	}
	// total read bytes
	var readBodySize int64 = 0
	// next time bytes to read
	var nextReadSize int
	var sliceReadSize int64 = 0
	var sliceMd5 = md5.New()
	var fileParts list.List

	for {
		// read finish
		if readBodySize == frame.BodyLength {
			totalCipherStr := manager.Md.Sum(nil)
			sliceCipherStr := sliceMd5.Sum(nil)
			md5 := hex.EncodeToString(totalCipherStr)
			sMd5 := hex.EncodeToString(sliceCipherStr)
			out.Close()
			sliceMd5.Reset()

			e10 := libcommon.MoveTmpFileTo(sMd5, out)
			if e10 != nil {
				libcommon.CloseAndDeleteTmpFile(out)
				return e10
			}
			tmpPart := &app.PartDO{Md5: sMd5, Size: sliceReadSize}
			fileParts.PushBack(tmpPart)
			logger.Info("upload finish, total read bytes", readBodySize, "| MD5:", md5)
			app.UpdateUploads()

			finalFile := &app.FileVO{
				Md5: md5,
				PartNumber: fileParts.Len(),
				Group: app.GROUP,
				Instance: app.INSTANCE_ID,
				Finish: 1,
				FileSize: readBodySize,
			}
			parts := make([]app.PartDO, fileParts.Len())
			index := 0
			for ele := fileParts.Front(); ele != nil; ele = ele.Next() {
				parts[index] = *ele.Value.(*app.PartDO)
				index++
			}
			finalFile.Parts = parts

			// stoe := libservice.StorageAddFile(md5, app.GROUP, &fileParts)
			stoe := libservicev2.InsertFile(finalFile, nil)
			if stoe != nil {
				return stoe
			}
			// mark the file is multi part or single part
			var path string
			if fileParts.Len() > 1 {
				path = app.GROUP + "/" + app.INSTANCE_ID + "/M/" + md5
			} else {
				path = app.GROUP + "/" + app.INSTANCE_ID + "/S/" + md5
			}

			responseMeta := &bridgev2.UploadFileResponseMeta{path}
			frame := &bridgev2.Frame{}
			frame.SetStatus(bridgev2.STATUS_SUCCESS)
			frame.SetMeta(responseMeta)
			frame.SetMetaBodyLength(0)

			e13 := manager.Send(frame)
			if e13 != nil {
				return e13
			}
			return nil
		}
		// left bytes is more than a buffer
		if (frame.BodyLength-readBodySize)/int64(app.BUFF_SIZE) >= 1 {
			nextReadSize = int(app.BUFF_SIZE)
		} else { // left bytes less than a buffer
			nextReadSize = int(frame.BodyLength - readBodySize)
		}
		logger.Trace("read next bytes:", nextReadSize, "total is:", frame.BodyLength)
		len1, e3 := bridgev2.ReadBytes(buffer, nextReadSize, manager, manager.Md)
		if e3 == nil && len1 == nextReadSize {
			// if sliceReadSize > sliceSize then create a new slice file
			if sliceReadSize+int64(len1) > app.SLICE_SIZE {
				// write bytes to file
				leftN := app.SLICE_SIZE - sliceReadSize
				rightN := int64(len1) - (app.SLICE_SIZE - sliceReadSize)
				len2, e1 := out.Write(buffer[0:leftN])
				len4, e11 := sliceMd5.Write(buffer[0:leftN])
				if e1 != nil || e11 != nil || int64(len2) != leftN || int64(len4) != leftN {
					logger.Error("write out error:", e1, "|", e11)
					libcommon.CloseAndDeleteTmpFile(out)
					return errors.New("write out error(0)")
				}

				// close slice file and create a new slice file （承上启下）
				out.Close()
				sliceCipherStr := sliceMd5.Sum(nil)
				sMd5 := hex.EncodeToString(sliceCipherStr)
				sliceMd5.Reset()
				e10 := libcommon.MoveTmpFileTo(sMd5, out)
				if e10 != nil {
					return e10
				}
				tmpPart := &app.PartDO{Md5: sMd5, Size: app.SLICE_SIZE}
				fileParts.PushBack(tmpPart)
				app.UpdateDiskUsage(app.SLICE_SIZE)

				out12, e12 := libcommon.CreateTmpFile()
				if e12 != nil {
					return e12
				}
				out = out12
				len6, e2 := out.Write(buffer[leftN:len1])
				len7, e12 := sliceMd5.Write(buffer[leftN:len1])
				if e2 != nil || e12 != nil || int64(len6) != rightN || int64(len7) != rightN {
					logger.Error("write out error:", e2, "|", e12)
					libcommon.CloseAndDeleteTmpFile(out)
					return errors.New("write out error(1)")
				}
				sliceReadSize = rightN
			} else {
				// write bytes to file
				len2, e1 := out.Write(buffer[0:len1])
				len4, e3 := sliceMd5.Write(buffer[0:len1])
				// write error
				if e1 != nil || e3 != nil || len2 != len1 || len4 != len1 {
					logger.Error("write out error:", e1)
					libcommon.CloseAndDeleteTmpFile(out)
					return errors.New("write out error(0)")
				}
				sliceReadSize += int64(len1)
			}
			readBodySize += int64(len1)
		} else {
			logger.Error("error read body:", e3)
			libcommon.CloseAndDeleteTmpFile(out)
			// stop loop
			return e3
		}
	}
}


// download file handler.
func DownFileHandler(manager *bridgev2.ConnectionManager, frame *bridgev2.Frame) error {
	if frame == nil {
		return bridgev2.NULL_FRAME_ERR
	}


	var meta = &bridgev2.DownloadFileMeta{}
	e1 := json.Unmarshal(frame.FrameMeta, meta)
	if e1 != nil {
		return e1
	}

	resMeta := &bridgev2.DownloadFileResponseMeta{}
	responseFrame := &bridgev2.Frame{}

	if mat, _ := regexp.Match(app.PATH_REGEX, []byte(meta.Path)); !mat {
		logger.Debug("error file path format")
		resMeta.Exist = false
		responseFrame.SetStatus(bridgev2.STATUS_SUCCESS)
		responseFrame.SetMeta(resMeta)
		responseFrame.SetMetaBodyLength(0)
		return manager.Send(frame)
	}
	md5 := regexp.MustCompile(app.PATH_REGEX).ReplaceAllString(meta.Path, "${4}")

	// fullFile, e11 := libservice.GetFullFileByMd5(md5, 1)
	fileInfo, e2 := libservicev2.GetFullFileByMd5(md5, 1)
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
		startPos, endPos, totalLen := libcommon.GetReadPositions(fileInfo, meta.Start, meta.Offset)
		responseFrame.SetMetaBodyLength(totalLen)
		logger.Debug("download from: ", startPos.PartIndex, ":", startPos.PartStart)
		logger.Debug("download to  : ", endPos.PartIndex, ":", endPos.PartStart)
		logger.Debug("download size: ", totalLen)
		app.UpdateDownloads()
		responseFrame.BodyWriterHandler = func(manager *bridgev2.ConnectionManager, frame *bridgev2.Frame) error {
			return WriteDownloadStream(fileInfo, startPos, endPos, manager.Conn)
		}
		return manager.Send(responseFrame)
	} else {
		resMeta.Exist = false
		responseFrame.SetStatus(bridgev2.STATUS_SUCCESS)
		responseFrame.SetMeta(resMeta)
		responseFrame.SetMetaBodyLength(0)
		return manager.Send(responseFrame)
	}
}


// download writer handler.
func WriteDownloadStream(fullFile *app.FileVO, startPos *bridgev2.ReadPos, endPos *bridgev2.ReadPos, out io.Writer) error {
	buffer, _ := bridgev2.MakeBytes(app.BUFF_SIZE, false, 0, false)
	defer bridgev2.RecycleBytes(buffer)
	for i := range fullFile.Parts {
		var start int64 = 0
		var offset int64 = 0
		if i < startPos.PartIndex {
			continue
		} else if i == startPos.PartIndex {
			start = startPos.PartStart
		} else {
			start = 0
		}
		if i > endPos.PartIndex {
			break
		} else if i == endPos.PartIndex {
			offset = endPos.PartStart - start
		} else {
			offset = fullFile.Parts[i].Size - start
		}
		fi, e := file.GetFile(libcommon.GetFilePathByMd5(fullFile.Parts[i].Md5))
		if e != nil {
			return e
		}
		if e := libcommon.SeekWriteOut(fi, start, offset, buffer, out); e != nil {
			fi.Close()
			return e
		}
		fi.Close()
	}
	return nil
}
