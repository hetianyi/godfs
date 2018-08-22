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
    "container/list"
    "errors"
    "io"
    "fmt"
    "bytes"
    "math"
    "time"
    "hash"
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



// parse trackers into a list
func ParseTrackers(tracker string) *list.List {
    sp := strings.Split(tracker, ",")
    ls := list.New()
    for i := range sp {
        trimS := strings.TrimSpace(sp[i])
        if len(trimS) > 0 {
            ls.PushBack(trimS)
        }

    }
    return ls
}



func SeekWriteOut(in io.ReadSeeker, start int64, offset int64, buffer []byte, out io.Writer) error {
    // total read bytes
    var readBodySize int64 = 0
    // next time bytes to read
    var nextReadSize int
    _, e1 := in.Seek(start, 0)
    if e1 != nil {
        return e1
    }
    for {
        // left bytes is more than a buffer
        if (offset - readBodySize) / int64(len(buffer)) >= 1 {
            nextReadSize = len(buffer)
        } else {// left bytes less than a buffer
            nextReadSize = int(offset - readBodySize)
        }
        if nextReadSize == 0 {
            break
        }
        len, e2 := in.Read(buffer[0:nextReadSize])
        if e2 == nil {
            wl, e5 := out.Write(buffer[0:len])
            readBodySize += int64(len)
            logger.Trace("write:", readBodySize)
            if e5 != nil || wl != len {
                return errors.New("error handle download file")
            }
            app.UpdateIOOUT(int64(wl))
        } else {
            if e2 == io.EOF {
                return nil
            }
            return e2
        }
    }
    return nil
}

func WriteOut(in io.Reader, offset int64, buffer []byte, out io.Writer, md hash.Hash, callback func(len int)) error {
    // total read bytes
    var readBodySize int64 = 0
    // next time bytes to read
    var nextReadSize int
    for {
        // left bytes is more than a buffer
        if (offset - readBodySize) / int64(len(buffer)) >= 1 {
            nextReadSize = len(buffer)
        } else {// left bytes less than a buffer
            nextReadSize = int(offset - readBodySize)
        }
        if nextReadSize == 0 {
            break
        }
        len, e2 := in.Read(buffer[0:nextReadSize])
        if e2 == nil {
            wl, e5 := out.Write(buffer[0:len])
            callback(len)
            if md != nil {
                mdwl, e6 := md.Write(buffer[0:len])
                if e6 != nil || mdwl != len {
                    return errors.New("error write md")
                }
            }
            if e5 != nil || wl != len {
                return errors.New("error write out")
            }
            readBodySize += int64(len)
            logger.Trace("write:", readBodySize)
        } else {
            if e2 == io.EOF {
                return nil
            }
            return e2
        }
    }
    return nil
}



func ShowPercent(total *int64, finish *int64, stop *bool, startTime time.Time) {
    //int(math.Floor(12312300000*100*1.0/91231231234))
    timer := time.NewTicker(time.Millisecond * 1000)
    var buffer bytes.Buffer
    shine := true
    var lastRead int64 = 0
    var lastFinish int64 = 0
    for !*stop {
        buffer.Reset()
        <-timer.C
        tot := *total
        lastRead = *finish - lastFinish // total read bytes during last 100ms.
        lastFinish = *finish
        percent := int(math.Floor(float64(lastFinish*100*1.0/tot)))
        percent1 := int(math.Floor(float64(lastFinish*10*1.0/tot)))
        buffer.WriteString("[")

        for i := 0; i < percent1; i++ {
            buffer.WriteString("=")
        }
        if percent1 < 10 {
            if shine {
                buffer.WriteString(">")
            } else {
                buffer.WriteString(" ")
            }
        } else {
            buffer.WriteString("=")
        }
        for i := 0; i < 10-percent1-1; i++ {
            buffer.WriteString(" ")
        }

        shine = !shine
        now := time.Now()

        buffer.WriteString("]")
        buffer.WriteString(FixLength(percent, 3, " "))
        buffer.WriteString("%")
        buffer.WriteString(" [")
        buffer.WriteString(timeutil.GetHumanReadableDuration(startTime, now))
        buffer.WriteString("]")
        buffer.WriteString(" - ")
        buffer.WriteString(HumanReadable(lastRead, 1000))
        fmt.Print(buffer.String() + "\r")
    }

    buffer.Reset()
    timer.Stop()
}

//将数字补齐为固定宽度，不足宽度在前补齐
func FixLength(num int, width int, fixChar string) string {
    snum := strconv.Itoa(num)
    curLen := len(snum)
    if curLen < width {
        for i := curLen; i < width; i++ {
            snum = fixChar + snum
        }

    }
    return snum
}

// at 'during' ms read total 'len' bytes
func HumanReadable(len int64, during int64) string {
    if len < 1024 {
        return strconv.FormatInt(len * 1000 / during, 10) + "B/s       "
    } else if len < 1048576 {
        return strconv.FormatInt(len*1.0 / 1024 * 1000 / during, 10) + "KB/s       "
    } else if len < 1073741824 {
        return fmt.Sprintf("%.2f", float64(len) / 1048576 * 1000 / float64(during)) + "MB/s       "
    } else {
        return fmt.Sprintf("%.2f", float64(len)/1073741824 * 1000 / float64(during)) + "GB/s       "
    }
}