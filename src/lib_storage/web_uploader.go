package lib_storage

import (
    "os"
    "app"
    "util/logger"
    "lib_common"
    "errors"
    "util/file"
    "net/http"
    "strings"
    "regexp"
    "bytes"
    "util/common"
    "lib_common/bridge"
)

type IHandler interface {
    initConfig()
    onTextField(paraName string, paraValue string) error
    onFileField(filename string) error
    onFileEnd() error
    writeFile(b []byte, start int, length int) error
    onRequestInputStreamInterrupt() error
    onError(e interface{}) error
    onFileSizeExceed() error
    onUploadFinish() error
    beginUpload() error
    writeBack(response string)
}


//文件上传处理类
//使用者只需继承此类，然后在公用方法里面参与上传事件的处理即可<br>
//最简单的过程即为覆写onFileField()函数
type FileUploadHandler struct {
    buffersize int
    maxUploadSize int64
    cpulimit int
    bufsleep int
    filename string
    boundary string
    file *os.File
    paraMap map[string]string
    totalbytes int64
    buffertimes int64
    writer http.ResponseWriter
    request *http.Request
}


func (handler *FileUploadHandler) initConfig() {
    handler.buffersize = app.BUFF_SIZE
    handler.maxUploadSize = 1024*1024*1024*2
    handler.cpulimit = 1024
    handler.bufsleep = 1
    handler.filename = ""
    handler.boundary = ""
    handler.file = nil
    handler.paraMap = make(map[string] string)
    handler.totalbytes = 0
    handler.buffertimes = 0
}


func (handler *FileUploadHandler) onTextField(paraName string, paraValue string) {
    logger.Info(paraName, "=", paraValue)
}

func (handler *FileUploadHandler) onFileField(filename string) error {
    logger.Info("upload file name:", filename)
    if handler.file != nil {
        handler.file.Close()
    }
    out, oe := lib_common.CreateTmpFile()
    if oe != nil {
        return oe
    }
    handler.file = out
    return nil
}

func (handler *FileUploadHandler) onFileEnd() error {
    logger.Info("file upload finish")
    if handler.file != nil {
        handler.file.Close()
    }
    return nil
}

func (handler *FileUploadHandler) writeFile(b []byte, start int, length int) error {
    if handler.file != nil {
        len, e := handler.file.Write(b[start:start + length])
        if e != nil {
            return e
        }
        if len != length {
            return errors.New("error write file[1]")
        }
    }
    return errors.New("error write file[2]")
}

func (handler *FileUploadHandler) onRequestInputStreamInterrupt() error {
    logger.Info("file upload interrupt")
    if handler.file != nil {
        handler.file.Close()
        if !file.Delete(handler.file.Name()) {
            logger.Error("error delete tmp file[1")
        }
    }
    return nil
}

func (handler *FileUploadHandler) onError(e interface{}) error {
    if handler.file != nil {
        handler.file.Close()
        if !file.Delete(handler.file.Name()) {
            logger.Error("error delete tmp file[1")
        }
    }
    return nil
}

func (handler *FileUploadHandler) onFileSizeExceed() {
    logger.Error("file length exceed")
}

func (handler *FileUploadHandler) onUploadFinish() error {
    logger.Debug("upload finish")
    return nil
}

func (handler *FileUploadHandler) writeBack(response string) {
    handler.writer.Write([]byte(response))
}

