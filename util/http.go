package util

import (
	"net/http"
	"strconv"
)

func HttpFileNotFoundError(w http.ResponseWriter) {
	HttpWriteResponse(w, http.StatusNotFound, "Not Found.")
}

func HttpInternalServerError(w http.ResponseWriter, message string) {
	HttpWriteResponse(w, http.StatusInternalServerError, message)
}

func HttpForbiddenError(w http.ResponseWriter, message string) {
	HttpWriteResponse(w, http.StatusForbidden, message)
}

// HttpWriteResponse writes error response.
func HttpWriteResponse(writer http.ResponseWriter, statusCode int, message string) {
	writer.WriteHeader(statusCode)
	writer.Write([]byte(strconv.Itoa(statusCode) + " " + message))
}
