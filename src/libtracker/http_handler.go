package libtracker

import (
	"app"
	"bytes"
	"container/list"
	json "github.com/json-iterator/go"
	"libcommon"
	"libcommon/bridgev2"
	"net/http"
	"strconv"
	"strings"
	"util/common"
	httputil "util/http"
	"util/logger"
)

const MaxUploadFileSize = 1024 * 10

var (
	download_servers_placeholder     = "<%download_upstream_servers%>"
	uploadable_servers_placeholder   = "<%uploadable_upstream_servers%>"
	download_locations_placeholder   = "<%download_locations%>"
	upload_locations_placeholder     = "<%upload_locations%>"
	all_upstream_servers_placeholder = "<%all_upstream_servers%>"
	all_servers_placeholder          = "<%all_servers%>"
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

	// parse storage server info which collected by nginx client using curl
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
						if c.Uuid == s.Uuid {
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

	// parse nginx template file
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
			content = strings.Replace(content, download_servers_placeholder, "\n", -1)
			content = strings.Replace(content, uploadable_servers_placeholder, "\n", -1)
			content = strings.Replace(content, download_locations_placeholder, "\n", -1)
			content = strings.Replace(content, all_upstream_servers_placeholder, "\n", -1)
			content = strings.Replace(content, all_servers_placeholder, "\n", -1)
		} else {
			// allServers is used by download upstreams
			var downloadServers list.List
			// uploadableServers is used by upload upstreams
			var uploadableServers list.List
			// mergedServers is used by stream servers
			var allServers list.List

			common.WalkList(&servers, func(item interface{}) bool {
				exist := false
				// collect download servers
				server := item.(*app.StorageDO)
				if server.HttpEnable {
					common.WalkList(&downloadServers, func(item interface{}) bool {
						it := item.(*app.StorageDO)
						if it.Host == server.Host && it.HttpPort == server.HttpPort {
							exist = true
							return true
						}
						return false
					})
					if !exist {
						downloadServers.PushBack(server)
					}
				}

				// collect upload servers
				exist = false
				if server.HttpEnable && !server.ReadOnly {
					common.WalkList(&uploadableServers, func(item interface{}) bool {
						it := item.(*app.StorageDO)
						if it.Host == server.Host && it.HttpPort == server.HttpPort {
							exist = true
							return true
						}
						return false
					})
					if !exist {
						uploadableServers.PushBack(server)
					}
				}

				// collect all servers
				exist = false
				common.WalkList(&allServers, func(item interface{}) bool {
					it := item.(*app.StorageDO)
					if it.Host == server.Host && it.Port == server.Port {
						exist = true
						return true
					}
					return false
				})
				if !exist {
					allServers.PushBack(server)
				}

				return false
			})

			var bf bytes.Buffer

			// replace placeholders
			if downloadServers.Len() > 0 {
				// group servers by group field
				gMap := GroupByGroup(&downloadServers)
				if gMap != nil {
					for k, v := range *gMap {
						if v != nil && v.Len() > 0 {
							bf.WriteString("\n    upstream download_servers_" + k + " {\n")
							common.WalkList(v, func(item interface{}) bool {
								it := item.(*app.StorageDO)
								bf.WriteString("        server " + it.Host + ":" + strconv.Itoa(it.HttpPort) + " weight=1;\n")
								return false
							})
							bf.WriteString("    }\n")
						}
					}
				}
			}
			content = strings.Replace(content, download_servers_placeholder, bf.String(), -1)

			bf.Reset()

			// replace placeholders
			if downloadServers.Len() > 0 {
				// group servers by group field
				gMap := GroupByGroup(&downloadServers)
				if gMap != nil {
					for k, v := range *gMap {
						if v != nil && v.Len() > 0 {
							bf.WriteString("\n        location /download/" + k + " {\n")
							bf.WriteString("            proxy_next_upstream http_404 http_500;\n")
							bf.WriteString("            proxy_pass http://download_servers_" + k + ";\n")
							bf.WriteString("        }\n")
						}
					}
				}
			}
			content = strings.Replace(content, download_locations_placeholder, bf.String(), -1)

			bf.Reset()

			if uploadableServers.Len() > 0 {
				// all upload servers
				bf.WriteString("\n    upstream upload_servers {\n")
				common.WalkList(&uploadableServers, func(item interface{}) bool {
					it := item.(*app.StorageDO)
					bf.WriteString("        server " + it.Host + ":" + strconv.Itoa(it.HttpPort) + " weight=1;\n")
					return false
				})
				bf.WriteString("    }\n")

				// grouped upload servers
				gMap := GroupByGroup(&uploadableServers)
				if gMap != nil {
					for k, v := range *gMap {
						if v != nil && v.Len() > 0 {
							bf.WriteString("\n    upstream upload_servers_" + k + " {\n")
							common.WalkList(v, func(item interface{}) bool {
								it := item.(*app.StorageDO)
								bf.WriteString("        server " + it.Host + ":" + strconv.Itoa(it.HttpPort) + " weight=1;\n")
								return false
							})
							bf.WriteString("    }\n")
						}
					}
				}

			}
			content = strings.Replace(content, uploadable_servers_placeholder, bf.String(), -1)

			bf.Reset()

			if uploadableServers.Len() > 0 {
				// all upload servers
				bf.WriteString("\n        location /upload {\n")
				bf.WriteString("            proxy_next_upstream non_idempotent http_404;\n")
				bf.WriteString("            proxy_pass http://upload_servers;\n")
				bf.WriteString("        }\n")

				// grouped upload servers
				gMap := GroupByGroup(&uploadableServers)
				if gMap != nil {
					for k, v := range *gMap {
						if v != nil && v.Len() > 0 {
							bf.WriteString("\n        location /upload/"+ k +" {\n")
							bf.WriteString("            proxy_next_upstream non_idempotent http_404;\n")
							bf.WriteString("            proxy_pass http://upload_servers_"+ k +" ;\n")
							bf.WriteString("        }\n")
						}
					}
				}

			}
			content = strings.Replace(content, upload_locations_placeholder, bf.String(), -1)

			bf.Reset()

			// replace placeholders
			if allServers.Len() > 0 {
				// group servers by group field
				gMap := GroupByPort(&allServers)
				if gMap != nil {
					for k, v := range *gMap {
						if v != nil && v.Len() > 0 {
							bf.WriteString("\n    upstream servers_" + strconv.Itoa(k) + " {\n")
							common.WalkList(v, func(item interface{}) bool {
								it := item.(*app.StorageDO)
								bf.WriteString("        server " + it.Host + ":" + strconv.Itoa(it.Port) + ";\n")
								return false
							})
							bf.WriteString("    }\n")
						}
					}
				}
			}
			content = strings.Replace(content, all_upstream_servers_placeholder, bf.String(), -1)

			bf.Reset()

			// replace placeholders
			if allServers.Len() > 0 {
				// group servers by group field
				gMap := GroupByPort(&allServers)
				if gMap != nil {
					for k, v := range *gMap {
						if v != nil && v.Len() > 0 {
							bf.WriteString("\n    server {\n")
							bf.WriteString("        listen " + strconv.Itoa(k) + ";\n")
							bf.WriteString("        proxy_pass servers_" + strconv.Itoa(k) + ";\n")
							bf.WriteString("    }\n")
						}
					}
				}
			}
			content = strings.Replace(content, all_servers_placeholder, bf.String(), -1)

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

func GroupByGroup(servers *list.List) *map[string]*list.List {
	if servers == nil {
		return nil
	}
	tmpMap := make(map[string]*list.List)
	common.WalkList(servers, func(item interface{}) bool {
		server := item.(*app.StorageDO)
		group := server.Group
		groupList := tmpMap[group]
		if groupList == nil {
			groupList = list.New()
			tmpMap[group] = groupList
		}
		groupList.PushBack(server)
		return false
	})
	return &tmpMap
}

func GroupByPort(servers *list.List) *map[int]*list.List {
	if servers == nil {
		return nil
	}
	tmpMap := make(map[int]*list.List)
	common.WalkList(servers, func(item interface{}) bool {
		server := item.(*app.StorageDO)
		port := server.Port
		groupList := tmpMap[port]
		if groupList == nil {
			groupList = list.New()
			tmpMap[port] = groupList
		}
		groupList.PushBack(server)
		return false
	})
	return &tmpMap
}