func (handler *FileUploadHandler) beginUpload() {
    handler.initConfig()
    total := handler.request.ContentLength
    if total > handler.maxUploadSize || total <= 0 {//文件超过限制
        handler.onFileSizeExceed()
        handler.writeBack("{file length exceed}")
        return
    }

    headerContentType := handler.request.Header["Content-Type"]
    contentType := ""
    agent := ""
    if headerContentType != nil || len(headerContentType) == 0 {
        contentType = headerContentType[0]
    }
    for k, v := range handler.request.Header {
        if strings.Index(k, "agent") != -1 {
            if v != nil || len(v) == 0 {
                agent = v[0]
            }
        }
    }

    if mat, _ := regexp.Match("^multipart/form-data; boundary=.*$", []byte(contentType)); mat {
        handler.boundary = regexp.MustCompile("^.*boundary=(.*)$").ReplaceAllString(contentType, "${1}")
        paraSeparator := "--" + handler.boundary
        endSeparator := "--" + handler.boundary + "--"

        //修正，缓冲区不能太小
        if handler.buffersize < 1024 {
            handler.buffersize = 1024
        }

        paraMap := make(map[string]string)
        readname := false
        readStrValue := false
        readByteValue := false
        skipToFile := false
        skipToValue := false

        endRead := false

        paraName := ""
        paraValue := ""
        begin := false

        ips := handler.request.Body
        tmp := make([]byte, 1)
        move := make([]byte, 2)
        skip := make([]byte, 2)
        var b byte
        var bf = bytes.Buffer{}
        var bf_pos int

        common.Try(func() {
            sb := ""
            for {
                len, re11 := bridge.ReadBytes(tmp, 1, ips, nil)
                if re11 != nil || len != 1 {
                    break
                }
                if bf_pos == handler.buffersize {
                    if readStrValue {
                        if "" == paraValue {
                            paraValue = string(bf.Bytes()[0:bf_pos - 1])
                        } else {
                            paraValue += string(bf.Bytes()[0:bf_pos - 1])
                        }
                        logger.Debug(paraValue)
                        bf.Reset()
                        bf_pos = 0
                    } else {
                        if begin {
                            handler.writeFile([]byte{13, 10}, 0, 2)
                            begin = false
                        }
                        handler.writeFile(bf.Bytes(), 0, bf.Len())
                        bf.Reset()
                        bf_pos = 0
                    }
                }

                b = tmp[0]
                bf.Write(tmp)
                bf_pos++
                if b == byte(13) {
                    move[0] = b
                    len, re11 := bridge.ReadBytes(tmp, 1, ips, nil)
                    if re11 != nil || len != 1 {
                        break
                    }
                    if re11 == nil && len == 1 {
                        b = tmp[0]
                        move[1] = b
                        if b == byte(10) {
                            sb = string(bf.Bytes()[0:bf_pos - 1])
                            if sb == endSeparator {
                                if readStrValue {
                                    paraMap[paraName] = paraValue
                                    handler.onTextField(paraName, paraValue)
                                    paraValue = ""
                                }
                                if readByteValue {
                                    handler.onFileEnd()
                                }
                                endRead = true
                                break
                            }
                            if sb == paraSeparator {
                                if readStrValue {
                                    paraMap[paraName] = paraValue
                                    handler.onTextField(paraName, paraValue)
                                    paraValue = ""
                                }
                                if readByteValue {
                                    handler.onFileEnd()
                                }
                                readname = true
                                readStrValue = false
                                readByteValue = false
                                skipToFile = false
                                skipToValue = false
                                bf.Reset()
                                bf_pos = 0
                                continue
                            }

                            if readStrValue {
                                if "" == paraValue {
                                    paraValue = sb
                                } else {
                                    paraValue += sb
                                }
                                bf.Reset()
                                bf_pos = 0
                                continue
                            }
                            if readByteValue {
                                if begin {
                                    handler.writeFile([]byte{13, 10}, 0, 2)
                                }
                                handler.writeFile(bf.Bytes(), 0, bf_pos - 1)
                                begin = true
                                bf.Reset()
                                bf_pos = 0
                                continue
                            }
                            if skipToValue {
                                readname = false
                                readStrValue = true
                                readByteValue = false
                                skipToFile = false
                                skipToValue = false
                                bf.Reset()
                                bf_pos = 0
                                continue
                            }
                            if skipToFile {
                                readname = false
                                readStrValue = false
                                readByteValue = true
                                skipToFile = false
                                skipToValue = false
                                bridge.ReadBytes(skip, 2, ips, nil)
                                bf.Reset()
                                bf_pos = 0
                                begin = false
                                continue
                            }
                            if readname {
                                if mat, _ := regexp.Match("^.*filename=\".*\"$", []byte(sb)); mat {
                                    paraName = regexp.MustCompile("^.*filename=\"(.*)\"$").ReplaceAllString(sb, "${1}")
                                    if strings.Contains(paraName, "\\") {
                                        paraName = paraName[strings.LastIndex(paraName, "\\") + 1:]
                                    }
                                    handler.filename = paraName
                                    handler.onFileField(handler.filename)
                                    readname = false
                                    readStrValue = false
                                    readByteValue = false
                                    skipToFile = true
                                    skipToValue = false
                                } else {
                                    paraName = regexp.MustCompile("^.*name=\"(.*)\"$").ReplaceAllString(sb, "${1}")
                                    readname = false
                                    readStrValue = false
                                    readByteValue = false
                                    skipToFile = false
                                    skipToValue = true
                                }
                                bf.Reset()
                                bf_pos = 0
                                continue
                            }
                        } else {
                            if bf_pos == handler.buffersize {
                                if readStrValue {
                                    if "" == paraValue {
                                        paraValue = string(bf.Bytes()[0:bf_pos - 1])
                                    } else {
                                        paraValue += string(bf.Bytes()[0:bf_pos - 1])
                                    }
                                    bf.Reset()
                                    bf_pos = 0
                                } else {
                                    if begin {
                                        handler.writeFile([]byte{13, 10}, 0, 2)
                                        begin = false
                                    }
                                    handler.writeFile(bf.Bytes(), 0, bf.Len())
                                    bf.Reset()
                                    bf_pos = 0
                                }
                            }
                            b = tmp[0]
                            bf.Write(tmp)
                            bf_pos++
                        }
                    } else {
                        endRead = true
                        break
                    }
                }
            }
            if strings.Contains(agent, "Shockwave") || strings.Contains(agent, "Flash") {
                handler.onTextField(paraName, paraValue)
            }
            endRead = true
        }, func(i interface{}) {
            logger.Error(i)
            handler.onError(i)
        })

        if !endRead {
            handler.onRequestInputStreamInterrupt()
        } else {
            handler.onUploadFinish()
        }
    } else {
        logger.Error("error contentType format")
    }
}


func WebUploadHandler(writer http.ResponseWriter, request *http.Request) {
    handler := &FileUploadHandler{
        writer: writer,
        request: request,
    }
    handler.beginUpload()
}