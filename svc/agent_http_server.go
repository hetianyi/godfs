package svc

import (
	"container/list"
	"github.com/gorilla/mux"
	"github.com/hetianyi/godfs/api"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/logger"
	"io"
	"net/http"
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

	group := strings.TrimSpace(r.URL.Query().Get("group"))

	headers := r.Header
	copyHeaders := make(map[string]string)
	for k, v := range headers {
		copyHeaders[k] = v[0]
	}

	var exclude = list.New()                  // excluded storage list
	var selectedStorage *common.StorageServer // target server for file uploading.
	var lastErr error

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

			logger.Info("agent upload to target server: ", selectedStorage.Host, ":", convert.Uint16ToStr(selectedStorage.HttpPort), "(", selectedStorage.InstanceId, ")")

			req, err := http.NewRequest("POST",
				"http://"+selectedStorage.GetHost()+":"+convert.Uint16ToStr(selectedStorage.HttpPort)+r.RequestURI,
				r.Body)
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

			for k, v := range resp.Header {
				for _, h := range v {
					w.Header().Add(k, h)
				}
			}
			b := make([]byte, resp.ContentLength)
			resp.Body.Read(b)
			w.WriteHeader(code)
			w.Write(b)
			break
		}
	}, func(e interface{}) {
		lastErr = e.(error)
		panic(lastErr)
	})
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
	qs := r.URL.Query()
	if qs != nil {
		fid = qs.Get("id")
		token = qs.Get("token")
		timestamp = qs.Get("timestamp")
		if token == "" {
			token = qs.Get("tk")
		}
		if timestamp == "" {
			timestamp = qs.Get("ts")
		}
	}

	info, curSecret, err := util.ParseAlias(fid, common.InitializedAgentConfiguration.Secret)
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

	initialInstance := false
	if time.Now().Unix()-info.CreateTime < 300 { // 5min
		initialInstance = true
	}

	var exclude = list.New()                  // excluded storage list
	var selectedStorage *common.StorageServer // target server for file uploading.
	var lastErr error

	gox.Try(func() {
		for {
			if initialInstance {
				initialInstance = false
				ins := api.FilterInstanceByInstanceId(info.InstanceId)
				if ins == nil {
					continue
				}
				selectedStorage = &common.StorageServer{
					Server: ins.Server,
					Group:  info.Group,
				}
				logger.Info("download from source server: ", selectedStorage.Host, ":", convert.Uint16ToStr(selectedStorage.HttpPort), "(", selectedStorage.InstanceId, ")")
			}
			if selectedStorage == nil {
				// select storage server.
				selectedStorage = clientAPI.SelectStorageServer(info.Group, false, exclude)
			}
			if selectedStorage == nil {
				if lastErr == nil {
					lastErr = api.NoStorageServerErr
				}
				break
			}

			req, err := http.NewRequest("GET",
				"http://"+selectedStorage.GetHost()+":"+convert.Uint16ToStr(selectedStorage.HttpPort)+r.RequestURI, nil)
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

			for k, v := range resp.Header {
				for _, h := range v {
					w.Header().Set(k, h)
				}
			}
			//b := make([]byte, resp.ContentLength)
			//resp.Body.Read(b)
			w.WriteHeader(code)
			io.Copy(w, resp.Body)
			break
		}
	}, func(e interface{}) {
		lastErr = e.(error)
		panic(lastErr)
	})
}
