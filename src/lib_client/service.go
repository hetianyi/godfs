package lib_client

import (
    "lib_common/header"
    "util/logger"
    "lib_common"
    "encoding/json"
    "util/file"
    "io"
    "net"
    "regexp"
    "errors"
    "strconv"
    "sync"
    "app"
)


// each client has one tcp connection with storage server,
// once the connection is broken, the client will destroy.
// one client can only do 1 operation at a time.
var connection net.Conn
var connectString string
var operationLock *sync.Mutex
var secretString string

func NewClient(host string, port int, secret string) {
    operationLock = new(sync.Mutex)
    secretString = secret
    connectString = host + ":" + strconv.Itoa(port)
    con, e := connect()
    if e != nil {
        logger.Error(e)
    }
    connection = con
}

// close client connection and set it to nil
func closeConn() {
    if connection != nil {
        connection.Close()
    }
    connection = nil
}

func connect() (net.Conn, error) {
    logger.Debug("connecting to storage server...")
    con, e := net.Dial("tcp", connectString)
    if e != nil {
        logger.Error(e)
        return nil, e
    }
    e1 := validConn(con)
    if e1 != nil {
        logger.Error("error validate connection:", e1)
        return nil, e1
    }
    logger.Debug("successful validate connection:", e1)
    return con, nil
}

func getConn() net.Conn {
    if connection != nil {
        return connection
    }
    con, e := connect()
    if e != nil {
        logger.Error(e)
        return nil
    }
    connection = con
    return connection
}

func validConn(con net.Conn) error {
    logger.Debug("validate connection...")
    var head = &header.ConnectionHead{Secret: secretString}
    metaLenBytes, bodyLenBytes, metaBytes, e1 := lib_common.PrepareMetaData(0, head)
    if e1 != nil {
        return e1
    }
    e2 := lib_common.WriteMeta(8, metaLenBytes, bodyLenBytes, metaBytes, con)
    if e2 != nil {
        return e2
    }
    _, meta, _, e3 := lib_common.ReadConnMeta(con)
    if e3 != nil {
        return e3
    }
    var response = &header.ConnectionHeadResponse{}
    e4 := json.Unmarshal([]byte(meta), response)
    if e4 != nil {
        return e4
    }
    if response.Status != 0 {
        return errors.New("error connect to storage server, status:" + strconv.Itoa(response.Status))
    }
    return nil
}




//client demo for upload file to storage server.
func Upload(path string) error {
    conn := getConn()
    if conn != nil {
        fi, e := file.GetFile(path)
        if e == nil {
            defer fi.Close()
            fInfo, _ := fi.Stat()
            operation := 2
            meta := &header.UploadRequestMeta{FileSize: fInfo.Size()}

            metaSize, bodySize, metaBytes, e2 := lib_common.PrepareMetaData(fInfo.Size(), meta)
            if e2 != nil {
                closeConn()
                return e2
            }
            e3 := lib_common.WriteMeta(operation, metaSize, bodySize, metaBytes, conn)
            if e3 != nil {
                closeConn()
                return e3
            }
            // begin upload file body bytes
            buff := make([]byte, app.BUFF_SIZE)
            for {
                len5, e4 := fi.Read(buff)
                if e4 != nil && e4 != io.EOF {
                    closeConn()
                    return e4
                }
                if len5 > 0 {
                    len3, e5 := conn.Write(buff[0:len5])
                    if e5 != nil || len3 != len(buff[0:len5]) {
                        closeConn()
                        return e5
                    }
                    if e5 == io.EOF {
                        logger.Debug("upload finish")
                    }
                } else {
                    if e4 != io.EOF {
                        closeConn()
                        return e4
                    } else {
                        logger.Debug("upload finish")
                    }
                    break
                }
            }
            _, respMeta, _, e6 := lib_common.ReadConnMeta(conn)
            if e6 != nil {
                return e6
            }
            var resp = &header.UploadResponseMeta{}
            e7 := json.Unmarshal([]byte(respMeta), resp)
            if e7 != nil {
                closeConn()
                return e7
            }
            if resp.Status != 0 {
                return errors.New("error response status from server:" + strconv.Itoa(resp.Status))
            }
            return nil
        } else {
            return e
        }
    } else {
        return errors.New("error connect to storage server")
    }
}


func CheckFileExists(md5 string) (bool, error) {
    conn := getConn()
    if conn != nil {
        operation := 5
        meta := &header.QueryFileRequestMeta{PathOrMd5: md5}
        metaSize, bodySize, metaBytes, e2 := lib_common.PrepareMetaData(0, meta)
        if e2 != nil {
            closeConn()
            return false, e2
        }
        e3 := lib_common.WriteMeta(operation, metaSize, bodySize, metaBytes, conn)
        if e3 != nil {
            closeConn()
            return false, e3
        }

        _, respMeta, _, e6 := lib_common.ReadConnMeta(conn)
        if e6 != nil {
            return false, e6
        }
        logger.Debug(respMeta)
        var resp = &header.QueryFileResponseMeta{}
        e7 := json.Unmarshal([]byte(respMeta), resp)
        if e7 != nil {
            closeConn()
            return false, e7
        }
        if resp.Status != 0 {
            return false, errors.New("error response status from server:" + strconv.Itoa(resp.Status))
        } else {
            if resp.Exist {
                return true, nil
            }
            return false, nil
        }
        return false, nil
    } else {
        return false, errors.New("error connect to storage server")
    }
}


func DownloadFile(path string, writer io.WriteCloser) error {
    if mat, _ := regexp.Match(app.PATH_REGEX, []byte(path)); !mat {
        return errors.New("file path format error")
    }
    conn := getConn()
    if conn != nil {
        operation := 6
        meta := &header.DownloadFileRequestMeta{Path: path}
        metaSize, bodySize, metaBytes, e2 := lib_common.PrepareMetaData(0, meta)
        if e2 != nil {
            closeConn()
            return e2
        }
        e3 := lib_common.WriteMeta(operation, metaSize, bodySize, metaBytes, conn)
        if e3 != nil {
            closeConn()
            return e3
        }

        _, respMeta, respBodySize, e6 := lib_common.ReadConnMeta(conn)
        if e6 != nil {
            return e6
        }
        var resp = &header.DownloadFileResponseMeta{}
        e7 := json.Unmarshal([]byte(respMeta), resp)
        if e7 != nil {
            closeConn()
            return e7
        }
        if resp.Status != 0 {
            if resp.Status == 4 {
                return errors.New("file not found")
            }
            return errors.New("error response status from server:" + strconv.Itoa(resp.Status))
        } else {
            // begin download
            // begin upload file body bytes
            buff := make([]byte, app.BUFF_SIZE)
            e4 := lib_common.ReadConnDownloadBody(respBodySize, buff, conn, writer)
            if e4 != nil {
                closeConn()
                return e4
            }
            return nil
        }
    } else {
        return errors.New("error connect to storage server")
    }
}
