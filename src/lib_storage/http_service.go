package lib_storage

import (
    "net/http"
    "strings"
    "regexp"
    "util/logger"
    "util/file"
    "strconv"
    "io"
    "app"
    "time"
)

const (
    pathRegexRestful = "^/download/([0-9a-zA-Z_]+)/([0-9a-zA-Z_]+)/([0-9a-fA-F]{32})(/([^/]*))?$"
)
var (
    compiledRegexpRestful *regexp.Regexp
    gmtLocation, _ = time.LoadLocation("GMT")
)

func init() {
    compiledRegexpRestful = regexp.MustCompile(pathRegexRestful)
}

// storage server provide http download service
func DownloadHandler(writer http.ResponseWriter, request *http.Request) {

    qIndex := strings.Index(request.RequestURI, "?")
    var servletPath = request.RequestURI
    if qIndex != -1 {
        servletPath = request.RequestURI[0:qIndex]
    }


    mat, _ := regexp.Match(pathRegexRestful, []byte(servletPath))
    if !mat {
        writer.WriteHeader(404)
        writer.Write([]byte("Not found."))
        return
    }


    var md5 string
    md5 = compiledRegexpRestful.ReplaceAllString(servletPath, "${3}")
    headers := writer.Header()
    eTag := request.Header["If-None-Match"]
    // 304 Not Modified
    if eTag != nil && len(eTag) > 0 && eTag[0] == "\"" + md5 + "\"" {
        setMimeHeaders(md5, &headers)
        writer.WriteHeader(304)
        return
    }

    fn := compiledRegexpRestful.ReplaceAllString(servletPath, "${5}")
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

    finalPath := GetFilePathByMd5(md5)
    logger.Debug("file path is :", finalPath)
    if file.Exists(finalPath) {
        downFile, e := file.GetFile(finalPath)
        if e != nil {
            writer.WriteHeader(500)
            return
        } else {
            fInfo, _ := downFile.Stat()
            ext := file.GetFileExt(fn)
            headers.Set("Content-Type", *app.GetContentTypeHeader(ext))
            headers.Set("Content-Length", strconv.FormatInt(fInfo.Size(), 10))
            //fn = url.QueryEscape(fn)
            if !app.MIME_TYPES_ENABLE {
                headers.Set("Expires", "0")
                headers.Set("Pragma", "public")
                headers.Set("Content-Transfer-Encoding", "binary")
                headers.Set("Cache-Control", "must-revalidate, post-check=0, pre-check=0")
                headers.Set("Content-Disposition", "attachment;filename=" + fn)
            } else {
                gmtLocation, _ := time.LoadLocation("GMT")
                headers.Set("Last-Modified", fInfo.ModTime().In(gmtLocation).Format(time.RFC1123))
                headers.Set("Expires", time.Now().Add(time.Hour * 2400).In(gmtLocation).Format(time.RFC1123))
                setMimeHeaders(md5, &headers)
            }

            var buff = make([]byte, 1024*10)
            for {
                len, e2 := downFile.Read(buff)
                if e2 == nil || e2 == io.EOF {
                    wl, e5 := writer.Write(buff[0:len])
                    if e2 == io.EOF {
                        logger.Info("file download success")
                        downFile.Close()
                        break
                    }
                    if e5 != nil || wl != len {
                        logger.Error("error write download file:", e5)
                        downFile.Close()
                        break
                    }
                } else {
                    logger.Error("error read download file:", e2)
                    downFile.Close()
                    break
                }
            }
        }
    } else {
        writer.WriteHeader(404)
        writer.Write([]byte("Not found."))
        return
    }
}


func setMimeHeaders(md5 string, headers *http.Header) {
    headers.Set("Cache-Control", "max-age=604800")
    headers.Set("Access-Control-Allow-Origin", "*")
    headers.Set("date", time.Now().In(gmtLocation).Format(time.RFC1123))
    headers.Set("Etag", "\"" + md5 + "\"")
    headers.Set("Connection", "keep-alive")
    headers.Set("Cache-Control", "public")
}


