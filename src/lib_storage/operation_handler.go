package lib_storage

import (
    "lib_common/bridge"
    "encoding/json"
    "app"
    "util/logger"
    "errors"
    "crypto/md5"
    "container/list"
    "encoding/hex"
    "lib_service"
    "hash"
    "lib_common"
    "io"
    "regexp"
    "util/file"
)

// validate client handler
func validateClientHandler(request *bridge.Meta, connBridge *bridge.Bridge) error {
    var head = &bridge.OperationValidationRequest{}
    e1 := json.Unmarshal(request.MetaBody, head)
    var response = &bridge.OperationValidationResponse{}
    if e1 == nil {
        if head.Secret == app.SECRET {
            response.Status = bridge.STATUS_OK
        } else {
            response.Status = bridge.STATUS_BAD_SECRET
        }
    } else {
        response.Status = bridge.STATUS_INTERNAL_SERVER_ERROR
    }
    e3 := connBridge.SendResponse(response, 0, nil)
    if e1 != nil {
        return e1
    }
    if e3 != nil {
        return e3
    }
    return nil
}

// 处理文件上传请求
func uploadHandler(request *bridge.Meta, buffer []byte, md hash.Hash, conn io.ReadCloser, connBridge *bridge.Bridge) error {
    logger.Info("begin read file body, file len is ", request.BodyLength/1024, "KB")

    defer func() {
        md.Reset()
    }()
    out, oe := lib_common.CreateTmpFile()
    if oe != nil {
        return oe
    }
    // total read bytes
    var readBodySize uint64 = 0
    // next time bytes to read
    var nextReadSize int
    var sliceReadSize int64 = 0
    var sliceMd5 = md5.New()
    var sliceIds list.List

    for {
        //read finish
        if readBodySize == request.BodyLength {
            totalCipherStr := md.Sum(nil)
            sliceCipherStr := sliceMd5.Sum(nil)
            md5  := hex.EncodeToString(totalCipherStr)
            sMd5 := hex.EncodeToString(sliceCipherStr)
            out.Close()
            md.Reset()
            sliceMd5.Reset()

            e10 := lib_common.MoveTmpFileTo(sMd5, out)
            if e10 != nil {
                return e10
            }
            // save slice info to db
            pid, e8 := lib_service.AddPart(sMd5, sliceReadSize)
            if e8 != nil {
                return e8
            }
            sliceIds.PushBack(pid)
            logger.Info("上传结束，读取字节：", readBodySize, " MD5= " , md5)

            stoe := lib_service.StorageAddFile(md5, &sliceIds)
            if stoe != nil {
                return stoe
            }
            // mark the file is multi part or single part
            var path string
            if sliceIds.Len() > 1 {
                path = app.GROUP + "/" + app.INSTANCE_ID + "/M/" + md5
            } else {
                path = app.GROUP + "/" + app.INSTANCE_ID + "/S/" + md5
            }

            var response = &bridge.OperationUploadFileResponse {
                Status: bridge.STATUS_OK,
                Path: path,
            }
            e13 := connBridge.SendResponse(response, 0, nil)
            if e13 != nil {
                return e13
            }
            return nil
        }
        // left bytes is more than a buffer
        if (request.BodyLength - readBodySize) / uint64(app.BUFF_SIZE) >= 1 {
            nextReadSize = int(app.BUFF_SIZE)
        } else {// left bytes less than a buffer
            nextReadSize = int(request.BodyLength - readBodySize)
        }
        logger.Trace("read next bytes:", nextReadSize, "total is:", request.BodyLength)
        len1, e3 := bridge.ReadBytes(buffer, nextReadSize, conn)
        if e3 == nil && len1 == nextReadSize {
            // if sliceReadSize > sliceSize then create a new slice file
            if sliceReadSize + int64(len1) > app.SLICE_SIZE {
                // write bytes to file
                leftN := app.SLICE_SIZE - sliceReadSize
                rightN := int64(len1) - (app.SLICE_SIZE - sliceReadSize)
                len2, e1 := out.Write(buffer[0:leftN])
                len4, e11 := sliceMd5.Write(buffer[0:leftN])
                if e1 != nil || e11 != nil || int64(len2) != leftN || int64(len4) != leftN {
                    logger.Error("write out error:", e1, "|", e11)
                    lib_common.CloseAndDeleteTmpFile(out)
                    return errors.New("write out error(0)")
                }

                // close slice file and create a new slice file （承上启下）
                out.Close()
                sliceCipherStr := sliceMd5.Sum(nil)
                sMd5 := hex.EncodeToString(sliceCipherStr)
                sliceMd5.Reset()
                e10 := lib_common.MoveTmpFileTo(sMd5, out)
                if e10 != nil {
                    return e10
                }
                // save slice info to db
                pid, e8 := lib_service.AddPart(sMd5, app.SLICE_SIZE)
                if e8 != nil {
                    return e8
                }
                sliceIds.PushBack(pid)

                out12, e12 := lib_common.CreateTmpFile()
                if e12 != nil {
                    return e12
                }
                out = out12
                len6, e2 := out.Write(buffer[leftN:len1])
                len7, e12 := sliceMd5.Write(buffer[leftN:len1])
                if e2 != nil || e12 != nil || int64(len6) != rightN || int64(len7) != rightN {
                    logger.Error("write out error:", e2, "|", e12)
                    lib_common.CloseAndDeleteTmpFile(out)
                    return errors.New("write out error(1)")
                }
                sliceReadSize = rightN
            } else {
                // write bytes to file
                len2, e1 := out.Write(buffer[0:len1])
                len3, e2 := md.Write(buffer[0:len1])
                len4, e3 := sliceMd5.Write(buffer[0:len1])
                // write error
                if e1 != nil || e2 != nil || e3 != nil || len2 != len1 || len3 != len1 || len4 != len1 {
                    logger.Error("write out error:", e1, "|", e2)
                    lib_common.CloseAndDeleteTmpFile(out)
                    return errors.New("write out error(0)")
                }
                sliceReadSize += int64(len1)
            }
            readBodySize += uint64(len1)
        } else {
            logger.Error("error read body:", e3)
            lib_common.CloseAndDeleteTmpFile(out)
            // 终止循环
            return e3
        }
    }
}



