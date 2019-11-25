package svc

import (
	"bytes"
	"container/list"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/hetianyi/godfs/binlog"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/file"
	"github.com/hetianyi/gox/httpx"
	"github.com/hetianyi/gox/logger"
	"github.com/hetianyi/gox/uuid"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	rangeHeader        = "^bytes=([0-9]+)-([0-9]+)?$"
	FORM_TEXT          = "text"
	FORM_FILE          = "file"
	ContentTypePattern = "^multipart/form-data; boundary=(.*)$"
)

var (
	compiledRegexpRangeHeader *regexp.Regexp
	gmtLocation, _            = time.LoadLocation("GMT")
	RegexContentTypePattern   = regexp.MustCompile(ContentTypePattern)
)

type FormEntry struct {
	Index          int    `json:"index"`
	Type           string `json:"type"`
	ParameterName  string `json:"name"`
	ParameterValue string `json:"value"`
	Size           int64  `json:"size,omitempty"`
	Group          string `json:"group,omitempty"`
	InstanceId     string `json:"instanceId,omitempty"`
	Md5            string `json:"md5,omitempty"`
	FileId         string `json:"fileId,omitempty"`
}

func init() {
	compiledRegexpRangeHeader = regexp.MustCompile(rangeHeader)
}

// StartStorageHttpServer starts an storage http server.
func StartStorageHttpServer(c *common.StorageConfig) {
	r := mux.NewRouter()
	r.HandleFunc("/up", httpUpload).Methods("POST")
	r.HandleFunc("/upload", httpUpload1).Methods("POST")
	r.HandleFunc("/upload1", httpUpload).Methods("POST")
	r.HandleFunc("/dl", httpDownload).Methods("GET")
	r.HandleFunc("/download", httpDownload).Methods("GET")

	srv := &http.Server{
		Handler:           r,
		Addr:              c.BindAddress + ":" + convert.IntToStr(c.HttpPort),
		ReadHeaderTimeout: time.Second * 15,
		WriteTimeout:      0,
		ReadTimeout:       0,
		MaxHeaderBytes:    1 << 20, // 1MB
	}
	go func() {
		logger.Info("http server listening on ", c.BindAddress, ":", c.HttpPort)
		if err := srv.ListenAndServe(); err != nil {
			logger.Fatal(err)
		}
	}()
}

