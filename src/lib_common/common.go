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
    "lib_common/bridge"
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


// return file path using md5
func GetFilePathByMd5(md5 string) string {
    dig1 := strings.ToUpper(md5[0:2])
    dig2 := strings.ToUpper(md5[2:4])
    return app.BASE_PATH + "/data/" + dig1 + "/" + dig2 + "/" + md5
}

// get read position from parts of files
// returns part index, current part start read position and total read bytes length
func GetReadPositions(fullFile *bridge.File, start int64, offset int64) (*bridge.ReadPos, *bridge.ReadPos, int64) {
    var fileLen int64
    for i := range fullFile.Parts {
        fileLen += fullFile.Parts[i].FileSize
    }
    // exceed
    if start >= fileLen {
        return nil, nil, 0
    }
    var addLen int64 = 0
    var end = start + offset
    if offset < 0 || end > fileLen {
        end = fileLen
    }
    var startPos *bridge.ReadPos
    var endPos = &bridge.ReadPos{PartIndex: len(fullFile.Parts) - 1, PartStart: fullFile.Parts[len(fullFile.Parts) - 1].FileSize}
    for i := range fullFile.Parts {
        fInfo, _ := os.Stat(GetFilePathByMd5(fullFile.Parts[i].Md5))
        if start >= addLen + fInfo.Size() {
            addLen += fInfo.Size()
            continue
        } else {
            startPos = &bridge.ReadPos{PartIndex: i, PartStart: start - addLen}
            addLen += fInfo.Size()
            break
        }
    }
    addLen = 0
    for i := range fullFile.Parts {
        fInfo, _ := os.Stat(GetFilePathByMd5(fullFile.Parts[i].Md5))
        if end > addLen + fInfo.Size() {
            addLen += fInfo.Size()
            continue
        } else {
            endPos = &bridge.ReadPos{PartIndex: i, PartStart: end - addLen}
            addLen += fInfo.Size()
            break
        }
    }
    return startPos, endPos, end - start
}