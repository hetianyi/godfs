package libstorage

import (
	"app"
	"libcommon"
	"libservice"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"util/file"
	"util/logger"
)

const (
	pathRegexRestful = "^/download/([0-9a-zA-Z_]{1,10})/([0-9a-zA-Z_]{1,10})/([MS])/([0-9a-fA-F]{32})(/([^/]*))?$"
	rangeHeader      = "^bytes=([0-9]+)-([0-9]+)?$"
)

var (
	compiledRegexpRestful     *regexp.Regexp
	compiledRegexpRangeHeader *regexp.Regexp
	gmtLocation, _            = time.LoadLocation("GMT")
)

func init() {
	compiledRegexpRestful = regexp.MustCompile(pathRegexRestful)
	compiledRegexpRangeHeader = regexp.MustCompile(rangeHeader)
}

// storage server provide http download service
func DownloadHandler(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	method := request.Method
	if method != http.MethodGet && method != http.MethodOptions {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		writer.Write([]byte(strconv.Itoa(http.StatusMethodNotAllowed) + " Method '" + method + "' not allowed."))
		return
	}

	// handle http options method
	headers := writer.Header()
	origin := ""
	origins := request.Header["Origin"]
	if origins != nil && len(origins) > 0 {
		origin = origins[0]
	}
	if app.CheckOriginAllow(origin) {
		headers.Set("Access-Control-Allow-Origin", origin)
	}
	headers.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	headers.Set("Access-Control-Allow-Credentials", "true")
	headers.Set("Access-Control-Allow-Headers", "*")
	if method == http.MethodOptions {
		writer.WriteHeader(205)
		return
	}

	if app.HTTP_AUTH != "" {
		user, pass, _ := request.BasicAuth()
		if app.HTTP_AUTH != user+":"+pass {
			writer.WriteHeader(403)
			writer.Write([]byte("403 Forbidden."))
			return
		}
	}

	qIndex := strings.Index(request.RequestURI, "?")
	var servletPath = request.RequestURI
	if qIndex != -1 {
		servletPath = request.RequestURI[0:qIndex]
	}

	mat, _ := regexp.Match(pathRegexRestful, []byte(servletPath))
	if !mat {
		writer.WriteHeader(404)
		writer.Write([]byte("404 Not Found."))
		return
	}

	var md5 string
	md5 = compiledRegexpRestful.ReplaceAllString(servletPath, "${4}")

	fn := compiledRegexpRestful.ReplaceAllString(servletPath, "${6}")
	if fn == "" {
		queryValues := request.URL.Query()
		fns := queryValues["fn"]
		if fns != nil && len(fns) != 0 {
			fn = fns[0]
		} else {
			fn = md5
		}
	}
	logger.Debug("custom download file name is:", fn)
	ext := file.GetFileExt(fn)
	headers.Set("Content-Type", *app.GetContentTypeHeader(ext))

	// 304 Not Modified
	eTag := request.Header["If-None-Match"]
	if app.MIME_TYPES_ENABLE && eTag != nil && len(eTag) > 0 && eTag[0] == "\""+md5+"\"" {
		setMimeHeaders(md5, &headers)
		writer.WriteHeader(304)
		return
	}

	fullFile, e11 := libservice.GetFullFileByMd5(md5, 1)
	if e11 != nil {
		writer.WriteHeader(500)
		writer.Write([]byte("500 Internal Server Error"))
		return
	}
	if fullFile == nil {
		writer.WriteHeader(404)
		writer.Write([]byte("404 Not Found."))
		return
	}
	if len(fullFile.Parts) == 0 {
		writer.WriteHeader(500)
		writer.Write([]byte("500 Internal Server Error"))
		return
	}

	var fileSize int64 = 0
	for i := range fullFile.Parts {
		fileSize += fullFile.Parts[i].FileSize
	}

	// parse header: range
	rangeH := request.Header["Range"]
	var rangeS string
	if rangeH != nil && len(rangeH) > 0 {
		rangeS = rangeH[0]
	}
	logger.Trace(rangeS)
	start, end := parseHeaderRange(rangeS)
	if start <= 0 {
		start = 0
	}

	if end <= 0 || end > (fileSize-1) || end == start {
		end = fileSize
	}
	startPos, endPos, totalLen := libcommon.GetReadPositions(fullFile, start, end-start)

	headers.Set("Accept-Ranges", "bytes")
	headers.Set("Content-Length", strconv.FormatInt(totalLen, 10))
	headers.Set("Content-Range", "bytes "+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end-1, 10)+"/"+strconv.FormatInt(fileSize, 10))

	logger.Trace("range:", start, "-", end)
	if app.MIME_TYPES_ENABLE && app.SupportWebContent(ext) {
		gmtLocation, _ := time.LoadLocation("GMT")
		fInfo, _ := os.Stat(libcommon.GetFilePathByMd5(fullFile.Parts[0].Md5))
		headers.Set("Last-Modified", fInfo.ModTime().In(gmtLocation).Format(time.RFC1123))
		headers.Set("Expires", time.Now().Add(time.Hour*2400).In(gmtLocation).Format(time.RFC1123))
		setMimeHeaders(md5, &headers)
	} else {
		headers.Set("Expires", "0")
		headers.Set("Pragma", "public")
		//headers.Set("Accept-Ranges", "bytes")
		headers.Set("Content-Transfer-Encoding", "binary")
		headers.Set("Cache-Control", "must-revalidate, post-check=0, pre-check=0")
		headers.Set("Content-Disposition", "attachment;filename=\""+fn+"\"")
	}
	// adapt different clients
	// such as chrome needs 200 xunlei needs 206
	if rangeS == "" {
		writer.WriteHeader(200)
	} else {
		writer.WriteHeader(206)
	}

	app.UpdateDownloads()
	WriteDownloadStream(fullFile, startPos, endPos, writer)
}

func setMimeHeaders(md5 string, headers *http.Header) {
	//headers.Set("Cache-Control", "public")
	headers.Set("Cache-Control", "max-age=604800")
	headers.Set("Access-Control-Allow-Origin", "*")
	headers.Set("date", time.Now().In(gmtLocation).Format(time.RFC1123))
	headers.Set("Etag", "\""+md5+"\"")
	//headers.Set("Connection", "keep-alive")
}

// if end is 0, then the end represents max
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
