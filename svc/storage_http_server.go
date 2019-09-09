package svc

import (
	"github.com/gorilla/mux"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/file"
	"github.com/hetianyi/gox/logger"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	rangeHeader = "^bytes=([0-9]+)-([0-9]+)?$"
)

var (
	compiledRegexpRangeHeader *regexp.Regexp
	gmtLocation, _            = time.LoadLocation("GMT")
)

func init() {
	compiledRegexpRangeHeader = regexp.MustCompile(rangeHeader)
}

// StartStorageHttpServer starts an storage http server.
func StartStorageHttpServer(c *common.StorageConfig) {
	r := mux.NewRouter()
	r.HandleFunc("/", httpUpload).Methods("POST")
	r.HandleFunc("/", httpDownload).Methods("GET")

	srv := &http.Server{
		Handler: r,
		Addr:    c.BindAddress + ":" + convert.IntToStr(c.HttpPort),
		// Good practice: enforce timeouts for servers you create!
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

func httpUpload(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	logger.Debug("accept new request")

}

func httpDownload(w http.ResponseWriter, r *http.Request) {
	logger.Debug("accept download file request")
	defer func() {
		r.Body.Close()
		logger.Debug("download file finish")
	}()

	// TODO check refer
	// TODO check auth
	// TODO cache support

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
	fileName := ""
	ext := ""
	qs := r.URL.Query()
	if qs != nil {
		fid = qs.Get("id")
		fileName = qs.Get("fn")
		ext = file.GetFileExt(fileName)
	}

	info, err := util.ParseAlias(fid)
	if err != nil {
		logger.Debug("error parse alias: ", err)
		fileNotFound(w)
		return
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
		fileNotFound(w)
		return
	}
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		logger.Debug("error stat file: ", info.Path, ": ", err)
		fileNotFound(w)
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
			internalServerError(w)
			return
		}
	}

	headers.Set("Content-Length", convert.Int64ToStr(totalLen))
	headers.Set("Accept-Ranges", "bytes")
	if rangeS != "" {
		headers.Set("Content-Range", "bytes "+convert.Int64ToStr(start)+"-"+convert.Int64ToStr(end-1)+"/"+convert.Int64ToStr(fileInfo.Size()-4))
	}

	if t := common.GetMimeType(ext); common.InitializedStorageConfiguration.EnableMimeTypes && t != "" {
		gmtLocation, _ := time.LoadLocation("GMT")
		headers.Set("Last-Modified", fileInfo.ModTime().In(gmtLocation).Format(time.RFC1123))
		headers.Set("Expires", time.Now().Add(time.Hour*2400).In(gmtLocation).Format(time.RFC1123))
		setMimeHeaders(strings.Split(info.Path, "/")[2], &headers)
		headers.Set("Content-Type", t)
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

func fileNotFound(w http.ResponseWriter) {
	writeErrResponse(w, http.StatusNotFound, "Not Found.")
}

func internalServerError(w http.ResponseWriter) {
	writeErrResponse(w, http.StatusInternalServerError, "Not Found.")
}

// WriteErrResponse write error response
func writeErrResponse(writer http.ResponseWriter, statusCode int, message string) {
	writer.WriteHeader(statusCode)
	writer.Write([]byte(strconv.Itoa(statusCode) + " " + message))
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
