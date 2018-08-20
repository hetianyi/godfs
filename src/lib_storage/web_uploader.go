package lib_storage

import (
    "net/http"
    "util/file"
    "bytes"
    "regexp"
    "io"
    "errors"
    "util/timeutil"
    "util/logger"
    "container/list"
)

const ContentDispositionPattern = "^Content-Disposition: form-data; name=\"([^\"]+)\"$"
const ContentTypePattern = "^multipart/form-data; boundary=(.*)$"


type FileFormReader struct {
    request *http.Request
    buffer *bytes.Buffer
}

func (reader *FileFormReader) Unread(read []byte) {
    reader.buffer.Write(read)
}

func (reader *FileFormReader) Read(buff []byte) (int, error) {
    if reader.buffer.Len() > 0 {
       if len(buff) <= reader.buffer.Len() {
           return reader.buffer.Read(buff)
       } else {
           reStartPos, e := reader.buffer.Read(buff)
           // read directly from reader
           len, e := reader.request.Body.Read(buff[reStartPos:])
           if e != nil {
               if e != io.EOF {
                   return 0, e
               }
           }
           return reStartPos + len, nil
       }
    }
    // read directly from reader
    return reader.request.Body.Read(buff)
}



var backSpace = []byte{13, 10}

//文件上传处理类
//使用者只需继承此类，然后在公用方法里面参与上传事件的处理即可<br>
//最简单的过程即为覆写onFileField()函数
type FileUploadHandlerV1 struct {
    writer http.ResponseWriter
    request *http.Request
    params map[string]*list.List
}


func (handler *FileUploadHandlerV1) writeBack(response string) {
    handler.writer.Write([]byte(response))
}

func (handler *FileUploadHandlerV1) onTextField(name string, value string) {
    logger.Debug("text field:>>", name, "=", value)
    if handler.params == nil {
        handler.params = make(map[string]*list.List)
    }
    ls := handler.params[name]
    if ls == nil {
        ls = new(list.List)
    }
    ls.PushBack(value)
}


func (handler *FileUploadHandlerV1) beginUpload() error {

    logger.Info("begin read file form, content len is", handler.request.ContentLength/1024, "KB")
    //buff := make([]byte, 10240)
    var formReader = &FileFormReader{
        request: handler.request,
        buffer: new(bytes.Buffer),
    }

    headerContentType := handler.request.Header["Content-Type"]
    contentType := ""
    if headerContentType != nil || len(headerContentType) == 0 {
        contentType = headerContentType[0]
    }
    if mat, _ := regexp.Match(ContentTypePattern, []byte(contentType)); mat {
        boundary := regexp.MustCompile(ContentTypePattern).ReplaceAllString(contentType, "${1}")
        paraSeparator := "--" + boundary
        endSeparator := "--" + boundary + "--"
        for {
            line, e := readNextLine(formReader)
            //logger.Debug(">>>>>"+line)
            if e != nil {
                logger.Error("upload error0:", e)
                break
            }
            // if it is paraSeparator, then start read new form text field or file field
            if paraSeparator == line {
                contentDisposition, e1 := readNextLine(formReader)
                if e1 != nil {
                    logger.Error("upload error1:", e1)
                    break
                } else {
                    logger.Debug("contentDisposition>>  ", contentDisposition)
                    mat1, e := regexp.Match(ContentDispositionPattern, []byte(contentDisposition))
                    if e != nil {
                        return e
                    }
                    paramName := ""
                    paramValue := ""
                    if mat1 {
                        paramName = regexp.MustCompile(ContentDispositionPattern).ReplaceAllString(contentDisposition, "${1}")
                    }

                    paramContentType, e2 := readNextLine(formReader)
                    if e2 != nil {
                        logger.Error("upload error2:", e2)
                        break
                    } else {
                        logger.Debug("Content-Type ? >>  ", paramContentType)
                        if paramContentType == "" { // read text parameter field
                            param, e3 := readNextLine(formReader)
                            if e3 != nil {
                                logger.Error("upload error3:", e3)
                                break
                            } else {
                                //logger.Debug("Text parameter >>  ", param)
                                paramValue = param
                                handler.onTextField(paramName, paramValue)
                            }
                        } else { // parse content type
                            _, e3 := readNextLine(formReader)
                            if e3 != nil {
                                logger.Error("upload error4:", e3)
                                break
                            } else { // read file body
                                fi, _ := file.CreateFile("E:/" + timeutil.GetUUID() + ".exe")
                                e4 := readFileBody(formReader, paraSeparator, fi)
                                fi.Close()
                                if e4 != nil {
                                    logger.Error("upload error5:", e4)
                                    break
                                }
                            }
                        }
                    }

                }
            } else if endSeparator == line {
                // form stream hit end
                logger.Info("upload end.")
                break
            } else {
                logger.Error("unknown line")
            }
        }
    }
    return nil
}

