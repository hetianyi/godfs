package lib_storage

import (
    "net/http"
    "bytes"
    "regexp"
    "io"
    "errors"
    "util/logger"
    "container/list"
    "crypto/md5"
    "hash"
    "app"
    "lib_common"
    "os"
    "encoding/hex"
    "lib_service"
    "encoding/json"
    "lib_common/bridge"
    "util/common"
)

const ContentDispositionPattern = "^Content-Disposition: form-data; name=\"([^\"]*)\"$"
const FileContentDispositionPattern = "^Content-Disposition: form-data; name=\"([^\"]*)\"; filename=\"([^\"]*)\".*$"
const ContentTypePattern = "^multipart/form-data; boundary=(.*)$"

// simple test case:
// curl -vF "myfile=@/foo/bar.txt" "http://localhost:8001/upload"

type FileFormReader struct {
    request *http.Request
    buffer *bytes.Buffer
}
type HttpUploadResponse struct {
    Status string `json:"status"`     // 分片所属文件的id
    FormData map[string][]string `json:"formData"`     // 分片所属文件的id
}

func (reader *FileFormReader) Unread(read []byte) {
    reader.buffer.Write(read)
}

func (reader *FileFormReader) Read(buff []byte) (int, error) {
    if reader.buffer.Len() > 0 {
       if len(buff) <= reader.buffer.Len() {
           return reader.buffer.Read(buff)
       } else {
           reStartPos, e := reader.buffer.Read(buff)
           // read directly from reader
           len, e := reader.request.Body.Read(buff[reStartPos:])
           if e != nil {
               if e != io.EOF {
                   return 0, e
               }
           }
           return reStartPos + len, nil
       }
    }
    // read directly from reader
    return reader.request.Body.Read(buff)
}



var backSpace = []byte{13, 10}

//文件上传处理类
//使用者只需继承此类，然后在公用方法里面参与上传事件的处理即可<br>
//最简单的过程即为覆写onFileField()函数
type FileUploadHandlerV1 struct {
    writer http.ResponseWriter
    request *http.Request
    params map[string]*list.List
}


func (handler *FileUploadHandlerV1) writeBack(response string) {
    handler.writer.Write([]byte(response))
}

func (handler *FileUploadHandlerV1) onTextField(name string, value string) {
    logger.Debug("text field:>>", name, "=", value)
    if handler.params == nil {
        handler.params = make(map[string]*list.List)
    }
    ls := handler.params[name]
    if ls == nil {
        ls = new(list.List)
    }
    ls.PushBack(value)
    handler.params[name] = ls
}

