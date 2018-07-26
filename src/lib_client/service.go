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
    "strings"
)


// each client has one tcp connection with storage server,
// once the connection is broken, the client will destroy.
// one client can only do 1 operation at a time.
var StorageServers list.List
var addLock *sync.Mutex
var NO_TRACKER_ERROR = errors.New("no tracker server available")
var NO_STORAGE_ERROR = errors.New("no storage server available")

func init() {
    addLock = new(sync.Mutex)
}

type IClient interface {
    Close()
    Upload(path string) (string, error)
    QueryFile(md5 string) (*bridge.File, error)
    DownloadFile(path string, writerHandler func(fileLen uint64, writer io.WriteCloser) error) error
}

type Client struct {
    //operationLock *sync.Mutex
    TrackerManagers list.List
}

type TrackerManager struct {
    brokenChan *chan int
    connBridge *bridge.Bridge
}

func NewClient() (*Client, error) {
    ls := lib_common.ParseTrackers(app.TRACKERS)
    if ls.Len() == 0 {
        return nil, NO_TRACKER_ERROR
    }
    client := &Client{}

    var chanList list.List
    for e := ls.Front(); e != nil; e = e.Next() {
        cha := make(chan int)
        chanList.PushBack(cha)
        manageTracker(e.Value.(string), client, &cha)
    }
    // wait till all tracker storage server info sync finished.
    /*for e := chanList.Front(); e != nil; e = e.Next() {
        <- e.Value.(chan int)
    }*/
    if client.TrackerManagers.Len() == 0 {
        logger.Fatal("no tracker available")
    }
    return client, nil
}


func manageTracker(connectString string, client *Client, cha *chan int) {
    /*var ele *list.Element
    brokenChan := make(chan int)
    for {
        connBridge := initTracker(connectString)
        if connBridge != nil {
            tracker := &TrackerManager{connBridge: connBridge, brokenChan: &brokenChan}
            ele = client.TrackerManagers.PushBack(tracker)
            *cha <- 1
            break
        }
        time.Sleep(time.Second * 10)
    }
    for {
        // if bridge is broken, it will notify chan here.
        <- brokenChan
        ele.Value.(*TrackerManager).connBridge = nil
        connBridge := initTracker(connectString)
        if connBridge != nil {
            ele.Value.(*TrackerManager).connBridge = connBridge
        }
        time.Sleep(time.Second * 10)
    }*/

    connBridge := initTracker(connectString)
    if connBridge != nil {
        brokenChan := make(chan int)
        tracker := &TrackerManager{connBridge: connBridge, brokenChan: &brokenChan}
        client.TrackerManagers.PushBack(tracker)
    }
}