func readNextLine(reader *FileFormReader) (string, error) {
    buff := make([]byte, 1)
    last := make([]byte, 2)
    var strBuff bytes.Buffer
    for {
        len, e := reader.Read(buff)
        if e != nil {
            if e != io.EOF {
                return "", e
            }
        }
        if len != 1 {
            return "", errors.New("error read from stream")
        }
        last[0] = last[1]
        last[1] = buff[0]
        strBuff.Write(buff)
        if bytes.Equal(backSpace, last) {
            return string(strBuff.Bytes()[0:strBuff.Len() - 2]), nil
        }
    }
}


func readFileBody(reader *FileFormReader, separator string, writer io.Writer) error {
    separator = "\r\n" + separator
    buff1 := make([]byte, 1024*5)
    buff2 := make([]byte, len(separator))
    tail := make([]byte, len(separator)*2)
    for {
        len1, e1 := reader.Read(buff1)
        if e1 != nil {
            if e1 != io.EOF {
                return e1
            }
        }
        if len1 == 0 {
            return nil
        }
        // whether buff1 contains separator
        i1 := bytes.Index(buff1, []byte(separator))
        if i1 != -1 {
            writer.Write(buff1[0:i1])
            reader.Unread(buff1[i1 + 2:len1]) // skip "\r\n"
            break
        } else {
            len2, e2 := reader.Read(buff2)
            if e2 != nil {
                if e2 != io.EOF {
                    return e2
                }
            }
            if len2 == 0 {
                return nil
            }
            // []byte tail is last bytes of buff1 and first bytes of buff2 in case of broken separator.
            if len1 >= len(separator) {
                ByteCopy(tail, 0, len(separator), buff1[len1 - len(separator):len1])
            }
            if len2 >= len(separator) {
                ByteCopy(tail, len(separator), len(tail), buff2[0:len(separator)])
            }

            i2 := bytes.Index(tail, []byte(separator))
            if i2 != -1 {
                if i2 < len(separator) {
                    writer.Write(buff1[0:len1 - i2])
                    reader.Unread(buff1[len1 - i2 + 2:len1])
                    reader.Unread(buff2[0:len2])
                } else {
                    writer.Write(buff1[0:len1])
                    reader.Unread(buff2[i2 - len(separator) + 2:len2])
                }
                break
            } else {
                writer.Write(buff1[0:len1])
                reader.Unread(buff2[0:len2])
            }
        }
    }
    return nil
}







func ByteCopy(src []byte, start int, end int, cp []byte) {
    for i := range src {
        if i >= start && i < end {
            src[i] = cp[i]
        } else {
            break
        }
    }
}


func parseParameterName(contentDisposition string) {

}









func WebUploadHandlerV1(writer http.ResponseWriter, request *http.Request) {
    handler := &FileUploadHandlerV1{
        writer: writer,
        request: request,
    }
    handler.beginUpload()
}