// begin read request entity and parse form field
func (handler *FileUploadHandlerV1) beginUpload() (*HttpUploadResponse, error) {

    logger.Info("begin read file form, content len is", handler.request.ContentLength/1024, "KB")
    //buff := make([]byte, 10240)
    var formReader = &FileFormReader{
        request: handler.request,
        buffer: new(bytes.Buffer),
    }

    var ret = &HttpUploadResponse {
        FormData: make(map[string][]string),
    }

    var fileStages list.List

    headerContentType := handler.request.Header["Content-Type"]
    contentType := ""
    if headerContentType != nil || len(headerContentType) == 0 {
        contentType = headerContentType[0]
    }
    if mat, _ := regexp.Match(ContentTypePattern, []byte(contentType)); mat {
        boundary := regexp.MustCompile(ContentTypePattern).ReplaceAllString(contentType, "${1}")
        paraSeparator := "--" + boundary
        endSeparator := "--" + boundary + "--"
        // calculate md5
        md := md5.New()
        buffer, _ := bridge.MakeBytes(app.BUFF_SIZE, false, 0)
        for {
            line, e := readNextLine(formReader)
            //logger.Debug(">>>>>"+line)
            if e != nil {
                logger.Error("upload error0:", e)
                break
            }
            // if it is paraSeparator, then start read new form text field or file field
            if paraSeparator == line {
                contentDisposition, e1 := readNextLine(formReader)
                if e1 != nil {
                    logger.Error("upload error1:", e1)
                    break
                } else {
                    logger.Debug("contentDisposition>>  ", contentDisposition)
                    mat1, e := regexp.Match(ContentDispositionPattern, []byte(contentDisposition))
                    if e != nil {
                        return nil, e
                    }
                    paramName := ""
                    paramValue := ""
                    if mat1 {
                        paramName = regexp.MustCompile(ContentDispositionPattern).ReplaceAllString(contentDisposition, "${1}")
                    }

                    paramContentType, e2 := readNextLine(formReader)
                    if e2 != nil {
                        logger.Error("upload error2:", e2)
                        break
                    } else {
                        logger.Debug("Content-Type ? >>  ", paramContentType)
                        if paramContentType == "" { // read text parameter field
                            param, e3 := readNextLine(formReader)
                            if e3 != nil {
                                logger.Error("upload error3:", e3)
                                break
                            } else {
                                //logger.Debug("Text parameter >>  ", param)
                                paramValue = param
                                handler.onTextField(paramName, paramValue)
                            }
                        } else { // parse content type

                            mat2, _ := regexp.Match(FileContentDispositionPattern, []byte(contentDisposition))
                            if e != nil {
                                return nil, e
                            }
                            fileName := ""
                            if mat2 {
                                fileName = regexp.MustCompile(FileContentDispositionPattern).ReplaceAllString(contentDisposition, "${2}")
                            }

                            _, e3 := readNextLine(formReader)
                            if e3 != nil {
                                logger.Error("upload error4:", e3)
                                break
                            } else { // read file body
                                stageUploadStatus, e4 := readFileBody(formReader, buffer, paraSeparator, md)
                                if e4 != nil {
                                    logger.Error("upload error5:", e4)
                                    break
                                }
                                fileStages.PushBack(stageUploadStatus)
                                if fileName != "" {
                                    handler.onTextField(fileName, stageUploadStatus.path)
                                }
                            }
                        }
                    }

                }
            } else if endSeparator == line {
                // form stream hit end
                logger.Info("upload end.")
                break
            } else {
                logger.Error("unknown line")
            }
        }
    }
    // copy string list to string[]
    for k, v := range handler.params {
        if v != nil {
            tmp := make([]string, v.Len())
            index := 0
            for ele := v.Front(); ele != nil; ele = ele.Next() {
                tmp[index] = ele.Value.(string)
                index++
            }
            ret.FormData[k] = tmp
        }
    }
    ret.Status = "success"
    return ret, nil
}

// TODO fix in case of memory fragmentation
func readNextLine(reader *FileFormReader) (string, error) {
    buff := make([]byte, 1)
    last := make([]byte, 2)
    var strBuff bytes.Buffer
    for {
        len, e := reader.Read(buff)
        if e != nil {
            if e != io.EOF {
                return "", e
            }
        }
        if len != 1 {
            return "", errors.New("error read from stream")
        }
        app.UpdateIOIN(int64(len))
        last[0] = last[1]
        last[1] = buff[0]
        strBuff.Write(buff)
        if bytes.Equal(backSpace, last) {
            return string(strBuff.Bytes()[0:strBuff.Len() - 2]), nil
        }
    }
}



type StageUploadStatus struct {
    readBodySize uint64
    sliceReadSize int64
    md hash.Hash
    sliceMd5 hash.Hash
    sliceIds *list.List
    out *os.File
    path string
}

