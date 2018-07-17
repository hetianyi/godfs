package lib_storage

import (
    "net/http"
    "strings"
    "regexp"
    "util/logger"
    "util/file"
    "strconv"
    "io"
)

const (
    pathRegex = "^/download/([0-9a-zA-Z_]+)/([0-9a-zA-Z_]+)/([0-9a-fA-F]{32})$"
)
var (
    compiledRegexp *regexp.Regexp
)

func init() {
    compiledRegexp = regexp.MustCompile(pathRegex)
}

// storage server provide http download service
func DownloadHandler(writer http.ResponseWriter, request *http.Request) {
    qIndex := strings.Index(request.RequestURI, "?")
    var servletPath = request.RequestURI
    if qIndex != -1 {
        servletPath = request.RequestURI[0:qIndex]
    }

    var md5 string
    if mat, _ := regexp.Match(pathRegex, []byte(servletPath)); !mat {
        writer.WriteHeader(404)
        writer.Write([]byte("Not found."))
        return
    }
    md5 = compiledRegexp.ReplaceAllString(servletPath, "${3}")
    finalPath := GetFilePathByMd5(md5)
    logger.Info("file path is :", finalPath)
    if file.Exists(finalPath) {
        downFile, e := file.GetFile(finalPath)
        if e != nil {
            writer.WriteHeader(500)
            return
        } else {
            fInfo, _ := downFile.Stat()
            headers := writer.Header()
            headers.Set("Pragma", "public")
            headers.Set("Expires", "0")
            headers.Set("Cache-Control", "must-revalidate, post-check=0, pre-check=0")
            headers.Set("Content-Type", "application/octet-stream")
            headers.Set("Content-Length", strconv.FormatInt(fInfo.Size(), 10))
            headers.Set("Content-Disposition", "attachment;filename=测试.rar")
            headers.Set("Content-Transfer-Encoding", "binary")

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


