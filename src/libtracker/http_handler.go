package libtracker

import (
	"net/http"
	"strconv"
	"util/logger"
	"libcommon/bridgev2"
	"libcommon"
	"strings"
	"container/list"
	httputil "util/http"
	"util/common"
	"app"
	"bytes"
	json "github.com/json-iterator/go"
)

const MaxUploadFileSize = 1024*10
var (
	storage_servers_placeholder = "<%storage_servers%>"
	uploadable_servers_placeholder = "<%uploadable_servers%>"
	merged_servers_placeholder = "<%merged_servers%>"
)

// configure nginx template
func ConfigureNginxHandler(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	if !httputil.MethodAllow(writer, request, http.MethodPost, http.MethodOptions) {
		return
	}
	if !httputil.AccessCheck(writer, request) {
		return
	}

	e1 := request.ParseMultipartForm(MaxUploadFileSize)
	if e1 != nil {
		logger.Error("error parse form data:", e1)
		httputil.WriteErrResponse(writer, http.StatusBadRequest, "Bad request.")
		return
	}

	var servers list.List

	for k, v := range request.MultipartForm.Value {
		if k != "servers" {
			continue
		}
		if v == nil || len(v) == 0 {
			continue
		}
		for i := range v {
			var ss []app.StorageDO
			e3 := json.UnmarshalFromString(v[i], &ss)
			if e3 != nil {
				logger.Error("cannot parse server from json:", e3)
				httputil.WriteErrResponse(writer, http.StatusInternalServerError, "Internal Server Error.")
				return
			}
			if ss != nil && len(ss) > 0 {
				for i := range ss {
					s := ss[i]
					exist := false
					common.WalkList(&servers, func(item interface{}) bool {
						c := item.(*app.StorageDO)
						if c.Host == s.Host && c.Port == s.Port {
							exist = true
							return true
						}
						return false
					})
					if !exist {
						servers.PushBack(&s)
					}
				}
			}
		}
	}

	for k, v := range request.MultipartForm.File {
		if k != "template" && k != "servers" {
			continue
		}
		fh := v[0]
		logger.Debug("upload file name:", fh.Filename)
		if fh.Size > MaxUploadFileSize {
			httputil.WriteErrResponse(writer, http.StatusRequestEntityTooLarge, "Request entity too large.")
			return
		}
		f, e2 := fh.Open()
		if e2 != nil {
			logger.Error("error open file:", e2)
			httputil.WriteErrResponse(writer, http.StatusInternalServerError, "Cannot read template file.")
			return
		}
		buffer, e3 := bridgev2.MakeBytes(fh.Size, false, 0, true)
		if e3 != nil {
			logger.Error("cannot create buffer bytes:", e3)
			httputil.WriteErrResponse(writer, http.StatusInternalServerError, "Cannot read template file.")
			return
		}
		f.Read(buffer)
		content := string(buffer)
		if servers.Len() == 0 {
			content = strings.Replace(content, storage_servers_placeholder, "\n", -1)
			content = strings.Replace(content, uploadable_servers_placeholder, "\n", -1)
			content = strings.Replace(content, merged_servers_placeholder, "\n", -1)
		} else {
			// allServers is used by download upstreams
			var allServers list.List
			// uploadableServers is used by upload upstreams
			var uploadableServers list.List
			// mergedServers is used by stream servers
			var mergedServers list.List

			common.WalkList(&servers, func(item interface{}) bool {
				server := item.(*app.StorageDO)
				allServers.PushBack(server)
				exist := false
				if server.HttpEnable && !server.ReadOnly {
					common.WalkList(&uploadableServers, func(item interface{}) bool {
						it := item.(*app.StorageDO)
						if it.Host == server.Host && it.Port == server.Port {
							exist = true
							return true
						}
						return false
					})
					if !exist {
						uploadableServers.PushBack(server)
					}
				}
				exist = false
				common.WalkList(&mergedServers, func(item interface{}) bool {
					it := item.(*app.StorageDO)
					if it.Port == server.Port {
						exist = true
						return true
					}
					return false
				})
				if !exist {
					mergedServers.PushBack(server)
				}
				return false
			})

			var bf bytes.Buffer

			// replace placeholders
			if allServers.Len() > 0 {
				bf.WriteString("\n    upstream storage_servers {\n")
				common.WalkList(&allServers, func(item interface{}) bool {
					it := item.(*app.StorageDO)
					bf.WriteString("        server "+ it.Host + ":" + strconv.Itoa(it.HttpPort) +" weight=1;\n")
					return false
				})
				bf.WriteString("    }\n")
			}
			content = strings.Replace(content, storage_servers_placeholder, bf.String(), -1)

			bf.Reset()

			if uploadableServers.Len() > 0 {
				bf.WriteString("\n    upstream uploadable_servers {\n")
				common.WalkList(&uploadableServers, func(item interface{}) bool {
					it := item.(*app.StorageDO)
					bf.WriteString("        server "+ it.Host + ":" + strconv.Itoa(it.HttpPort) +" weight=1;\n")
					return false
				})
				bf.WriteString("    }\n")
			}
			content = strings.Replace(content, uploadable_servers_placeholder, bf.String(), -1)

			bf.Reset()

			common.WalkList(&mergedServers, func(item interface{}) bool {
				it := item.(*app.StorageDO)
				bf.WriteString("\n    server {\n")
				bf.WriteString("        listen " + strconv.Itoa(it.Port) +";\n")
				bf.WriteString("        proxy_pass "+ it.Host + ":" + strconv.Itoa(it.Port) +";\n")
				bf.WriteString("    }\n")
				return false
			})
			content = strings.Replace(content, merged_servers_placeholder, bf.String(), -1)
		}

		// handle http options method
		headers := writer.Header()
		headers.Set("Content-Transfer-Encoding", "binary")
		headers.Set("Content-Disposition", "attachment;filename=\"nginx.conf\"")
		writer.WriteHeader(200)
		writer.Write([]byte(content))
		break
	}

}

// get all storage servers
func GetAllStorageServers(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	if !httputil.MethodAllow(writer, request, http.MethodGet, http.MethodOptions) {
		return
	}
	if !httputil.AccessCheck(writer, request) {
		return
	}
	servers := libcommon.GetAllStorageServers()
	ret, _ := json.Marshal(servers)

	headers := writer.Header()
	headers.Set("Content-Transfer-Encoding", "binary")
	headers.Set("Content-Disposition", "attachment;filename=\"servers.json\"")
	writer.WriteHeader(200)
	writer.Write([]byte(ret))
}