// 处理文件上传请求
func queryFileHandler(request *bridge.Meta, connBridge *bridge.Bridge) error {
    var queryMeta = &bridge.OperationQueryFileRequest{}
    e1 := json.Unmarshal(request.MetaBody, queryMeta)
    var response = &bridge.OperationQueryFileResponse{}
    if e1 != nil {
        response.Status = bridge.STATUS_INTERNAL_SERVER_ERROR
        response.Exist = false
        // ignore if it write success
        connBridge.SendResponse(response, 0, nil)
        return e1
    }

    var md5 string

    if mat1, _ := regexp.Match("[0-9a-f]{32}", []byte(queryMeta.PathOrMd5)); mat1 {
        md5 = queryMeta.PathOrMd5
    } else if mat2, _ := regexp.Match(app.PATH_REGEX, []byte(queryMeta.PathOrMd5)); mat2 {
        md5 = regexp.MustCompile(app.PATH_REGEX).ReplaceAllString(queryMeta.PathOrMd5, "${3}")
    } else {
        response.Status = bridge.STATUS_NOT_FOUND
        response.Exist = false
        return connBridge.SendResponse(response, 0, nil)
    }

    fid, e4 := lib_service.GetFileId(md5)
    if e4 != nil {// error query file
        response.Status = bridge.STATUS_NOT_FOUND
        response.Exist = false
        return connBridge.SendResponse(response, 0, nil)
    } else {
        if fid == 0 {//file not found
            response.Status = bridge.STATUS_NOT_FOUND
            response.Exist = false
            return connBridge.SendResponse(response, 0, nil)
        } else {//file exists
            response.Status = bridge.STATUS_OK
            response.Exist = true
            return connBridge.SendResponse(response, 0, nil)
        }
    }
}



// 处理文件下载请求
func downloadFileHandler(request *bridge.Meta, buffer []byte, connBridge *bridge.Bridge) error {
    var meta = &bridge.OperationDownloadFileRequest{}
    e1 := json.Unmarshal(request.MetaBody, meta)
    var response = &bridge.OperationDownloadFileResponse{}
    if e1 != nil {
        response.Status = bridge.STATUS_INTERNAL_SERVER_ERROR
        // ignore if it write success
        connBridge.SendResponse(response, 0, nil)
        return e1
    }
    if mat, _ := regexp.Match(app.PATH_REGEX, []byte(meta.Path)); !mat {
        response.Status = bridge.STATUS_NOT_FOUND
        // ignore if it write success
        e2 := connBridge.SendResponse(response, 0, nil)
        if e2 != nil {
            return e2
        }
        return nil
    }
    md5 := regexp.MustCompile(app.PATH_REGEX).ReplaceAllString(meta.Path, "${4}")
    fullFile, e11 := lib_service.GetFullFile(md5)
    if e11 != nil {
        response.Status = bridge.STATUS_INTERNAL_SERVER_ERROR
        // ignore if it write success
        e2 := connBridge.SendResponse(response, 0, nil)
        if e2 != nil {
            return e2
        }
        return nil
    }
    if fullFile == nil {
        response.Status = bridge.STATUS_NOT_FOUND
        // ignore if it write success
        e2 := connBridge.SendResponse(response, 0, nil)
        if e2 != nil {
            return e2
        }
        return nil
    }
    if len(fullFile.Parts) == 0 {
        response.Status = bridge.STATUS_INTERNAL_SERVER_ERROR
        // ignore if it write success
        e2 := connBridge.SendResponse(response, 0, nil)
        if e2 != nil {
            return e2
        }
        return nil
    }
    var fileSize int64 = 0
    for i := range fullFile.Parts {
        fileSize += fullFile.Parts[i].FileSize
    }

    response.Status = bridge.STATUS_OK
    return connBridge.SendResponse(response, uint64(fileSize), func(out io.WriteCloser) error {
        var read int64 = 0
        for i := range fullFile.Parts {
            downFile, e14 := file.GetFile(GetFilePathByMd5(fullFile.Parts[i].Md5))
            if e14 != nil {
                return e14
            }
            for {
                len, e2 := downFile.Read(buffer)
                if e2 == nil || e2 == io.EOF {
                    logger.Debug("total bytes:", fileSize, ", write:", read)
                    wl, e5 := out.Write(buffer[0:len])
                    read+=int64(len)
                    if e2 == io.EOF {
                        downFile.Close()
                        break
                    }
                    if e5 != nil || wl != len {
                        downFile.Close()
                        return errors.New("error handle download file")
                    }
                } else {
                    downFile.Close()
                    return e2
                }
            }
        }
        return nil
    })
}