// begin read file part
func readFileBody(reader *FileFormReader, buffer []byte, separator string, md hash.Hash) (*StageUploadStatus, error) {
    defer func() {
        md.Reset()
    }()
    out, oe := lib_common.CreateTmpFile()
    if oe != nil {
        return nil, oe
    }
    stateUploadStatus := &StageUploadStatus {
        readBodySize: 0,
        sliceReadSize: 0,
        sliceMd5: md5.New(),
        md: md,
        sliceIds: list.New(),
        out: out,
    }
    separator = "\r\n" + separator
    buff1 := buffer
    // TODO fix in case of memory fragmentation
    buff2, _ := bridge.MakeBytes(uint64(len(separator)), true, 1024)
    tail, _ := bridge.MakeBytes(uint64(len(separator)*2), true, 1024)
    for {
        len1, e1 := reader.Read(buff1)
        if e1 != nil {
            if e1 != io.EOF {
                return nil, e1
            }
        }
        if len1 == 0 {
            return nil, errors.New("read file body failed1")
        }
        app.UpdateIOIN(int64(len1))
        // whether buff1 contains separator
        i1 := bytes.Index(buff1, []byte(separator))
        if i1 != -1 {
            out.Write(buff1[0:i1])
            e8 := handleStagePartFile(buff1[0:i1], stateUploadStatus)
            if e8 != nil {
                return nil, e8
            }
            reader.Unread(buff1[i1 + 2:len1]) // skip "\r\n"
            break
        } else {
            len2, e2 := reader.Read(buff2)
            if e2 != nil {
                if e2 != io.EOF {
                    return nil, e2
                }
            }
            if len2 == 0 {
                return nil, errors.New("read file body failed2")
            }
            app.UpdateIOIN(int64(len2))
            // []byte tail is last bytes of buff1 and first bytes of buff2 in case of broken separator.
            if len1 >= len(separator) {
                ByteCopy(tail, 0, len(separator), buff1[len1 - len(separator):len1])
            }
            if len2 >= len(separator) {
                ByteCopy(tail, len(separator), len(tail), buff2[0:len(separator)])
            }

            i2 := bytes.Index(tail, []byte(separator))
            if i2 != -1 {
                if i2 < len(separator) {
                    e8 := handleStagePartFile(buff1[0:len1 - i2], stateUploadStatus)
                    if e8 != nil {
                        return nil, e8
                    }
                    reader.Unread(buff1[len1 - i2 + 2:len1])
                    reader.Unread(buff2[0:len2])
                } else {
                    e8 := handleStagePartFile(buff1[0:len1], stateUploadStatus)
                    if e8 != nil {
                        return nil, e8
                    }
                    reader.Unread(buff2[i2 - len(separator) + 2:len2])
                }
                break
            } else {
                e8 := handleStagePartFile(buff1[0:len1], stateUploadStatus)
                if e8 != nil {
                    return nil, e8
                }
                reader.Unread(buff2[0:len2])
            }
        }
    }
    stateUploadStatus.out.Close()
    if stateUploadStatus.sliceReadSize > 0 {
        sliceCipherStr := stateUploadStatus.sliceMd5.Sum(nil)
        sMd5 := hex.EncodeToString(sliceCipherStr)
        stateUploadStatus.sliceMd5.Reset()
        e10 := lib_common.MoveTmpFileTo(sMd5, stateUploadStatus.out)
        if e10 != nil {
            return nil, e10
        }
        // save slice info to db
        pid, e8 := lib_service.AddPart(sMd5, stateUploadStatus.sliceReadSize)
        if e8 != nil {
            return nil, e8
        }
        stateUploadStatus.sliceIds.PushBack(pid)
    }
    sliceCipherStr := md.Sum(nil)
    sMd5 := hex.EncodeToString(sliceCipherStr)
    logger.Debug("http upload file md5 is", sMd5, "part num:", stateUploadStatus.sliceIds.Len())
    stoe := lib_service.StorageAddFile(sMd5, app.GROUP, stateUploadStatus.sliceIds)
    if stoe != nil {
        return nil, stoe
    }
    // mark the file is multi part or single part
    if stateUploadStatus.sliceIds.Len() > 1 {
        stateUploadStatus.path = app.GROUP + "/" + app.INSTANCE_ID + "/M/" + sMd5
    } else {
        stateUploadStatus.path = app.GROUP + "/" + app.INSTANCE_ID + "/S/" + sMd5
    }
    logger.Debug("http upload fid is", stateUploadStatus.path)
    app.UpdateUploads()
    return stateUploadStatus, nil
}

