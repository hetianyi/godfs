package lib_client

import (
    "util/logger"
    "encoding/json"
    "util/file"
    "io"
    "regexp"
    "errors"
    "strconv"
    "sync"
    "app"
    "lib_common/bridge"
    "container/list"
    "math/rand"
    "strings"
    "lib_common"
    "time"
)


// each client has one tcp connection with storage server,
// once the connection is broken, the client will destroy.
// one client can only do 1 operation at a time.
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
    TrackerMaintainer *TrackerMaintainer
    connPool *ClientConnectionPool
    MaxConnPerServer int // 客户端和每个服务建立的最大连接数，web项目中建议设置为和最大线程相同的数量
}


func NewClient(MaxConnPerServer int) *Client {
    logger.Debug("init godfs client.")
    connPool := &ClientConnectionPool{}
    connPool.Init(MaxConnPerServer)
    return &Client{connPool: connPool}
}


//client demo for upload file to storage server.
func (client *Client) Upload(path string, group string, startTime time.Time) (string, error) {
    fi, e := file.GetFile(path)
    if e == nil {
        defer fi.Close()
        logger.Info("upload file:", fi.Name())
        logger.Debug("pre check file md5:", fi.Name())
        md5, ee := file.GetFileMd5(path)
        if ee == nil {
            qfi, ee1 := client.QueryFile(md5)
            if qfi != nil {
                sm := "S"
                if qfi.PartNum > 1 {
                    sm = "M"
                }
                return qfi.Group + "/" + qfi.Instance + "/" + sm + "/" + qfi.Md5, nil
            } else {
                logger.Debug("error query file info from tracker server:", ee1)
            }
        } else {
            logger.Debug("error check file md5:", ee,", skip pre check.")
        }
        fInfo, _ := fi.Stat()
        uploadMeta := &bridge.OperationUploadFileRequest{
            FileSize: uint64(fInfo.Size()),
            FileExt: file.GetFileExt(fInfo.Name()),
            Md5: "",
        }

        var excludes list.List
        var connBridge *bridge.Bridge
        var member *bridge.ExpireMember
        for {
            mem := selectStorageServer(group, "", &excludes)
            // no available storage
            if mem == nil {
                return "", NO_STORAGE_ERROR
            }
            logger.Info("using storage server:", mem.BindAddr + ":" + strconv.Itoa(mem.Port))
            cb, e12 := client.connPool.GetConnBridge(mem)
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
            var finish, total int64
            var stopFlag = false
            defer func() {stopFlag = true}()
            total = fInfo.Size()
            finish = 0
            go lib_common.ShowPercent(&total, &finish, &stopFlag, startTime)
            for {
                len5, e4 := fi.Read(buff)
                if e4 != nil && e4 != io.EOF {
                    return e4
                }
                if len5 > 0 {
                    len3, e5 := out.Write(buff[0:len5])
                    finish += int64(len5)
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
            client.connPool.ReturnBrokenConnBridge(member, connBridge)
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
            client.connPool.ReturnBrokenConnBridge(member, connBridge)
            return "", e3
        }
        client.connPool.ReturnConnBridge(member, connBridge)
        return fid, nil
    } else {
        return "", e
    }
}




func (client *Client) QueryFile(pathOrMd5 string) (*bridge.File, error) {
    logger.Debug("query file info:", pathOrMd5)
    var result *bridge.File
    for ele := client.TrackerMaintainer.TrackerInstances.Front(); ele != nil; ele = ele.Next() {
        queryMeta := &bridge.OperationQueryFileRequest{PathOrMd5: pathOrMd5}
        connBridge := ele.Value.(*TrackerInstance).connBridge
        e11 := connBridge.SendRequest(bridge.O_QUERY_FILE, queryMeta, 0, nil)
        if e11 != nil {
            connBridge.Close()
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
    return download(path, start, offset, false, client, writerHandler)
}

func download(path string, start int64, offset int64, fromSrc bool, client *Client, writerHandler func(fileLen uint64, reader io.Reader) error) error {
    downloadMeta := &bridge.OperationDownloadFileRequest {
        Path: path,
        Start: start,
        Offset: offset,
    }
    group := regexp.MustCompile(app.PATH_REGEX).ReplaceAllString(path, "${1}")
    instanceId := regexp.MustCompile(app.PATH_REGEX).ReplaceAllString(path, "${2}")

    var excludes list.List
    var connBridge *bridge.Bridge
    var member *bridge.ExpireMember
    for {
        mem := selectStorageServer(group, "", &excludes)
        // no available storage
        if mem == nil {
            return NO_STORAGE_ERROR
        }
        logger.Info("using storage server:", mem.BindAddr + ":" + strconv.Itoa(mem.Port))
        cb, e12 := client.connPool.GetConnBridge(mem)
        if e12 != nil {
            logger.Error(e12)
            if e12 != MAX_CONN_EXCEED_ERROR {
                excludes.PushBack(mem)
            }
            continue
        }
        connBridge = cb
        member = mem
        break
    }
    logger.Debug("download from:", *member)

    e2 := connBridge.SendRequest(bridge.O_DOWNLOAD_FILE, downloadMeta, 0, nil)
    if e2 != nil {
        client.connPool.ReturnBrokenConnBridge(member, connBridge)
        // if download fail, try to download from file source server
        if !fromSrc && member.InstanceId != instanceId {
            return download(path, start, offset, true, client, writerHandler)
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
        client.connPool.ReturnBrokenConnBridge(member, connBridge)
        // if download fail, try to download from file source server
        if !fromSrc && member.InstanceId != instanceId {
            return download(path, start, offset, true, client, writerHandler)
        }
        return e3
    } else {
        client.connPool.ReturnConnBridge(member, connBridge)
    }
    return nil
}



// TODO 新增连接池
// select a storage server matching given group and instanceId
// excludes contains fail storage and not gonna use this time.
func selectStorageServer(group string, instanceId string, excludes *list.List) *bridge.ExpireMember {
    memberIteLock.Lock()
    defer memberIteLock.Unlock()
    var pick list.List
    for ele := GroupMembers.Front(); ele != nil; ele = ele.Next() {
        b := ele.Value.(*bridge.ExpireMember)
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
            return ele.Value.(*bridge.ExpireMember)
        }
        index++
    }
    return nil
}

func containsMember(mem *bridge.ExpireMember, excludes *list.List) bool {
    if excludes == nil {
        return false
    }
    uid := GetStorageServerUID(mem)
    for ele := excludes.Front(); ele != nil; ele = ele.Next() {
        if GetStorageServerUID(ele.Value.(*bridge.ExpireMember)) == uid {
            return true
        }
    }
    return false
}






