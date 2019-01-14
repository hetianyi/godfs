package bridgev2

import (
    "app"
    "crypto/md5"
    "errors"
    "strconv"
    "util/json"
    "util/logger"
    "util/pool"
)

var connPool *pool.ClientConnectionPool

type BridgeClient struct {
    // storage server info
    server *app.ServerInfo
    connManager *ConnectionManager
}


func init() {
    connPool = &pool.ClientConnectionPool{}
    connPool.Init(50)
}


// create a new instance for bridgev2.Server
func NewClient(server *app.ServerInfo) *BridgeClient {
    return &BridgeClient {server, nil}
}


// shutdown this client and close connection
func (client *BridgeClient) Close() {
    client.connManager.state = STATE_DISCONNECTED
    if client.connManager != nil {
        connPool.ReturnBrokenConnBridge(client.server, client.connManager.Conn)
    }
}


// shutdown this client not close connection
func (client *BridgeClient) Destroy() {
    client.connManager.state = STATE_DISCONNECTED
    if client.connManager != nil {
        connPool.ReturnConnBridge(client.server, client.connManager.Conn)
    }
}


// connect to server
func (client *BridgeClient) Connect() error {
    if client.connManager.state > STATE_NOT_CONNECT {
        panic(errors.New("already connected"))
    }
    conn, err := connPool.GetConn(client.server)
    if err != nil {
        return err
    }
    h, p := client.server.GetHostAndPortByAccessFlag()
    logger.Debug("connect to", h + ":" + strconv.Itoa(p), "success")
    client.connManager = &ConnectionManager{
        Conn: conn,
        Side: CLIENT_SIDE,
        Md: md5.New(),
    }
    client.connManager.state = STATE_CONNECTED
    return nil
}


// validate this connection.
func (client *BridgeClient) Validate() (*ConnectResponseMeta, error) {
    meta := &ConnectMeta{
        Secret: app.SECRET,
        UUID: app.UUID,
    }
    frame, e := client.sendReceive(FRAME_OPERATION_VALIDATE, STATE_CONNECTED, meta, 0)
    if e != nil {
        return nil, e
    }
    if frame.GetStatus() == STATUS_SUCCESS {
        client.connManager.state = STATE_VALIDATED
    }
    var res = &ConnectResponseMeta{}
    e1 := json.Unmarshal(frame.FrameMeta, res)
    if e1 != nil {
        return nil, e1
    }
    return res, nil
}


// synchronized storage members.
func (client *BridgeClient) SyncStorageMembers(storage *app.StorageDO) (*SyncStorageMembersResponseMeta, error) {
    frame, e := client.sendReceive(FRAME_OPERATION_SYNC_STORAGE_MEMBERS, STATE_VALIDATED, storage, 0)
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

// register files to tracker
func (client *BridgeClient) RegisterFiles(meta *RegisterFileMeta) (*RegisterFileResponseMeta, error) {
    frame, e := client.sendReceive(FRAME_OPERATION_REGISTER_FILES, STATE_VALIDATED, meta, 0)
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
func (client *BridgeClient) PullFiles(meta *PullFileMeta) (*PullFileResponseMeta, error) {
    frame, e := client.sendReceive(FRAME_OPERATION_REGISTER_FILES, STATE_VALIDATED, meta, 0)
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


// send request and receive response,
// returns response frame and error.
func (client *BridgeClient) sendReceive(operation byte,
                                        statusRequire int,
                                        meta interface{},
                                        bodyLength int64,
                                        ) (*Frame, error) {
    client.connManager.RequireStatus(statusRequire)
    frame := &Frame{}
    frame.SetOperation(operation)
    frame.SetMeta(meta)
    frame.SetMetaBodyLength(bodyLength)
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




