package lib_client

import (
    "util/logger"
    "encoding/json"
    "util/file"
    "io"
    "net"
    "regexp"
    "errors"
    "strconv"
    "sync"
    "app"
    "lib_common/bridge"
)


// each client has one tcp connection with storage server,
// once the connection is broken, the client will destroy.
// one client can only do 1 operation at a time.
var connection net.Conn
var operationLock *sync.Mutex
var secretString string
var connBridge *bridge.Bridge


type IClient interface {
    Close()
    Upload(path string) (string, error)
    CheckFileExists(md5 string) (bool, error)
    DownloadFile(path string, writerHandler func(fileLen uint64, writer io.WriteCloser) error) error
}

type Client struct {
    connBridge *bridge.Bridge
    buffer []byte
}

func NewClient(host string, port int, secret string) (*Client, error) {
    operationLock = new(sync.Mutex)
    secretString = secret
    connectString := host + ":" + strconv.Itoa(port)
    connBridge, e := connect(connectString, secret)
    if e != nil {
        return nil, e
    }
    client := &Client{
        connBridge: connBridge,
        buffer: make([]byte, app.BUFF_SIZE),
    }
    return client, nil
}

func connect(connectString string, secret string) (*bridge.Bridge, error) {
    logger.Debug("connecting to storage server...")
    con, e := net.Dial("tcp", connectString)
    if e != nil {
        logger.Error(e)
        return nil, e
    }
    connBridge := bridge.NewBridge(con)
    e1 := connBridge.ValidateConnection(secret)
    if e1 != nil {
        return nil, e1
    }
    logger.Debug("successful validate connection:", e1)
    return connBridge, nil
}


func (client *Client) Close() {
    client.connBridge.Close()
}


//client demo for upload file to storage server.
func (client *Client) Upload(path string) (string, error) {
    fi, e := file.GetFile(path)
    if e == nil {
        defer fi.Close()
        fInfo, _ := fi.Stat()

        uploadMeta := &bridge.OperationUploadFileRequest{
            FileSize: uint64(fInfo.Size()),
            FileExt: file.GetFileExt(fInfo.Name()),
            Md5: "",
        }
        e2 := client.connBridge.SendRequest(bridge.O_UPLOAD, uploadMeta, uint64(fInfo.Size()), func(out io.WriteCloser) error {
            // begin upload file body bytes
            buff := client.buffer
            for {
                len5, e4 := fi.Read(buff)
                if e4 != nil && e4 != io.EOF {
                    return e4
                }
                if len5 > 0 {
                    len3, e5 := out.Write(buff[0:len5])
                    if e5 != nil || len3 != len(buff[0:len5]) {
                        return e5
                    }
                    if e5 == io.EOF {
                        logger.Debug("upload finish")
                    }
                } else {
                    if e4 != io.EOF {
                        return e4
                    } else {
                        logger.Debug("upload finish")
                    }
                    break
                }
            }
            return nil
        })
        if e2 != nil {
            return "", e2
        }

        var fid string
        // receive response
        e3 := client.connBridge.ReceiveResponse(func(response *bridge.Meta, in io.Reader) error {
            if response.Err != nil {
                return response.Err
            }
            var uploadResponse = &bridge.OperationUploadFileResponse{}
            e4 := json.Unmarshal(response.MetaBody, uploadResponse)
            if e4 != nil {
                return e4
            }
            if uploadResponse.Status != bridge.STATUS_OK {
                return errors.New("error connect to server, server response status:" + strconv.Itoa(uploadResponse.Status))
            }
            fid = uploadResponse.Path
            // connect success
            return nil
        })
        if e3 != nil {
            return "", e3
        }
        return fid, nil
    } else {
        return "", e
    }
}


func (client *Client) CheckFileExists(pathOrMd5 string) (bool, error) {
    queryMeta := &bridge.OperationQueryFileRequest{PathOrMd5: pathOrMd5}
    e2 := client.connBridge.SendRequest(bridge.O_QUERY_FILE, queryMeta, 0, nil)
    if e2 != nil {
        return false, e2
    }

    var exist bool
    // receive response
    e3 := client.connBridge.ReceiveResponse(func(response *bridge.Meta, in io.Reader) error {
        if response.Err != nil {
            return response.Err
        }
        var queryResponse = &bridge.OperationQueryFileResponse{}
        e4 := json.Unmarshal(response.MetaBody, queryResponse)
        if e4 != nil {
            return e4
        }
        if queryResponse.Status != bridge.STATUS_OK && queryResponse.Status != bridge.STATUS_NOT_FOUND {
            return errors.New("error connect to server, server response status:" + strconv.Itoa(queryResponse.Status))
        }
        exist = queryResponse.Exist
        // connect success
        return nil
    })
    if e3 != nil {
        return false, e3
    }
    return exist, nil
}


func (client *Client) DownloadFile(path string, start int64, offset int64, writerHandler func(fileLen uint64, reader io.Reader) error) error {
    if mat, _ := regexp.Match(app.PATH_REGEX, []byte(path)); !mat {
        return errors.New("file path format error")
    }
    downloadMeta := &bridge.OperationDownloadFileRequest{
        Path: path,
        Start: start,
        Offset: offset,
    }
    e2 := client.connBridge.SendRequest(bridge.O_DOWNLOAD_FILE, downloadMeta, 0, nil)
    if e2 != nil {
        return e2
    }

    // receive response
    e3 := client.connBridge.ReceiveResponse(func(response *bridge.Meta, in io.Reader) error {
        if response.Err != nil {
            return response.Err
        }
        var downloadResponse = &bridge.OperationDownloadFileResponse{}
        e4 := json.Unmarshal(response.MetaBody, downloadResponse)
        if e4 != nil {
            return e4
        }
        if downloadResponse.Status == bridge.STATUS_NOT_FOUND {
            return bridge.FILE_NOT_FOUND_ERROR
        }
        if downloadResponse.Status != bridge.STATUS_OK {
            logger.Error("error connect to server, server response status:" + strconv.Itoa(downloadResponse.Status))
            return bridge.DOWNLOAD_FILE_ERROR
        }
        return writerHandler(response.BodyLength, client.connBridge.GetConn())
    })
    if e3 != nil {
        return e3
    }
    return nil
}