// httpUpload handles http file upload.
func httpUpload(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	logger.Debug("accept new upload request")

	// file is private or public
	s := strings.TrimSpace(r.URL.Query().Get("s"))
	isPrivate := common.InitializedStorageConfiguration.PublicAccessMode
	if s == "false" || s == "0" {
		isPrivate = false
	} else if s == "true" || s == "1" {
		isPrivate = true
	}

	// formEntries stores form's text fields and file fields.
	formEntries := list.New()
	var result = make(map[string]interface{})
	result["accessMode"] = gox.TValue(isPrivate, "private", "public")

	// define a handler for file upload.
	handler := &httpx.FileUploadHandler{
		Request: r,
	}

	formEntryIndex := 0

	// handle form text field.
	handler.OnFormField = func(paraName, paraValue string) {
		logger.Debug("form parameter: name=", paraName, ", value=", paraValue)
		formEntryIndex++
		formEntries.PushBack(FormEntry{
			Index:          formEntryIndex,
			Type:           FORM_TEXT,
			ParameterName:  paraName,
			ParameterValue: paraValue,
			Size:           0,
		})
	}

	// handle form file field.
	handler.OnFileField = func(paraName, fileName string) *httpx.FileTransactionProcessor {
		tmpFileName := ""
		var out *os.File
		var proxy *DigestProxyWriter
		return &httpx.FileTransactionProcessor{
			Before: func() error {
				tmpFileName = common.InitializedStorageConfiguration.TmpDir + "/" + uuid.UUID()
				o, err := file.CreateFile(tmpFileName)
				if err != nil {
					return err
				}
				out = o
				proxy = &DigestProxyWriter{
					crcH: util.CreateCrc32Hash(),
					md5H: util.CreateMd5Hash(),
					out:  out,
				}
				return nil
			},
			Error: func(err error) {
				// close tmp file and delete it.
				if out != nil {
					out.Close()
				}
				file.Delete(tmpFileName)
			},
			Success: func() error {
				logger.Debug("write tail")
				// write reference count mark.
				_, err := out.Write(tailRefCount)
				if err != nil {
					return err
				}
				out.Close()

				// get file info.
				fInfo, err := os.Stat(tmpFileName)
				if err != nil {
					return err
				}

				// get crc and md5.
				crc32String := util.GetCrc32HashString(proxy.crcH)
				md5String := util.GetMd5HashString(proxy.md5H)

				// build target dir and fileId.
				targetDir := strings.ToUpper(strings.Join([]string{crc32String[len(crc32String)-4 : len(crc32String)-2], "/",
					crc32String[len(crc32String)-2:]}, ""))
				targetLoc := common.InitializedStorageConfiguration.DataDir + "/" + targetDir
				targetFile := common.InitializedStorageConfiguration.DataDir + "/" + targetDir + "/" + md5String
				finalFileId := common.InitializedStorageConfiguration.Group + "/" + targetDir + "/" + md5String
				logger.Debug("create alias")
				finalFileId = util.CreateAlias(finalFileId, common.InitializedStorageConfiguration.InstanceId, isPrivate, time.Now())

				if !file.Exists(targetLoc) {
					if err := file.CreateDirs(targetLoc); err != nil {
						return err
					}
				}
				if !file.Exists(targetFile) {
					logger.Debug("file not exists, move to target dir.")
					if err := file.MoveFile(tmpFileName, targetFile); err != nil {
						return err
					}
				} else {
					logger.Debug("file already exists, increasing reference count.")
					// increase file reference count.
					if err = updateFileReferenceCount(targetFile, 1); err != nil {
						return err
					}
				}
				// write binlog.
				logger.Debug("writing binlog...")
				if err = writableBinlogManager.Write(binlog.CreateLocalBinlog(finalFileId,
					fInfo.Size()-int64(len(tailRefCount)), common.InitializedStorageConfiguration.InstanceId, time.Now(), 1)); err != nil {
					return errors.New("error writing binlog: " + err.Error())
				}

				logger.Debug("add dataset...")
				if err := Add(finalFileId); err != nil {
					return err
				}
				logger.Debug("add dataset success")

				// append form entry.
				formEntryIndex++
				formEntries.PushBack(FormEntry{
					Index:          formEntryIndex,
					Type:           FORM_FILE,
					ParameterName:  paraName,
					ParameterValue: fileName,
					Size:           fInfo.Size() - int64(len(tailRefCount)),
					Group:          common.InitializedStorageConfiguration.Group,
					InstanceId:     common.InitializedStorageConfiguration.InstanceId,
					FileId:         finalFileId,
				})
				return nil
			},
			Write: func(bs []byte) error {
				_, err := proxy.Write(bs)
				return err
			},
		}
	}

	// begin to parse form.
	if err := handler.Parse(); err != nil {
		logger.Error("error upload files: ", err)
	}

	// result form field entries
	formEntriesArray := make([]FormEntry, formEntries.Len())
	i := 0
	gox.WalkList(formEntries, func(item interface{}) bool {
		formEntriesArray[i] = item.(FormEntry)
		i++
		return false
	})

	result["form"] = formEntriesArray
	retJSON, err := json.Marshal(result)
	if err != nil {
		logger.Debug(err)
		util.HttpInternalServerError(w, "Internal Server Error.")
		return
	}

	// write response.
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	util.HttpWriteResponse(w, http.StatusOK, string(retJSON))
}

