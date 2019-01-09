package bridgev2

import (
    "libclient"
    "util/logger"
    "strconv"
    "crypto/md5"
    "errors"
    "app"
    "util/json"
)

var connPool *libclient.ClientConnectionPool

type BridgeClient struct {
    // storage server info
    member *app.Member
    connManager *ConnectionManager
    // connect state
    // 0: not connect
    // 1: connected but not validate
    // 2: validated
    // 3: disconnected
    state int
}

func init() {
    connPool = &libclient.ClientConnectionPool{}
    connPool.Init(50)
}

// create a new instance for bridgev2.Server
func NewClient(member *app.Member) (*BridgeClient, error) {
    client := &BridgeClient {member, nil, 0}
    return client, nil
}


func (client *BridgeClient) Close() {
    if client.connManager != nil {
        connPool.ReturnBrokenConnBridge(client.member, client.connManager.conn)
    }
}


func (client *BridgeClient) Connect() error {
    conn, err := connPool.GetConn(client.member)
    if err != nil {
        return err
    }
    h, p := client.member.GetHostAndPortByAccessFlag()
    logger.Debug("connect to", h + ":" + strconv.Itoa(p), "success")
    client.connManager = &ConnectionManager{
        conn: conn,
        side: CLIENT_SIDE,
        md: md5.New(),
    }
    client.state = 1
    return nil
}


func (client *BridgeClient) Validate() (*ConnectResponseMeta, error) {
    client.requireStatus(1)
    meta := &ConnectMeta{
        Secret: app.SECRET,
        UUID: app.UUID,
    }

    frame := &Frame{}
    frame.SetOperation(FRAME_OPERATION_VALIDATE)
    frame.SetMeta(meta)
    if err := client.connManager.Send(frame); err != nil {
        return nil, err
    }
    response, e := client.connManager.Receive()
    if e != nil {
        return nil, e
    }
    if response != nil {
        var res = &ConnectResponseMeta{}
        e := json.Unmarshal(response.frameMeta, res)
        if e != nil {
            return nil, e
        }
        return res, nil
    } else {
        return nil, errors.New("receive empty response from server")
    }
}

func (client *BridgeClient) requireStatus(requiredState int) error {
    if client.state < requiredState {
        panic(errors.New("connect state not satisfied, expect " + strconv.Itoa(requiredState) + ", now is " + strconv.Itoa(client.state)))
    }
    return nil
}






