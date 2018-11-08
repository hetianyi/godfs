package libdashboard

import (
	"net/http"
	httputil "util/http"
	"util/logger"
	"libcommon/bridge"
	"encoding/json"
	"libservice"
	"strconv"
    "io"
    "app"
)

type HttpResponse struct {
	Code    int         `json:"uuid"`
	Message string      `json:"uuid"`
	Result  interface{} `json:"uuid"`
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
		len , e2 := request.Body.Read(buffer)
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
                        resp := HttpResponse{
                            Code: 200,
                            Message: "success",
                        }
                        bs, e5 := httputil.MarshalHttpResponseEntity(resp)
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
        len , e2 := request.Body.Read(buffer)
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
                        resp := HttpResponse{
                            Code: 200,
                            Message: "success",
                        }
                        bs, e5 := httputil.MarshalHttpResponseEntity(resp)
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



