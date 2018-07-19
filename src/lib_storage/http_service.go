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
    "lib_common"
)

const (
    pathRegexRestful = "^/download/([0-9a-zA-Z_]+)/([0-9a-zA-Z_]+)/([0-9a-fA-F]{32})(/([^/]*))?$"
    rangeHeader = "^bytes=([0-9]+)-([0-9]+)?$"
)
var (
    compiledRegexpRestful *regexp.Regexp
    compiledRegexpRangeHeader *regexp.Regexp
    gmtLocation, _ = time.LoadLocation("GMT")
)

func init() {
    compiledRegexpRestful = regexp.MustCompile(pathRegexRestful)
    compiledRegexpRangeHeader = regexp.MustCompile(rangeHeader)
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
    if app.MIME_TYPES_ENABLE && eTag != nil && len(eTag) > 0 && eTag[0] == "\"" + md5 + "\"" {
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


            // parse header: range
            rangeH := request.Header["Range"]
            var rangeS string
            if rangeH != nil && len(rangeH) > 0 {
                rangeS = rangeH[0]
            }
            logger.Info(rangeS)
            start, end := parseHeaderRange(rangeS)
            if start <= 0 || start > (fInfo.Size() - 1) {
                start = 0
            }
            if start > (fInfo.Size() - 1) {
                start = fInfo.Size() - 1
            }
            if end <= 0 || end > (fInfo.Size() - 1) || end == start {
                end = fInfo.Size() - 1
            }

            headers.Set("Content-Type", *app.GetContentTypeHeader(ext))
            headers.Set("Accept-Ranges", "bytes")
            headers.Set("Content-Length", strconv.FormatInt(end - start + 1, 10))
            headers.Set("Content-Range", "bytes " + strconv.FormatInt(start, 10) + "-" + strconv.FormatInt(end, 10) + "/" + strconv.FormatInt(fInfo.Size(), 10))

            logger.Info("range:", start , "-", end)
            //fn = url.QueryEscape(fn)
            if !app.MIME_TYPES_ENABLE {
                headers.Set("Expires", "0")
                headers.Set("Pragma", "public")
                //headers.Set("Accept-Ranges", "bytes")
                headers.Set("Content-Transfer-Encoding", "binary")
                headers.Set("Cache-Control", "must-revalidate, post-check=0, pre-check=0")
                headers.Set("Content-Disposition", "attachment;filename=\"" + fn + "\"")
            } else {
                gmtLocation, _ := time.LoadLocation("GMT")
                headers.Set("Last-Modified", fInfo.ModTime().In(gmtLocation).Format(time.RFC1123))
                headers.Set("Expires", time.Now().Add(time.Hour * 2400).In(gmtLocation).Format(time.RFC1123))
                setMimeHeaders(md5, &headers)
            }
            // adapt different clients
            // such as chrome needs 200 xunlei needs 206
            if rangeS == "" {
                writer.WriteHeader(200)
            } else {
                writer.WriteHeader(206)
            }

            bodyBuffSize := 1024*10
            var buff = make([]byte, bodyBuffSize)
            read := 0
            readLen := end - start + 1
            var nextReadSize int
            downFile.Seek(start, 0)
            for {
                //read finish
                if int64(read) == readLen {
                    break
                }
                // left bytes is more than a buffer
                if (readLen - int64(read)) / int64(bodyBuffSize) >= 1 {
                    nextReadSize = int(bodyBuffSize)
                } else {// left bytes less than a buffer
                    nextReadSize = int(readLen - int64(read))
                }
                len, e2 := lib_common.ReadBytes(buff, nextReadSize, downFile)
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
                    read += wl
                    logger.Debug("write data....:", nextReadSize)
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
    //headers.Set("Cache-Control", "public")
    headers.Set("Cache-Control", "max-age=604800")
    headers.Set("Access-Control-Allow-Origin", "*")
    headers.Set("date", time.Now().In(gmtLocation).Format(time.RFC1123))
    headers.Set("Etag", "\"" + md5 + "\"")
    //headers.Set("Connection", "keep-alive")
}

// if end is 0, then the end represents max
func parseHeaderRange(rang string) (int64, int64) {
    if rang == "" {
        return 0,0
    }
    if mat, _ := regexp.Match(rangeHeader, []byte(rang)); !mat {
        return 0,0
    }
    s := compiledRegexpRangeHeader.ReplaceAllString(rang, "${1}")
    e := compiledRegexpRangeHeader.ReplaceAllString(rang, "${2}")
    uintS, e1 := strconv.ParseInt(s, 10, 64)
    uintE, e2 := strconv.ParseInt(e, 10, 64)
    if e1 != nil {
        return 0,0
    }
    if e2  != nil {
        return uintS, 0
    }
    return uintS, uintE
}


