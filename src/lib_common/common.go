package lib_common

import (
    "util/logger"
    "regexp"
    "strconv"
    "util/timeutil"
    "util/file"
    "app"
    "os"
    "strings"
)

func CreateTmpFile() (*os.File, error) {
    // begin upload file
    tmpFileName := timeutil.GetUUID()
    // using tmp ext and rename after upload success
    tmpPath := file.FixPath(app.BASE_PATH + "/data/tmp/" + tmpFileName)
    fi, e8 := file.CreateFile(tmpPath)
    if e8 != nil {
        return nil, e8
    }
    return fi, nil
}

func CloseAndDeleteTmpFile(fi *os.File) {
    fi.Close()
    file.Delete(fi.Name())
}

func MoveTmpFileTo(md5 string, fi *os.File) error {
    dig1 := strings.ToUpper(md5[0:2])
    dig2 := strings.ToUpper(md5[2:4])
    finalPath := app.BASE_PATH + "/data/" + dig1 + "/" + dig2
    if !file.Exists(finalPath) {
        e := file.CreateAllDir(finalPath)
        if e != nil {
            return e
        }
    }
    if !file.Exists(finalPath + "/" + md5) {
        eee := file.MoveFile(fi.Name(), finalPath + "/" + md5)
        if eee != nil {
            return eee
        }
    } else {
        s := file.Delete(fi.Name())
        if !s {
            logger.Error("error clean tmp file:", fi.Name())
        }
    }
    return nil
}

// 从string解析port端口，返回int类型的port端口
// 如果返回0表示port不合法
func ParsePort(port string) int {
    if len(port) < 1 {
        logger.Error("parameter 'port' not set yet, server will not exit now!")
        return 0
    }
    if b, _ := regexp.Match("^[1-9][0-9]{1,6}$", []byte(port)); b {
        p, e := strconv.Atoi(port)
        if e != nil || p > 65535 {
            logger.Error("parameter 'port' must be a valid port number!")
            return 0
        }
        return p
    }
    return 0
}
