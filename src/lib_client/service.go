package lib_client

import (
    "bytes"
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
)


// each client has one tcp connection with storage server,
// once the connection is broken, the client will destroy.
// one client can only do 1 operation at a time.
var conn net.Conn
var connecString string
var operationLock *sync.Mutex

func NewClient(host string, port int) {
    operationLock = make()
    connecString = host + ":" + strconv.Itoa(port)
    con, e := connect()
    if e != nil {
        logger.Error(e)
    }
    conn = con
}

// close client connection and set it to nil
func closeConn() {
    if conn != nil {
        conn.Close()
    }
    conn = nil
}

func connect() (net.Conn, error) {
    con, e := net.Dial("tcp", connecString)
    if e != nil {
        logger.Error(e)
        return nil, e
    }
    return con, nil
}

func getConn() net.Conn {
    if conn != nil {
        return conn
    }
    return connect(),_

}



//client demo for upload file to storage server.
func Upload(path string) error {
    conn, e := net.Dial("tcp", "127.0.0.1:1024")
    if e == nil {
        fi, e := file.GetFile(path)
        if e == nil {
            fInfo, _ := fi.Stat()

            operation := 2
            meta := &header.UploadRequestMeta{
                Secret: "OASAD834jA97AAQE761==",
                FileSize: fInfo.Size(),
            }

            metaSize, bodySize, metaBytes, e2 := lib_common.PrepareMetaData(fInfo.Size(), meta)
            if e2 != nil {
                logger.Fatal("meta prepare failed")
            }

            var headerBuff bytes.Buffer
            headerBuff.Write(header.OperationHeadByteMap[operation])
            headerBuff.Write(metaSize)
            headerBuff.Write(bodySize)

            len1, e2 := conn.Write(headerBuff.Bytes())
            if e2 != nil || len1 != headerBuff.Len() {
                logger.Fatal("error write meta len")
            }
            len2, e3 := conn.Write(metaBytes)
            if e3 != nil || len2 != len(metaBytes) {
                logger.Fatal("error write meta")
            }

            buff := make([]byte, 1024*30)
            for {
                len5, e := fi.Read(buff)
                if len5 > 0 {
                    len3, e4 := conn.Write(buff[0:len5])
                    if e4 != nil || len3 != len(buff[0:len5]) {
                        lib_common.Close(conn)
                        logger.Fatal("error write body:", e4)
                    }
                } else {
                    if e != io.EOF {
                        lib_common.Close(conn)
                        logger.Error(e)
                    } else {
                        logger.Info("上传完毕")
                    }
                    break
                }
            }
            _, respMeta, _, e6 := lib_common.ReadConnMeta(conn)
            if e6 != nil {
                logger.Fatal("error read response:", e6)
            }
            var resp = &header.UploadResponseMeta{}
            e7 := json.Unmarshal([]byte(respMeta), resp)
            if e7 != nil {
                lib_common.Close(conn)
                logger.Error(e7)
            }
            logger.Info(respMeta)
        } else {
            logger.Fatal("error open file:", e)
        }
    } else {
        logger.Error("error connect to storage server")
    }
    return e
}


func CheckFileExists(md5 string) error {
    conn, e := net.Dial("tcp", "127.0.0.1:1024")
    if e == nil {
        defer conn.Close()
        operation := 5
        meta := &header.QueryFileRequestMeta{
            Secret: "OASAD834jA97AAQE761==",
            Md5: md5,
        }

        metaSize, bodySize, metaBytes, e2 := lib_common.PrepareMetaData(0, meta)
        if e2 != nil {
            logger.Fatal("meta prepare failed")
        }

        var headerBuff bytes.Buffer
        headerBuff.Write(header.OperationHeadByteMap[operation])
        headerBuff.Write(metaSize)
        headerBuff.Write(bodySize)

        len1, e2 := conn.Write(headerBuff.Bytes())
        if e2 != nil || len1 != headerBuff.Len() {
            logger.Fatal("error write meta len")
        }
        len2, e3 := conn.Write(metaBytes)
        if e3 != nil || len2 != len(metaBytes) {
            logger.Fatal("error write meta")
        }

        _, respMeta, _, e6 := lib_common.ReadConnMeta(conn)
        if e6 != nil {
            logger.Fatal("error read response:", e6)
        }
        var resp = &header.UploadResponseMeta{}
        e7 := json.Unmarshal([]byte(respMeta), resp)
        if e7 != nil {
            lib_common.Close(conn)
            logger.Error(e7)
        }
        logger.Info(respMeta)
    } else {
        logger.Error("error connect to storage server")
    }
    return e
}


func DownloadFile(path string, writer io.WriteCloser) error {
    // TODO move
    pathRegex := "^/([0-9a-zA-Z_]+)/([0-9a-zA-Z_]+)/([0-9a-fA-F]{32})$"
    if mat, _ := regexp.Match(pathRegex, []byte(path)); !mat {
        return errors.New("path format error")
    }
    conn, e := net.Dial("tcp", "127.0.0.1:1024")
    if e == nil {
        defer conn.Close()
        operation := 6
        meta := &header.DownloadFileRequestMeta{
            Secret: "OASAD834jA97AAQE761==",
            Path: path,
        }

        metaSize, bodySize, metaBytes, e2 := lib_common.PrepareMetaData(0, meta)
        if e2 != nil {
            logger.Fatal("meta prepare failed")
        }

        var headerBuff bytes.Buffer
        headerBuff.Write(header.OperationHeadByteMap[operation])
        headerBuff.Write(metaSize)
        headerBuff.Write(bodySize)

        len1, e2 := conn.Write(headerBuff.Bytes())
        if e2 != nil || len1 != headerBuff.Len() {
            logger.Fatal("error write meta len")
        }
        len2, e3 := conn.Write(metaBytes)
        if e3 != nil || len2 != len(metaBytes) {
            logger.Fatal("error write meta")
        }

        _, respMeta, _, e6 := lib_common.ReadConnMeta(conn)
        if e6 != nil {
            logger.Fatal("error read response:", e6)
        }
        var resp = &header.UploadResponseMeta{}
        e7 := json.Unmarshal([]byte(respMeta), resp)
        if e7 != nil {
            lib_common.Close(conn)
            logger.Error(e7)
        }
        logger.Info(respMeta)
        if resp.Exist && resp.Status == 0 {
            // server status is ok, begin download
            fileLen := resp.FileSize
            buffer := make([]byte, 1024*30)
            e11 := lib_common.ReadConnDownloadBody(uint64(fileLen), buffer, conn, writer)
            if e11 != nil {
                logger.Error("download failed:", e11)
                return e11
            }
            logger.Info("download success")
        } else {
            logger.Error("file not found")
        }
    } else {
        logger.Error("error connect to storage server")
    }
    return e
}