func initTracker(connectString string) *bridge.Bridge {
    logger.Debug("connecting to storage server...")
    con, e := net.Dial("tcp", connectString)
    if e != nil {
        logger.Error(e)
        return nil
    }
    connBridge := bridge.NewBridge(con)
    e1 := connBridge.ValidateConnection(app.SECRET)
    if e1 != nil {
        logger.Error(e1)
        connBridge.Close()
        return nil
    }
    logger.Debug("successful validate connection:", e1)

    syncMeta := &bridge.OperationGetStorageServerRequest {}
    // send validate request
    e5 := connBridge.SendRequest(bridge.O_SYNC_STORAGE, syncMeta, 0, nil)
    if e5 != nil {
        logger.Error(e5)
        connBridge.Close()
        return nil
    }
    e6 := connBridge.ReceiveResponse(func(response *bridge.Meta, in io.Reader) error {
        if response.Err != nil {
            logger.Error(response.Err)
            return response.Err
        }
        var validateResp = &bridge.OperationGetStorageServerResponse{}
        logger.Debug("sync storage server response:", string(response.MetaBody))
        e3 := json.Unmarshal(response.MetaBody, validateResp)
        if e3 != nil {
            logger.Error(e3)
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
        logger.Error(e6)
        connBridge.Close()
        return nil
    }
    logger.Debug("successful validate connection:", e1)
    return connBridge
}


func addStorageServer(server *bridge.Member) {
    addLock.Lock()
    defer addLock.Unlock()
    s, _ := json.Marshal(server)
    logger.Debug("add storage server:", string(s))
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
    for ele := client.TrackerManagers.Front(); ele != nil; ele = ele.Next() {
        b := ele.Value.(*TrackerManager)
        logger.Debug("shutdown bridge ", b.connBridge.GetConn().RemoteAddr())
        b.connBridge.Close()
    }
}


//client demo for upload file to storage server.
func (client *Client) Upload(path string, group string) (string, error) {
    fi, e := file.GetFile(path)
    if e == nil {
        defer fi.Close()
        fInfo, _ := fi.Stat()

        uploadMeta := &bridge.OperationUploadFileRequest{
            FileSize: uint64(fInfo.Size()),
            FileExt: file.GetFileExt(fInfo.Name()),
            Md5: "",
        }

        var excludes list.List
        var connBridge *bridge.Bridge
        var member *bridge.Member
        for {
            mem := selectStorageServer(group, "", &excludes)
            // no available storage
            if mem == nil {
                return "", NO_STORAGE_ERROR
            }
            cb, e12 := GetConnBridge(mem)
            if e12 != nil {
                excludes.PushBack(mem)
                continue
            }
            connBridge = cb
            member = mem
            break
        }

        e2 := connBridge.SendRequest(bridge.O_UPLOAD, uploadMeta, uint64(fInfo.Size()), func(out io.WriteCloser) error {
            // begin upload file body bytes
            buff := make([]byte, app.BUFF_SIZE)
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
            connBridge.Close()
            return "", e2
        }

        var fid string
        // receive response
        e3 := connBridge.ReceiveResponse(func(response *bridge.Meta, in io.Reader) error {
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
            connBridge.Close()
            return "", e3
        }
        ReturnConnBridge(member, connBridge)
        return fid, nil
    } else {
        return "", e
    }
}




func (client *Client) QueryFile(pathOrMd5 string) (*bridge.File, error) {

    var result *bridge.File
    for ele := client.TrackerManagers.Front(); ele != nil; ele = ele.Next() {
        queryMeta := &bridge.OperationQueryFileRequest{PathOrMd5: pathOrMd5}
        connBridge := ele.Value.(*TrackerManager).connBridge
        e11 := connBridge.SendRequest(bridge.O_QUERY_FILE, queryMeta, 0, nil)
        if e11 != nil {
            connBridge.Close()
            *ele.Value.(*TrackerManager).brokenChan <- 1
            continue
        }
        e12 := connBridge.ReceiveResponse(func(response *bridge.Meta, in io.Reader) error {
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
            result = queryResponse.File
            return nil
        })
        if e12 != nil {
            connBridge.Close()
            *ele.Value.(*TrackerManager).brokenChan <- 1
            continue
        }
        if result != nil {
            return result, nil
        }
    }
    return result, nil
}


func (client *Client) DownloadFile(path string, start int64, offset int64, writerHandler func(fileLen uint64, reader io.Reader) error) error {
    path = strings.TrimSpace(path)
    if strings.Index(path, "/") != 0 {
        path = "/" + path
    }
    if mat, _ := regexp.Match(app.PATH_REGEX, []byte(path)); !mat {
        return errors.New("file path format error")
    }
    return download(path, start, offset, false, writerHandler)
}

func download(path string, start int64, offset int64, fromSrc bool, writerHandler func(fileLen uint64, reader io.Reader) error) error {
    downloadMeta := &bridge.OperationDownloadFileRequest{
        Path: path,
        Start: start,
        Offset: offset,
    }
    group := regexp.MustCompile(app.PATH_REGEX).ReplaceAllString(path, "${1}")
    instanceId := regexp.MustCompile(app.PATH_REGEX).ReplaceAllString(path, "${2}")

    var excludes list.List
    var connBridge *bridge.Bridge
    var member *bridge.Member
    for {
        mem := selectStorageServer(group, "", &excludes)
        // no available storage
        if mem == nil {
            return NO_TRACKER_ERROR
        }
        cb, e12 := GetConnBridge(mem)
        if e12 != nil {
            excludes.PushBack(mem)
            continue
        }
        connBridge = cb
        member = mem
        break
    }
    logger.Debug("download from:", *member)

    e2 := connBridge.SendRequest(bridge.O_DOWNLOAD_FILE, downloadMeta, 0, nil)
    if e2 != nil {
        // if download fail, try to download from file source server
        if !fromSrc && member.InstanceId != instanceId {
            return download(path, start, offset, true, writerHandler)
        }
        return e2
    }

    // receive response
    e3 := connBridge.ReceiveResponse(func(response *bridge.Meta, in io.Reader) error {
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
        return writerHandler(response.BodyLength, connBridge.GetConn())
    })
    if e3 != nil {
        // if download fail, try to download from file source server
        if !fromSrc && member.InstanceId != instanceId {
            return download(path, start, offset, true, writerHandler)
        }
        return e3
    }
    return nil
}



// TODO 新增连接池
// select a storage server matching given group and instanceId
// excludes contains fail storage and not gonna use this time.
func selectStorageServer(group string, instanceId string, excludes *list.List) *bridge.Member {
    var pick list.List
    for ele := StorageServers.Front(); ele != nil; ele = ele.Next() {
        b := ele.Value.(*bridge.Member)
        if containsMember(b, excludes) {
            continue
        }
        match1 := false
        match2 := false
        if group == "" || group == b.Group {
            match1 = true
        }
        if instanceId =="" || instanceId == b.InstanceId {
            match2 = true
        }
        if match1 && match2 {
            pick.PushBack(b)
        }
    }
    if pick.Len() == 0 {
        return nil
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

func containsMember(mem *bridge.Member, excludes *list.List) bool {
    if excludes == nil {
        return false
    }
    uid := GetStorageServerUID(mem)
    for ele := excludes.Front(); ele != nil; ele = ele.Next() {
        if GetStorageServerUID(ele.Value.(*bridge.Member)) == uid {
            return true
        }
    }
    return false
}






