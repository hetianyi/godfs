package bridgev2

import (
    "errors"
    "net"
    "strconv"
    "util/pool"
    "hash"
)

const (
    SERVER_SIDE = 1
    CLIENT_SIDE = 2

    STATE_NOT_CONNECT = 0
    STATE_CONNECTED = 1
    STATE_VALIDATED = 2
    STATE_DISCONNECTED = 3
)
// max client connection set to 1000
var connectionPool, _ = pool.NewPool(1000, 0)

// common connection manager
type ConnectionManager struct {
    Conn net.Conn // connection that being managed
    // represent this connection is server side(1) or client side(2)
    Side int
    Md hash.Hash
    // connect state
    // 0: not connect
    // 1: connected but not validate
    // 2: validated
    // 3: disconnected
    state int
    UUID string // storage uuid, this field is used by server side.
}


func (manager *ConnectionManager) Close() error {
    if manager.Conn != nil {
        return manager.Conn.Close()
    }
    return nil
}

// receive data frame from server/client
func (manager *ConnectionManager) Receive() (*Frame, error) {
    return readFrame(manager)
}

// send data to from server/client
func (manager *ConnectionManager) Send(frame *Frame) error {
    return writeFrame(manager, frame)
}


// assert status.
func (manager *ConnectionManager) RequireStatus(requiredState int) error {
    if manager.state < requiredState {
        panic(errors.New("connect state not satisfied, expect " + strconv.Itoa(requiredState) + ", now is " + strconv.Itoa(manager.state)))
    }
    return nil
}

