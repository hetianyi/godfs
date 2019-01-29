package http

import (
	"app"
	"net/http"
	"strconv"
)

// MethodAllow tell if http method is allowed
func MethodAllow(writer http.ResponseWriter, request *http.Request, expectMethod ...string) bool {
	method := request.Method
	match := false
	for i := range expectMethod {
		m := expectMethod[i]
		if m == method {
			match = true
			break
		}
	}
	if !match {
		WriteErrResponse(writer, http.StatusMethodNotAllowed, "Method '"+method+"' not allowed.")
		return false
	}
	return true
}

// AccessCheck check http basic auth
func AccessCheck(writer http.ResponseWriter, request *http.Request) bool {
	if app.HttpAuth != "" {
		user, pass, _ := request.BasicAuth()
		if app.HttpAuth == user+":"+pass {
			return true
		}
		WriteErrResponse(writer, http.StatusForbidden, "Forbidden.")
		return false
	}
	return true
}

// WriteErrResponse write error response
func WriteErrResponse(writer http.ResponseWriter, statusCode int, message string) {
	writer.WriteHeader(statusCode)
	writer.Write([]byte(strconv.Itoa(statusCode) + " " + message))
}