// httpUpload1 upload files using golang
func httpUpload1(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	logger.Debug("accept new upload request")

	// file is private or public
	s := strings.TrimSpace(r.URL.Query().Get("s"))
	isPrivate := common.InitializedStorageConfiguration.PublicAccessMode
	if s == "false" || s == "0" {
		isPrivate = false
	} else if s == "true" || s == "1" {
		isPrivate = true
	}

	// formEntries stores form's text fields and file fields.
	formEntries := list.New()
	var result = make(map[string]interface{})
	result["accessMode"] = gox.TValue(isPrivate, "private", "public")

	// get form boundary
	headerContentType := r.Header["Content-Type"]
	var contentType, boundary string
	if headerContentType != nil && len(headerContentType) > 0 {
		contentType = headerContentType[0]
	}
	if RegexContentTypePattern.Match([]byte(contentType)) {
		boundary = RegexContentTypePattern.ReplaceAllString(contentType, "${1}")
	}

	reader := multipart.NewReader(r.Body, boundary)

	var buffer = new(bytes.Buffer)
	var lastErr error
	formEntryIndex := 0

	for {
		buffer.Reset()
		p, err := reader.NextPart()
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Debug(err)
			lastErr = err
			break
		}

		// read text field.
		if strings.TrimSpace(p.FileName()) == "" {
			if _, err := io.Copy(buffer, p); err != nil {
				logger.Debug(err)
				lastErr = err
				break
			}
			formEntryIndex++
			formEntries.PushBack(FormEntry{
				Index:          formEntryIndex,
				Type:           FORM_TEXT,
				ParameterName:  p.FormName(),
				ParameterValue: string(buffer.Bytes()),
				Size:           0,
			})
			continue
		}

		// read file field.
		tmpFileName := common.InitializedStorageConfiguration.TmpDir + "/" + uuid.UUID()
		out, err := file.CreateFile(tmpFileName)
		if err != nil {
			logger.Debug(err)
			lastErr = err
			break
		}

		clean := func() {
			out.Close()
			file.Delete(tmpFileName)
		}
		proxy := &DigestProxyWriter{
			crcH: util.CreateCrc32Hash(),
			md5H: util.CreateMd5Hash(),
			out:  out,
		}
		n, err := io.Copy(proxy, p)
		if err != nil {
			logger.Debug(err)
			lastErr = err
			clean()
			break
		}
		logger.Debug("write tail")
		// write reference count mark.
		_, err = out.Write(tailRefCount)
		if err != nil {
			logger.Debug(err)
			lastErr = err
			clean()
			break
		}
		out.Close()

		// TODO duplicate code
		// get crc and md5.
		crc32String := util.GetCrc32HashString(proxy.crcH)
		md5String := util.GetMd5HashString(proxy.md5H)

		// build target dir and fileId.
		targetDir := strings.ToUpper(strings.Join([]string{crc32String[len(crc32String)-4 : len(crc32String)-2], "/",
			crc32String[len(crc32String)-2:]}, ""))
		targetLoc := common.InitializedStorageConfiguration.DataDir + "/" + targetDir
		targetFile := common.InitializedStorageConfiguration.DataDir + "/" + targetDir + "/" + md5String
		finalFileId := common.InitializedStorageConfiguration.Group + "/" + targetDir + "/" + md5String

		logger.Debug("create alias")

		finalFileId = util.CreateAlias(finalFileId, common.InitializedStorageConfiguration.InstanceId, isPrivate, time.Now())

		if !file.Exists(targetLoc) {
			if err := file.CreateDirs(targetLoc); err != nil {
				logger.Debug(err)
				lastErr = err
				clean()
				break
			}
		}
		if !file.Exists(targetFile) {
			logger.Debug("file not exists, move to target dir.")
			if err := file.MoveFile(tmpFileName, targetFile); err != nil {
				logger.Debug(err)
				lastErr = err
				clean()
				break
			}
		} else {
			logger.Debug("file already exists, increasing reference count.")
			// increase file reference count.
			if err = updateFileReferenceCount(targetFile, 1); err != nil {
				logger.Debug(err)
				lastErr = err
				clean()
				break
			}
		}

		// write binlog.
		logger.Debug("write binlog...")
		if err = writableBinlogManager.Write(binlog.CreateLocalBinlog(finalFileId,
			n, common.InitializedStorageConfiguration.InstanceId, time.Now(), 1)); err != nil {
			lastErr = errors.New("error writing binlog: " + err.Error())
			logger.Debug(lastErr)
			clean()
			break
		}

		logger.Debug("add dataset...")
		if err := Add(finalFileId); err != nil {
			logger.Debug(err)
			lastErr = err
			clean()
			break
		}
		logger.Debug("add dataset success")

		// append form entry.
		formEntryIndex++
		formEntries.PushBack(FormEntry{
			Index:          formEntryIndex,
			Type:           FORM_FILE,
			ParameterName:  p.FormName(),
			ParameterValue: p.FileName(),
			Size:           n,
			Group:          common.InitializedStorageConfiguration.Group,
			InstanceId:     common.InitializedStorageConfiguration.InstanceId,
			Md5:            md5String,
			FileId:         finalFileId,
		})
	}

	if lastErr != nil {
		util.HttpInternalServerError(w, "Internal Server Error")
		return
	}

	// result form field entries
	formEntriesArray := make([]FormEntry, formEntries.Len())
	i := 0
	gox.WalkList(formEntries, func(item interface{}) bool {
		formEntriesArray[i] = item.(FormEntry)
		i++
		return false
	})

	result["form"] = formEntriesArray
	retJSON, err := json.Marshal(result)
	if err != nil {
		logger.Debug(err)
		util.HttpInternalServerError(w, "Internal Server Error")
		return
	}

	// write response.
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	util.HttpWriteResponse(w, http.StatusOK, string(retJSON))
}

