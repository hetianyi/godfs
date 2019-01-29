package libstorage

import (
	"app"
	"bytes"
	"container/list"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	json "github.com/json-iterator/go"
	"hash"
	"io"
	"libcommon"
	"libcommon/bridgev2"
	"libservicev2"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"util/common"
	httputil "util/http"
	"util/logger"
	"util/timeutil"
)

const ContentDispositionPattern = "^Content-Disposition: form-data; name=\"([^\"]*)\"(; filename=\"([^\"]*)\".*)?$"
const ContentTypePattern = "^multipart/form-data; boundary=(.*)$"

// simple test case:
// curl -vF "myfile=@/foo/bar.txt" "http://localhost:8001/upload"

type FileFormReader struct {
	request *http.Request
	buffer  *bytes.Buffer
}

type FileInfo struct {
	Index    int    `json:"index"`
	FileName string `json:"fileName"`
	Path     string `json:"path"`
}

type HttpUploadResponse struct {
	Status   string              `json:"status"`   // handler result status
	FormData map[string][]string `json:"formData"` // form data
	FileInfo []FileInfo          `json:"fileInfo"` // files info for all uploaded file.
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

type FileUploadHandlerV1 struct {
	writer  http.ResponseWriter
	request *http.Request
	params  map[string]*list.List
}

type StageUploadStatus struct {
	readBodySize  uint64
	sliceReadSize int64
	md            hash.Hash
	sliceMd5      hash.Hash
	fileParts     *list.List
	fileName      string
	index         int
	out           *os.File
	path          string
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

// beginUpload begin read request entity and parse form field
func (handler *FileUploadHandlerV1) beginUpload() (*HttpUploadResponse, error) {
	beginTime := time.Now()
	// test code
	// defer increaseTest(1)
	// buff := make([]byte, 10240)
	var formReader = &FileFormReader{
		request: handler.request,
		buffer:  new(bytes.Buffer),
	}

	var ret = &HttpUploadResponse{
		FormData: make(map[string][]string),
	}

	var fileStages list.List
	var fileIndex = 0

	headerContentType := handler.request.Header["Content-Type"]
	contentType := ""
	if headerContentType != nil || len(headerContentType) == 0 {
		contentType = headerContentType[0]
	}
	if mat, _ := regexp.Match(ContentTypePattern, []byte(contentType)); mat {
		boundary := regexp.MustCompile(ContentTypePattern).ReplaceAllString(contentType, "${1}")
		logger.Info("begin read file form, start from "+timeutil.GetShortDateString(beginTime)+" content len is", handler.request.ContentLength/1024, "KB")
		defer func() {
			endTime := time.Now()
			logger.Info("end read file form,   end at     "+timeutil.GetShortDateString(endTime)+" content len is",
				handler.request.ContentLength/1024, "KB, time duration "+timeutil.GetHumanReadableDuration(beginTime, endTime))
		}()
		paraSeparator := "--" + boundary
		endSeparator := "--" + boundary + "--"
		logger.Debug("form boundary is", paraSeparator)
		// calculate md5
		md := md5.New()
		buffer, _ := bridgev2.MakeBytes(app.BufferSize, false, 0, false)
		defer bridgev2.RecycleBytes(buffer)
		for {
			line, e := readNextLine(formReader)
			if e != nil {
				logger.Error("upload error0:", e)
				return nil, e
			}
			// if it is paraSeparator, then start read new form text field or file field
			if paraSeparator == line {
				contentDisposition, e1 := readNextLine(formReader)
				if e1 != nil {
					logger.Error("upload error1:", e1)
					return nil, e1
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
						return nil, e2
					} else {
						logger.Debug("Content-Type ? >>  ", paramContentType)
						if paramContentType == "" { // read text parameter field
							param, e3 := readNextLine(formReader)
							if e3 != nil {
								logger.Error("upload error3:", e3)
								return nil, e3
							} else {
								paramValue = param
								logger.Debug("read text field", paramName, "=", paramValue)
								handler.onTextField(paramName, paramValue)
							}
						} else { // parse content type
							mat2, _ := regexp.Match(ContentDispositionPattern, []byte(contentDisposition))
							if e != nil {
								return nil, e
							}
							fileName := ""
							if mat2 {
								fileName = regexp.MustCompile(ContentDispositionPattern).ReplaceAllString(contentDisposition, "${3}")
								fileName = common.EncodeASCII(fileName)
								logger.Debug("read file field", paramName, "=", fileName)
							}

							_, e3 := readNextLine(formReader)
							if e3 != nil {
								logger.Error("upload error4:", e3)
								return nil, e3
							} else { // read file body
								stageUploadStatus, e4 := readFileBody(formReader, buffer, paraSeparator, md)
								if e4 != nil {
									logger.Error("upload error5:", e4)
									return nil, e4
								}
								fileStages.PushBack(stageUploadStatus)
								stageUploadStatus.index = fileIndex
								handler.onTextField(paramName, stageUploadStatus.path)
								stageUploadStatus.fileName = fileName
								fileIndex++
							}
						}
					}

				}
			} else if endSeparator == line {
				// form stream hit end
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

	fInfo := make([]FileInfo, fileStages.Len())
	k := 0
	for ele := fileStages.Front(); ele != nil; ele = ele.Next() {
		stage := ele.Value.(*StageUploadStatus)
		info := &FileInfo{
			Index:    stage.index,
			FileName: stage.fileName,
			Path:     stage.path,
		}
		fInfo[k] = *info
		k++
	}
	ret.FileInfo = fInfo
	ret.Status = "success"
	return ret, nil
}

func readNextLine(reader *FileFormReader) (string, error) {
	buff, _ := bridgev2.MakeBytes(int64(1), false, 0, false)
	last, _ := bridgev2.MakeBytes(int64(2), false, 0, false)
	defer bridgev2.RecycleBytes(buff)
	defer bridgev2.RecycleBytes(last)
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
			return string(strBuff.Bytes()[0 : strBuff.Len()-2]), nil
		}
	}
}

// readFileBody begin read file part
func readFileBody(reader *FileFormReader, buffer []byte, separator string, md hash.Hash) (*StageUploadStatus, error) {
	defer func() {
		md.Reset()
	}()
	out, oe := libcommon.CreateTmpFile()
	if oe != nil {
		return nil, oe
	}
	stateUploadStatus := &StageUploadStatus{
		readBodySize:  0,
		sliceReadSize: 0,
		sliceMd5:      md5.New(),
		md:            md,
		fileParts:     list.New(),
		out:           out,
	}
	separator = "\r\n" + separator
	buff1 := buffer
	buff2, _ := bridgev2.MakeBytes(int64(len(separator)), true, 1024, true)
	tail, _ := bridgev2.MakeBytes(int64(len(separator)*2), true, 1024, true)
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
			reader.Unread(buff1[i1+2 : len1]) // skip "\r\n"
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
				ByteCopy(tail, 0, len(separator), buff1[len1-len(separator):len1])
			}
			if len2 >= len(separator) {
				ByteCopy(tail, len(separator), len(tail), buff2[0:len(separator)])
			}

			i2 := bytes.Index(tail, []byte(separator))
			if i2 != -1 {
				if i2 < len(separator) {
					e8 := handleStagePartFile(buff1[0:len1-i2], stateUploadStatus)
					if e8 != nil {
						return nil, e8
					}
					reader.Unread(buff1[len1-i2+2 : len1])
					reader.Unread(buff2[0:len2])
				} else {
					e8 := handleStagePartFile(buff1[0:len1], stateUploadStatus)
					if e8 != nil {
						return nil, e8
					}
					reader.Unread(buff2[i2-len(separator)+2 : len2])
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
		e10 := libcommon.MoveTmpFileTo(sMd5, stateUploadStatus.out)
		if e10 != nil {
			return nil, e10
		}

		tmpPart := &app.PartDO{Md5: sMd5, Size: stateUploadStatus.sliceReadSize}
		stateUploadStatus.fileParts.PushBack(tmpPart)
		app.UpdateDiskUsage(stateUploadStatus.sliceReadSize)
	}
	sliceCipherStr := md.Sum(nil)
	sMd5 := hex.EncodeToString(sliceCipherStr)
	logger.Debug("http upload file md5 is", sMd5, "part num:", stateUploadStatus.fileParts.Len())

	finalFile := &app.FileVO{
		Md5:        sMd5,
		PartNumber: stateUploadStatus.fileParts.Len(),
		Group:      app.Group,
		Instance:   app.InstanceId,
		Finish:     1,
		FileSize:   0,
	}
	parts := make([]app.PartDO, stateUploadStatus.fileParts.Len())
	index := 0
	var totalSize int64 = 0
	for ele := stateUploadStatus.fileParts.Front(); ele != nil; ele = ele.Next() {
		parts[index] = *ele.Value.(*app.PartDO)
		totalSize += parts[index].Size
		index++
	}
	finalFile.Parts = parts
	finalFile.FileSize = totalSize
	// stoe := libservice.StorageAddFile(sMd5, app.Group, stateUploadStatus.fileParts)
	stoe := libservicev2.InsertFile(finalFile, nil)
	if stoe != nil {
		return nil, stoe
	}
	// mark the file is multi part or single part
	if stateUploadStatus.fileParts.Len() > 1 {
		stateUploadStatus.path = app.Group + "/" + app.InstanceId + "/M/" + sMd5
	} else {
		stateUploadStatus.path = app.Group + "/" + app.InstanceId + "/S/" + sMd5
	}
	logger.Info("http upload fid is", stateUploadStatus.path)
	app.UpdateUploads()
	return stateUploadStatus, nil
}

func handleStagePartFile(buffer []byte, status *StageUploadStatus) error {
	len1 := int64(len(buffer))
	status.readBodySize += uint64(len1)
	status.md.Write(buffer)
	if status.sliceReadSize+len1 > app.SliceSize {
		// write bytes to file
		leftN := app.SliceSize - status.sliceReadSize
		rightN := int64(len(buffer)) - (app.SliceSize - status.sliceReadSize)
		len2, e1 := status.out.Write(buffer[0:leftN])
		len4, e11 := status.sliceMd5.Write(buffer[0:leftN])
		if e1 != nil || e11 != nil || int64(len2) != leftN || int64(len4) != leftN {
			logger.Error("write out error:", e1, "|", e11)
			libcommon.CloseAndDeleteTmpFile(status.out)
			return errors.New("write out error(0)")
		}

		// close slice file and create a new slice file （承上启下）
		status.out.Close()
		sliceCipherStr := status.sliceMd5.Sum(nil)
		sMd5 := hex.EncodeToString(sliceCipherStr)
		status.sliceMd5.Reset()
		e10 := libcommon.MoveTmpFileTo(sMd5, status.out)
		if e10 != nil {
			return e10
		}
		tmpPart := &app.PartDO{Md5: sMd5, Size: app.SliceSize}
		status.fileParts.PushBack(tmpPart)
		app.UpdateDiskUsage(app.SliceSize)

		out12, e12 := libcommon.CreateTmpFile()
		if e12 != nil {
			return e12
		}
		status.out = out12
		len6, e2 := status.out.Write(buffer[leftN:len1])
		len7, e12 := status.sliceMd5.Write(buffer[leftN:len1])
		if e2 != nil || e12 != nil || int64(len6) != rightN || int64(len7) != rightN {
			logger.Error("write out error:", e2, "|", e12)
			libcommon.CloseAndDeleteTmpFile(status.out)
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
			libcommon.CloseAndDeleteTmpFile(status.out)
			return errors.New("write out error(0)")
		}
		status.sliceReadSize += int64(len1)
	}
	return nil
}

// ByteCopy copy bytes
func ByteCopy(src []byte, start int, end int, cp []byte) {
	for i := range src {
		if i >= start && i < end {
			src[i] = cp[i]
		} else {
			break
		}
	}
}

// WebUploadHandlerV1 web upload entry point
func WebUploadHandlerV1(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	method := request.Method
	if !httputil.MethodAllow(writer, request, http.MethodPost, http.MethodOptions) {
		return
	}

	// client validation
	if !httputil.AccessCheck(writer, request) {
		return
	}

	reqURI := request.URL.Path

	if mat, _ := regexp.Match("/upload/[0-9a-zA-Z_]{1,30}", []byte(reqURI)); mat {
		targetGroup := reqURI[8:]
		if strings.TrimSpace(targetGroup) != "" && app.Group != strings.TrimSpace(targetGroup) {
			logger.Debug("group not match, ignore upload")
			httputil.WriteErrResponse(writer, http.StatusNotFound, "Not Found.")
			return
		}
	}

	// check if client really want to upload file to this group.
	params := request.URL.Query()
	if params != nil {
		targetGroup := params["group"]
		if targetGroup != nil && len(targetGroup) != 0 && strings.TrimSpace(targetGroup[0]) != "" && app.Group != strings.TrimSpace(targetGroup[0]) {
			logger.Debug("group not match, ignore upload")
			httputil.WriteErrResponse(writer, http.StatusNotFound, "Not Found.")
			return
		}
	}

	// handle http options method
	headers := writer.Header()
	origin := ""
	origins := request.Header["Origin"]
	if origins != nil && len(origins) > 0 {
		origin = origins[0]
	}
	if app.CheckOriginAllow(origin) {
		headers.Set("Access-Control-Allow-Origin", origin)
	}
	headers.Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	headers.Set("Access-Control-Allow-Credentials", "true")
	headers.Set("Access-Control-Allow-Headers", "*")
	if method == http.MethodOptions {
		writer.WriteHeader(205)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	handler := &FileUploadHandlerV1{
		writer:  writer,
		request: request,
	}
	common.Try(func() {
		ret, e := handler.beginUpload()
		if e != nil {
			logger.Error("upload error1:", e)
			ret = &HttpUploadResponse{
				Status: "error",
			}
			bs, e1 := json.Marshal(ret)
			if e1 != nil {
				logger.Error("upload error2:", e1)
				httputil.WriteErrResponse(writer, http.StatusInternalServerError, "Internal Server Error.")
			} else {
				handler.writeBack(string(bs))
			}
		} else {
			bs, e1 := json.Marshal(ret)
			if e1 != nil {
				logger.Error("upload error3:", e)
				httputil.WriteErrResponse(writer, http.StatusInternalServerError, "Internal Server Error.")
			} else {
				handler.writeBack(string(bs))
			}
		}
	}, func(i interface{}) {
		writer.WriteHeader(500)
		ret := &HttpUploadResponse{
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

var count = 0
var testLock *sync.Mutex

func init() {
	testLock = new(sync.Mutex)
	// go statisticTest()
}

// code for test
func increaseTest(value int) int {
	testLock.Lock()
	defer testLock.Unlock()
	count += value
	return count
}

func statisticTest() {
	timer := time.NewTicker(time.Second)
	for {
		fmt.Print("\n\n========================================\nupload statistic:" + strconv.Itoa(increaseTest(0)) + "\n========================================\n\n")
		count = 0
		<-timer.C
	}
}
