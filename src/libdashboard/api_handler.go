package libdashboard

import (
	"app"
	"encoding/json"
	"io"
	"libcommon/bridge"
	"libservice"
	"net/http"
	"strconv"
	"time"
	httputil "util/http"
	"util/logger"
	"util/timeutil"
)

type HttpResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

func validateInsertWebTracker(tracker *bridge.WebTracker) bool {
	if tracker == nil {
		return false
	}
	if tracker.Port <= 0 || tracker.Port > 65535 {
		return false
	}

	if tracker.Secret == "" || tracker.Host == "" || (tracker.Status != 0 && tracker.Status != 1) {
		return false
	}
	return true
}

func validateDeleteWebTracker(tracker *bridge.WebTracker) bool {
	if tracker == nil {
		return false
	}

	if tracker.Id < 1 {
		return false
	}
	return true
}

func addWebTrackerHandler(writer http.ResponseWriter, request *http.Request) {
	logger.Info("perform add web tracker")
	if !httputil.MethodAllow(http.MethodPost, writer, request) {
		return
	}
	cl := request.ContentLength
	if cl > 1024 { // request body too large
		logger.Warn("request body too large")
		return
	}
	buffer, e1 := bridge.MakeBytes(uint64(cl), false, 0, true)
	if e1 != nil {
		logger.Error(e1)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(strconv.Itoa(http.StatusInternalServerError) + " internal server error."))
	} else {
		len, e2 := request.Body.Read(buffer)
		if e2 == nil || e2 == io.EOF {
			if int64(len) != cl {
				logger.Error("error read request body, no more bytes to read")
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(strconv.Itoa(http.StatusInternalServerError) + " incorrect parameter."))
			} else {
				webTracker := &bridge.WebTracker{}
				logger.Info(string(buffer[0:len]))
				e3 := json.Unmarshal(buffer[0:len], webTracker)
				if e3 != nil {
					logger.Error(e3)
				} else {
					if !validateInsertWebTracker(webTracker) {
						logger.Error("parameter error")
						writer.WriteHeader(http.StatusInternalServerError)
						writer.Write([]byte(strconv.Itoa(http.StatusInternalServerError) + " incorrect parameter."))
						return
					}
					e4 := libservice.AddWebTracker(webTracker)
					if e4 != nil {
						logger.Error(e4)
						writer.WriteHeader(http.StatusInternalServerError)
						writer.Write([]byte(strconv.Itoa(http.StatusInternalServerError) + " " + e4.Error() + "."))
					} else {
						resp := &HttpResponse{
							Code:    200,
							Message: "success",
						}
						bs, e5 := MarshalHttpResponseEntity(resp)
						if e5 != nil {
							logger.Error(e5)
							return
						}
						writer.Write(bs)
					}
				}
			}
		}
	}
}

func deleteWebTrackerHandler(writer http.ResponseWriter, request *http.Request) {
	logger.Info("perform delete web tracker")
	if !httputil.MethodAllow(http.MethodDelete, writer, request) {
		return
	}
	cl := request.ContentLength
	if cl > 1024 { // request body too large
		logger.Warn("request body too large")
		return
	}
	buffer, e1 := bridge.MakeBytes(uint64(cl), false, 0, true)
	if e1 != nil {
		logger.Error(e1)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(strconv.Itoa(http.StatusInternalServerError) + " internal server error."))
	} else {
		len, e2 := request.Body.Read(buffer)
		if e2 == nil || e2 == io.EOF {
			if int64(len) != cl {
				logger.Error("error read request body, no more bytes to read")
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(strconv.Itoa(http.StatusInternalServerError) + " incorrect parameter."))
			} else {
				webTracker := &bridge.WebTracker{}
				logger.Info(string(buffer[0:len]))
				e3 := json.Unmarshal(buffer[0:len], webTracker)
				if e3 != nil {
					logger.Error(e3)
				} else {
					if !validateDeleteWebTracker(webTracker) {
						logger.Error("parameter error")
						writer.WriteHeader(http.StatusInternalServerError)
						writer.Write([]byte(strconv.Itoa(http.StatusInternalServerError) + " incorrect parameter."))
						return
					}
					if libservice.UpdateWebTrackerStatus(webTracker.Id, app.STATUS_DELETED, nil) {
						resp := &HttpResponse{
							Code:    200,
							Message: "success",
						}
						bs, e5 := MarshalHttpResponseEntity(resp)
						if e5 != nil {
							logger.Error(e5)
							return
						}
						writer.Write(bs)
					}
				}
			}
		}
	}
}

func indexStatistic(writer http.ResponseWriter, request *http.Request) {
	logger.Info("perform fetch indexStatistic")
	if !httputil.MethodAllow(http.MethodGet, writer, request) {
		return
	}
	indexStatistic, e1 := libservice.GetIndexStatistic()
	if e1 != nil {
		logger.Error(e1)
		logger.Error("query error")
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(strconv.Itoa(http.StatusInternalServerError) + " query error."))
		return
	}
	now := time.Now()
	indexStatistic.UpTime = timeutil.GetLongHumanReadableDuration(time.Unix(app.START_TIME/1000, 0), now)
	resp := &HttpResponse{
		Code:    200,
		Message: "success",
		Result:  *indexStatistic,
	}
	bs, e5 := MarshalHttpResponseEntity(resp)
	if e5 != nil {
		logger.Error(e5)
		return
	}
	writer.Write(bs)
}

func MarshalHttpResponseEntity(i *HttpResponse) ([]byte, error) {
	return json.Marshal(*i)
}
