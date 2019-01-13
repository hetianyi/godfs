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

const (
    state_not_connect = 0
    state_connected = 1
    state_validated = 2
    state_disconnected = 3
)
var connPool *pool.ClientConnectionPool

type BridgeClient struct {
    // storage server info
    server *app.ServerInfo
    connManager *ConnectionManager
    // connect state
    // 0: not connect
    // 1: connected but not validate
    // 2: validated
    // 3: disconnected
    state int
}


func init() {
    connPool = &pool.ClientConnectionPool{}
    connPool.Init(50)
}


// create a new instance for bridgev2.Server
func NewClient(server *app.ServerInfo) *BridgeClient {
    return &BridgeClient {server, nil, 0}
}


// shutdown this client and close connection
func (client *BridgeClient) Close() {
    client.state = state_disconnected
    if client.connManager != nil {
        connPool.ReturnBrokenConnBridge(client.server, client.connManager.Conn)
    }
}


// shutdown this client not close connection
func (client *BridgeClient) Destroy() {
    client.state = state_disconnected
    if client.connManager != nil {
        connPool.ReturnConnBridge(client.server, client.connManager.Conn)
    }
}


// connect to server
func (client *BridgeClient) Connect() error {
    if client.state > state_not_connect {
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
    client.state = 1
    return nil
}


// validate this connection.
func (client *BridgeClient) Validate() (*ConnectResponseMeta, error) {
    meta := &ConnectMeta{
        Secret: app.SECRET,
        UUID: app.UUID,
    }
    frame, e := client.sendReceive(FRAME_OPERATION_VALIDATE, state_connected, meta, 0)
    if e != nil {
        return nil, e
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
    frame, e := client.sendReceive(FRAME_OPERATION_SYNC_STORAGE_MEMBERS, state_validated, storage, 0)
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
    frame, e := client.sendReceive(FRAME_OPERATION_REGISTER_FILES, state_validated, meta, 0)
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
    frame, e := client.sendReceive(FRAME_OPERATION_REGISTER_FILES, state_validated, meta, 0)
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
    client.requireStatus(statusRequire)
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

// assert status.
func (client *BridgeClient) requireStatus(requiredState int) error {
    if client.state < requiredState {
        panic(errors.New("connect state not satisfied, expect " + strconv.Itoa(requiredState) + ", now is " + strconv.Itoa(client.state)))
    }
    return nil
}