// httpDownload handles http file upload.
func httpDownload(w http.ResponseWriter, r *http.Request) {
	logger.Debug("accept download file request")
	defer func() {
		r.Body.Close()
		logger.Debug("download file finish")
	}()

	// TODO check refer

	// handle http options method
	headers := w.Header()
	// download method must be GET or OPTIONS
	method := r.Method
	headers.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	headers.Set("Access-Control-Allow-Credentials", "true")
	headers.Set("Access-Control-Allow-Headers", "*")
	if method == http.MethodOptions {
		w.WriteHeader(205)
		return
	}

	fid := ""
	token := ""
	timestamp := ""
	fileName := ""
	// if EnableMimeTypes is on and fileName is not empty,
	// EnableMimeTypes will be ignored.
	ext := ""
	qs := r.URL.Query()
	if qs != nil {
		fid = qs.Get("id")
		fileName = qs.Get("fn")
		ext = qs.Get("type")
		token = qs.Get("token")
		timestamp = qs.Get("timestamp")
		// custom content type
		if fileName == "" {
			fileName = qs.Get("fileName")
		}
		if fileName != "" {
			ext = file.GetFileExt(fileName)
		}
		if token == "" {
			token = qs.Get("tk")
		}
		if timestamp == "" {
			timestamp = qs.Get("ts")
		}
	}

	// query and determine if the file exists.
	if c, err := Contains(fid); !c || err != nil {
		logger.Debug("error query fileId: ", c, "<->", err)
		util.HttpFileNotFoundError(w)
		return
	}

	info, curSecret, err := util.ParseAlias(fid, common.InitializedStorageConfiguration.Secret)
	if err != nil {
		logger.Debug("error parse alias: ", err)
		util.HttpFileNotFoundError(w)
		return
	}

	// check token
	if info.IsPrivate {
		if len(token) != 32 || timestamp == "" {
			util.HttpForbiddenError(w, "Forbidden.")
			return
		}
		cToken := util.GenerateToken(fid, curSecret, timestamp)
		nts, err := convert.StrToInt64(timestamp)
		if err != nil {
			util.HttpForbiddenError(w, "Forbidden.")
			return
		}
		if token != cToken || nts < gox.GetTimestamp(time.Now()) {
			util.HttpForbiddenError(w, "Forbidden.")
			return
		}
	}

	// 304 Not Modified
	eTag := r.Header["If-None-Match"]
	if common.InitializedStorageConfiguration.EnableMimeTypes &&
		eTag != nil && len(eTag) > 0 && eTag[0] == "\""+strings.Split(info.Path, "/")[2]+"\"" {
		setMimeHeaders(strings.Split(info.Path, "/")[2], &headers)
		w.WriteHeader(304)
		return
	}

	// parse header: range
	rangeS := r.Header.Get("Range")
	start, end := parseHeaderRange(rangeS)

	filePath := strings.Join([]string{common.InitializedStorageConfiguration.DataDir, info.Path}, "/")
	outFile, err := file.GetFile(filePath)
	if err != nil {
		logger.Debug("error open file: ", info.Path, ": ", err)
		util.HttpFileNotFoundError(w)
		return
	}
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		logger.Debug("error stat file: ", info.Path, ": ", err)
		util.HttpFileNotFoundError(w)
		return
	}

	if start <= 0 {
		start = 0
	}
	if start > fileInfo.Size()-4 {
		start = fileInfo.Size() - 4
	}
	if end <= 0 || end >= (fileInfo.Size()-4) || end == start || end < start {
		end = fileInfo.Size() - 4
	}
	totalLen := end - start

	if rangeS != "" {
		if _, err = outFile.Seek(start, 0); err != nil {
			logger.Debug("error seek file: ", info.Path, ": ", err)
			util.HttpInternalServerError(w, "Not Found.")
			return
		}
	}

	headers.Set("Content-Length", convert.Int64ToStr(totalLen))
	headers.Set("Accept-Ranges", "bytes")
	if rangeS != "" {
		headers.Set("Content-Range", "bytes "+convert.Int64ToStr(start)+"-"+convert.Int64ToStr(end-1)+"/"+convert.Int64ToStr(fileInfo.Size()-4))
	}

	if fileName == "" && common.InitializedStorageConfiguration.EnableMimeTypes {
		gmtLocation, _ := time.LoadLocation("GMT")
		headers.Set("Last-Modified", fileInfo.ModTime().In(gmtLocation).Format(time.RFC1123))
		headers.Set("Expires", time.Now().Add(time.Hour*2400).In(gmtLocation).Format(time.RFC1123))
		setMimeHeaders(strings.Split(info.Path, "/")[2], &headers)
		headers.Set("Content-Type", common.GetMimeType(ext))
	} else {
		headers.Set("Expires", "0")
		headers.Set("Pragma", "public")
		// headers.Set("Accept-Ranges", "bytes")
		headers.Set("Content-Transfer-Encoding", "binary")
		headers.Set("Cache-Control", "must-revalidate, post-check=0, pre-check=0")
		headers.Set("Content-Disposition", "attachment;filename=\""+fileName+"\"")
	}

	if rangeS == "" {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(206)
	}
	_, err = io.Copy(w, io.LimitReader(outFile, totalLen))
	if err != nil {
		logger.Debug("error write file: ", err)
	}
}

// parseHeaderRange if end is 0, then the end represents max
func parseHeaderRange(rang string) (int64, int64) {
	if rang == "" {
		return 0, 0
	}
	if mat, _ := regexp.Match(rangeHeader, []byte(rang)); !mat {
		return 0, 0
	}
	s := compiledRegexpRangeHeader.ReplaceAllString(rang, "${1}")
	e := compiledRegexpRangeHeader.ReplaceAllString(rang, "${2}")
	uintS, e1 := strconv.ParseInt(s, 10, 64)
	uintE, e2 := strconv.ParseInt(e, 10, 64)
	if e1 != nil {
		return 0, 0
	}
	if e2 != nil {
		return uintS, 0
	}
	return uintS, uintE
}

func setMimeHeaders(md5 string, headers *http.Header) {
	// headers.Set("Cache-Control", "public")
	headers.Set("Cache-Control", "max-age=604800")
	headers.Set("Access-Control-Allow-Origin", "*")
	headers.Set("date", time.Now().In(gmtLocation).Format(time.RFC1123))
	headers.Set("Etag", "\""+md5+"\"")
	// headers.Set("Connection", "keep-alive")
}
