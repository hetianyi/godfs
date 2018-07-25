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
    "lib_common"
    "container/list"
    "math/rand"
)


// each client has one tcp connection with storage server,
// once the connection is broken, the client will destroy.
// one client can only do 1 operation at a time.
var StorageServers list.List
var addLock *sync.Mutex

func init() {
    addLock = new(sync.Mutex)
}

type IClient interface {
    Close()
    Upload(path string) (string, error)
    CheckFileExists(md5 string) (bool, error)
    DownloadFile(path string, writerHandler func(fileLen uint64, writer io.WriteCloser) error) error
}

type Client struct {
    buffer []byte
    //operationLock *sync.Mutex
    trackersConnBridge list.List
}

func NewClient() (*Client, error) {
    ls := lib_common.ParseTrackers(app.TRACKERS)
    if ls.Len() == 0 {
        return nil, errors.New("no tracker server configured")
    }
    client := &Client{
        buffer: make([]byte, app.BUFF_SIZE),
        //operationLock: new(sync.Mutex),
    }

    for e := ls.Front(); e != nil; e = e.Next() {
        trackerConnBridge, err := connectTracker(e.Value.(string))
        if err != nil {
            logger.Debug("error connect to tracker server", e.Value.(string), ":", err)
        } else {
            client.trackersConnBridge.PushBack(trackerConnBridge)
        }
    }
    return client, nil
}

func connectTracker(connectString string) (*bridge.Bridge, error) {
    logger.Debug("connecting to storage server...")
    con, e := net.Dial("tcp", connectString)
    if e != nil {
        logger.Error(e)
        return nil, e
    }
    connBridge := bridge.NewBridge(con)
    e1 := connBridge.ValidateConnection(app.SECRET)
    if e1 != nil {
        return nil, e1
    }
    logger.Debug("successful validate connection:", e1)


    syncMeta := &bridge.OperationGetStorageServerRequest {}
    // send validate request
    e5 := connBridge.SendRequest(bridge.O_SYNC_STORAGE, syncMeta, 0, nil)
    if e5 != nil {
        return nil, e5
    }
    e6 := connBridge.ReceiveResponse(func(response *bridge.Meta, in io.Reader) error {
        if response.Err != nil {
            return response.Err
        }
        var validateResp = &bridge.OperationGetStorageServerResponse{}
        logger.Debug("sync storage server response:", string(response.MetaBody))
        e3 := json.Unmarshal(response.MetaBody, validateResp)
        if e3 != nil {
            return e3
        }
        if validateResp.Status != bridge.STATUS_OK {
            return errors.New("error connect to server, server response status:" + strconv.Itoa(validateResp.Status))
        }
        if nil != validateResp.GroupMembers {
            for i := range validateResp.GroupMembers {
                addStorageServer(&validateResp.GroupMembers[i])
            }
        }
        // connect success
        return nil
    })
    if e6 != nil {
        return nil, e6
    }
    logger.Debug("successful validate connection:", e1)
    return connBridge, nil
}

func addStorageServer(server *bridge.Member) {
    addLock.Lock()
    defer addLock.Unlock()
    s, _ := json.Marshal(server)
    logger.Debug("add storage server:", s)
    uid := GetStorageServerUID(server)
    for ele := StorageServers.Front(); ele != nil; ele = ele.Next() {
        mem := ele.Value.(*bridge.Member)
        euid := GetStorageServerUID(mem)
        if uid == euid {
            logger.Debug("storage server exists, ignore:", s)
            return
        }
    }
    StorageServers.PushBack(server)
}


func (client *Client) Close() {
    for ele := client.trackersConnBridge.Front(); ele != nil; ele = ele.Next() {
        b := ele.Value.(*bridge.Bridge)
        logger.Debug("shutdown bridge ", b.GetConn().RemoteAddr())
        b.Close()
    }
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



// TODO 新增连接池
// select a storage server matching given group and instanceId
func selectStorageServer(group string, instanceId string) *bridge.Member {
    var pick list.List
    for ele := StorageServers.Front(); ele != nil; ele = ele.Next() {
        b := ele.Value.(*bridge.Member)
        match1 := false
        match2 := false
        if group == b.Group {
            match1 = true
        }
        if instanceId == b.InstanceId {
            match2 = true
        }
        if match1 && match2 {
            pick.PushBack(b)
        }
    }
    rd := rand.Intn(pick.Len())
    index := 0
    for ele := pick.Front(); ele != nil; ele = ele.Next() {
        if index == rd {
            return ele.Value.(*bridge.Member)
        }
        index++
    }
    return nil
}






