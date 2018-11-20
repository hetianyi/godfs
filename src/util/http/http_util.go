package http

import (
	"bytes"
	"libcommon/bridge"
	"net/http"
	"strconv"
	"util/logger"
)


func GetResponseBodyContent(resp *http.Response) (c string, e error) {
	body := resp.Body
	defer body.Close()
	bs, _ := bridge.MakeBytes(10240, false, 0, false)
	defer bridge.RecycleBytes(bs)
	var buffer bytes.Buffer
	for {
		len, err := body.Read(bs)
		if err == nil {
			buffer.Write(bs[0:len])
		} else {
			defer func() {
				e = err
			}()
			break
		}
	}
	return buffer.String(), e
}

func MethodAllow(expectMethod string, writer http.ResponseWriter, request *http.Request) bool {
	method := request.Method
	if expectMethod != method {
		logger.Warn("405 method not allowed:", request.RequestURI)
		writer.WriteHeader(http.StatusMethodNotAllowed)
		writer.Write([]byte(strconv.Itoa(http.StatusMethodNotAllowed) + " Method '"+ method +"' not allowed."))
		return false
	}
	return true
}