func handleStagePartFile(buffer []byte, status *StageUploadStatus) error {
    len1 := int64(len(buffer))
    status.readBodySize += uint64(len1)
    status.md.Write(buffer)
    if status.sliceReadSize + len1 > app.SLICE_SIZE {
        // write bytes to file
        leftN := app.SLICE_SIZE - status.sliceReadSize
        rightN := int64(len(buffer)) - (app.SLICE_SIZE - status.sliceReadSize)
        len2, e1 := status.out.Write(buffer[0:leftN])
        len4, e11 := status.sliceMd5.Write(buffer[0:leftN])
        if e1 != nil || e11 != nil || int64(len2) != leftN || int64(len4) != leftN {
            logger.Error("write out error:", e1, "|", e11)
            lib_common.CloseAndDeleteTmpFile(status.out)
            return errors.New("write out error(0)")
        }

        // close slice file and create a new slice file （承上启下）
        status.out.Close()
        sliceCipherStr := status.sliceMd5.Sum(nil)
        sMd5 := hex.EncodeToString(sliceCipherStr)
        status.sliceMd5.Reset()
        e10 := lib_common.MoveTmpFileTo(sMd5, status.out)
        if e10 != nil {
            return e10
        }
        // save slice info to db
        pid, e8 := lib_service.AddPart(sMd5, app.SLICE_SIZE)
        if e8 != nil {
            return e8
        }
        status.sliceIds.PushBack(pid)

        out12, e12 := lib_common.CreateTmpFile()
        if e12 != nil {
            return e12
        }
        status.out = out12
        len6, e2 := status.out.Write(buffer[leftN:len1])
        len7, e12 := status.sliceMd5.Write(buffer[leftN:len1])
        if e2 != nil || e12 != nil || int64(len6) != rightN || int64(len7) != rightN {
            logger.Error("write out error:", e2, "|", e12)
            lib_common.CloseAndDeleteTmpFile(status.out)
            return errors.New("write out error(1)")
        }
        status.sliceReadSize = rightN
    } else {
        // write bytes to file
        len2, e1 := status.out.Write(buffer[0:len1])
        len4, e3 := status.sliceMd5.Write(buffer[0:len1])
        // write error
        if e1 != nil || e3 != nil || int64(len2) != len1 || int64(len4) != len1 {
            logger.Error("write out error:", e1)
            lib_common.CloseAndDeleteTmpFile(status.out)
            return errors.New("write out error(0)")
        }
        status.sliceReadSize += int64(len1)
    }
    return nil
}

// copy bytes
func ByteCopy(src []byte, start int, end int, cp []byte) {
    for i := range src {
        if i >= start && i < end {
            src[i] = cp[i]
        } else {
            break
        }
    }
}


// web upload entry point
func WebUploadHandlerV1(writer http.ResponseWriter, request *http.Request) {

    if app.HTTP_AUTH != "" {
        user, pass, _ := request.BasicAuth()
        if app.HTTP_AUTH != user + ":" + pass {
            writer.WriteHeader(403)
            writer.Write([]byte("403 Forbidden."))
            return
        }
    }

    writer.Header().Set("Content-Type", "application/json;charset=UTF-8")
    handler := &FileUploadHandlerV1{
        writer: writer,
        request: request,
    }
    common.Try(func() {
        ret, e := handler.beginUpload()
        if e != nil {
            logger.Error("upload error1:", e)
            ret = &HttpUploadResponse {
                Status: "error",
            }
            bs, e1 := json.Marshal(ret)
            if e1 != nil {
                logger.Error("upload error2:", e1)
            } else {
                handler.writeBack(string(bs))
            }
        }
        bs, e1 := json.Marshal(ret)
        if e1 != nil {
            logger.Error("upload error3:", e)
        } else {
            handler.writeBack(string(bs))
        }
    }, func(i interface{}) {
        ret := &HttpUploadResponse {
            Status: "error",
        }
        bs, e1 := json.Marshal(ret)
        if e1 != nil {
            logger.Error("upload error2:", e1)
        } else {
            handler.writeBack(string(bs))
        }
    })
}