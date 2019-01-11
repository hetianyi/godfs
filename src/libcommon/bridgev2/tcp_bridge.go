package bridgev2

import (
    "net"
    "util/pool"
    "hash"
)

const (
    SERVER_SIDE = 1
    CLIENT_SIDE = 2
)

// max client connection set to 1000
var connectionPool, _ = pool.NewPool(1000, 0)

// common connection manager
type ConnectionManager struct {
    Conn net.Conn // connection that being managed
    // represent this connection is server side(1) or client side(2)
    Side int
    Md hash.Hash
}


func (manager *ConnectionManager) Close() error {
    if manager.Conn != nil {
        return manager.Conn.Close()
    }
    return nil
}

// receive data frame from server/client
func (manager *ConnectionManager) Receive() (*Frame, error) {
    return ReadFrame(manager)
}

// send data to from server/client
func (manager *ConnectionManager) Send(frame *Frame) error {
    return WriteFrame(manager, frame)
}



