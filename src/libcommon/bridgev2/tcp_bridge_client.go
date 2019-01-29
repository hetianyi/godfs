package bridgev2

import (
    "app"
    "crypto/md5"
    "errors"
    "strconv"
    json "github.com/json-iterator/go"
    "util/logger"
)

type TcpBridgeClient struct {
    // storage server info
    server *app.ServerInfo
    connManager *ConnectionManager
}

// create a new instance for bridgev2.Server
func NewTcpClient(server *app.ServerInfo) *TcpBridgeClient {
    return &TcpBridgeClient {server, nil}
}

func (client *TcpBridgeClient) GetConnManager() *ConnectionManager {
    if client.connManager == nil {
        return &ConnectionManager{}
    }
	return client.connManager
}

// connect to server
func (client *TcpBridgeClient) Connect() error {
    if client.connManager != nil && client.connManager.State > StateNotConnect {
        panic(errors.New("already connected"))
    }
    conn, err := connPool.GetConn(client.server)
    if err != nil {
        return err
    }
    h, p := client.server.GetHostAndPortByAccessFlag()
    logger.Debug("connect to", h + ":" + strconv.Itoa(p), "success")
    client.connManager = &ConnectionManager{
        server: client.server,
        Conn: conn,
        Side: ClientSide,
        Md: md5.New(),
    }
    client.connManager.State = StateConnected
    return nil
}


// validate this connection.
func (client *TcpBridgeClient) Validate() (*ConnectResponseMeta, error) {
    meta := &ConnectMeta{
        Secret: app.Secret,
        UUID: app.UUID,
    }
    frame, e := client.sendReceive(FrameOperationValidate, StateConnected, meta, 0, nil)
    if e != nil {
        return nil, e
    }
    var res = &ConnectResponseMeta{}
    e1 := json.Unmarshal(frame.FrameMeta, res)
    if e1 != nil {
        return nil, e1
    }
    if frame.GetStatus() == StatusSuccess {
        client.connManager.State = StateValidated
        client.connManager.UUID = res.UUID
    }
    return res, nil
}


// synchronized storage members.
func (client *TcpBridgeClient) SyncStorageMembers(storage *app.StorageDO) (*SyncStorageMembersResponseMeta, error) {
    frame, e := client.sendReceive(FrameOperationSyncStorageMember, StateValidated, storage, 0, nil)
    if e != nil {
        return nil, e
    }
    var res = &SyncStorageMembersResponseMeta{}
    e1 := json.Unmarshal(frame.FrameMeta, res)
    if e1 != nil {
        return nil, e1
    }
    return res, nil
}

// synchronized storage members.
func (client *TcpBridgeClient) SyncAllStorageServers(meta *SyncAllStorageServerMeta) (*SyncAllStorageServerResponseMeta, error) {
    frame, e := client.sendReceive(FrameOperationSyncAllStorageServers, StateValidated, meta, 0, nil)
    if e != nil {
        return nil, e
    }
    var res = &SyncAllStorageServerResponseMeta{}
    e1 := json.Unmarshal(frame.FrameMeta, res)
    if e1 != nil {
        return nil, e1
    }
    return res, nil
}

// register files to tracker
func (client *TcpBridgeClient) RegisterFiles(meta *RegisterFileMeta) (*RegisterFileResponseMeta, error) {
    frame, e := client.sendReceive(FrameOperationRegisterFiles, StateValidated, meta, 0, nil)
    if e != nil {
        return nil, e
    }
    var res = &RegisterFileResponseMeta{}
    e1 := json.Unmarshal(frame.FrameMeta, res)
    if e1 != nil {
        return nil, e1
    }
    return res, nil
}


// pull files from tracker
func (client *TcpBridgeClient) PullFiles(meta *PullFileMeta) (*PullFileResponseMeta, error) {
    frame, e := client.sendReceive(FrameOperationPullNewFiles, StateValidated, meta, 0, nil)
    if e != nil {
        return nil, e
    }
    var res = &PullFileResponseMeta{}
    e1 := json.Unmarshal(frame.FrameMeta, res)
    if e1 != nil {
        return nil, e1
    }
    return res, nil
}


// upload file to storage server.
func (client *TcpBridgeClient) UploadFile(meta *UploadFileMeta,
                                          bodyWriterHandler func(manager *ConnectionManager, frame *Frame) error,
                                         ) (*UploadFileResponseMeta, error) {
    frame, e := client.sendReceive(FrameOperationUploadFile, StateValidated, meta, meta.FileSize, bodyWriterHandler)
    if e != nil {
        return nil, e
    }
    var res = &UploadFileResponseMeta{}
    e1 := json.Unmarshal(frame.FrameMeta, res)
    if e1 != nil {
        return nil, e1
    }
    return res, nil
}


// pull files from tracker
func (client *TcpBridgeClient) QueryFile(meta *QueryFileMeta) (*QueryFileResponseMeta, error) {
    frame, e := client.sendReceive(FrameOperationQueryFile, StateValidated, meta, 0, nil)
    if e != nil {
        return nil, e
    }
    var res = &QueryFileResponseMeta{}
    e1 := json.Unmarshal(frame.FrameMeta, res)
    if e1 != nil {
        return nil, e1
    }
    return res, nil
}

// download file from storage server.
func (client *TcpBridgeClient) DownloadFile(meta *DownloadFileMeta) (*DownloadFileResponseMeta, *Frame, error) {
    frame, e := client.sendReceive(FrameOperationDownloadFile, StateValidated, meta, 0, nil)
    if e != nil {
        return nil, nil, e
    }
    var res = &DownloadFileResponseMeta{}
    e1 := json.Unmarshal(frame.FrameMeta, res)
    if e1 != nil {
        return nil, nil, e1
    }
    return res, frame, nil
}

// synchronize storage servers's statistic info from tracker server.
func (client *TcpBridgeClient) SyncStatistic(meta *SyncStatisticMeta) (*SyncStatisticResponseMeta, error) {
    frame, e := client.sendReceive(FrameOperationSyncStatistic, StateValidated, meta, 0, nil)
    if e != nil {
        return nil, e
    }
    var res = &SyncStatisticResponseMeta{}
    e1 := json.Unmarshal(frame.FrameMeta, res)
    if e1 != nil {
        return nil, e1
    }
    return res, nil
}



// send request and receive response,
// returns response frame and error.
func (client *TcpBridgeClient) sendReceive(operation byte,
                                        statusRequire int,
                                        meta interface{},
                                        bodyLength int64,
                                        bodyWriterHandler func(manager *ConnectionManager, frame *Frame) error,
                                        ) (*Frame, error) {
    client.connManager.RequireStatus(statusRequire)
    frame := &Frame{}
    frame.SetOperation(operation)
    frame.SetMeta(meta)
    frame.SetMetaBodyLength(bodyLength)
    frame.BodyWriterHandler = bodyWriterHandler
    if err := client.connManager.Send(frame); err != nil {
        return nil, err
    }
    response, e := client.connManager.Receive()
    if e != nil {
        return nil, e
    }
    if response != nil {
        return response, nil
    } else {
        return nil, errors.New("receive empty response from server")
    }
}




