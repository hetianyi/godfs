package svc

import (
	"container/list"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hetianyi/godfs/api"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/conn"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/file"
	"github.com/hetianyi/gox/httpx"
	"github.com/hetianyi/gox/logger"
	"github.com/hetianyi/gox/uuid"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	httpClient *http.Client
)

func init() {
	httpClient = &http.Client{
		Timeout: time.Second * 20,
	}
}

// StartStorageHttpServer starts an storage http server.
func StartAgentHttpServer(c *common.AgentConfig) {
	r := mux.NewRouter()
	r.HandleFunc("/ul", proxyHttpUpload1).Methods("POST")
	r.HandleFunc("/upload", proxyHttpUpload1).Methods("POST")
	// r.HandleFunc("/upload1", httpUpload).Methods("POST")
	r.HandleFunc("/dl", proxyHttpDownload).Methods("GET")
	r.HandleFunc("/download", proxyHttpDownload).Methods("GET")

	srv := &http.Server{
		Handler:           r,
		Addr:              c.BindAddress + ":" + convert.IntToStr(c.HttpPort),
		ReadHeaderTimeout: time.Second * 15,
		WriteTimeout:      0,
		ReadTimeout:       0,
		MaxHeaderBytes:    1 << 20, // 1MB
	}
	/*go func() {

	}()*/
	logger.Info("http server listening on ", c.BindAddress, ":", c.HttpPort)
	if err := srv.ListenAndServe(); err != nil {
		logger.Fatal(err)
	}
}

// httpUpload1 upload files using golang
func proxyHttpUpload1(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	logger.Debug("accept new upload request")

	increaseCountForTheSecond()

	logger.Debug("begin to upload file")

	fmt.Println(r.URL.Path)

	group := strings.TrimSpace(r.URL.Query().Get("group"))

	headers := r.Header
	copyHeaders := make(map[string]string)
	for k, v := range headers {
		copyHeaders[k] = v[0]
	}

	var exclude = list.New()                  // excluded storage list
	var selectedStorage *common.StorageServer // target server for file uploading.
	var lastErr error
	var lastConn *net.Conn

	gox.Try(func() {
		for {
			// select storage server.
			selectedStorage = clientAPI.SelectStorageServer(group, true, exclude)
			if selectedStorage == nil {
				if lastErr == nil {
					lastErr = api.NoStorageServerErr
				}
				break
			}

			req, err := http.NewRequest("POST", r.RequestURI, r.Body)
			if err != nil {
				logger.Error(err)
				continue
			}

			for k, v := range r.Header {
				for _, h := range v {
					req.Header.Add(k, h)
				}
			}

			resp, err := httpClient.Do(req)
			if err != nil {
				logger.Error(err)
				return
			}

			code := resp.StatusCode

			if code == http.StatusOK {
				logger.Debug("upload success")
			} else {
				logger.Debug("upload failed")
			}
			break
		}
	}, func(e interface{}) {
		lastErr = e.(error)
		panic(lastErr)
	})
	// lastConn should be returned and set to nil.
	if lastConn != nil {
		conn.ReturnConnection(selectedStorage, lastConn, nil, true)
	}
}

// httpDownload handles http file upload.
func proxyHttpDownload(w http.ResponseWriter, r *http.Request) {
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

	filePath := strings.Join([]string{common.InitializedStorageConfiguration.DataDir, info.Path}, "/")
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		logger.Debug("error stat file: ", info.Path, ": ", err)
		util.HttpFileNotFoundError(w)
		return
	}
	if fileInfo.Size() < 4 {
		logger.Debug("invalid file length")
		util.HttpInternalServerError(w, "Internal Server Error.")
		return
	}

	outFile, err := file.GetFile(filePath)
	if err != nil {
		logger.Debug("error open file: ", info.Path, ": ", err)
		util.HttpFileNotFoundError(w)
		return
	}
	sr := io.NewSectionReader(outFile, 0, fileInfo.Size()-4)

	if fileName != "" {
		headers.Set("Content-Disposition", "attachment;filename=\""+fileName+"\"")
	} else if fileName == "" && ext != "" {
		fileName = uuid.UUID() + "." + ext
	}
	httpx.ServeContent(w, r, fileName, fileInfo.ModTime(), sr, fileInfo.Size()-4)
}
