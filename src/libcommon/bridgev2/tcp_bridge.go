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
    conn net.Conn // connection that being managed
    // represent this connection is server side(1) or client side(2)
    side int
    md hash.Hash
}


func (manager *ConnectionManager) Close() error {
    if manager.conn != nil {
        return manager.conn.Close()
    }
    return nil
}


func (manager *ConnectionManager) Receive() (*Frame, error) {
    return readFrame(manager)
}

func (manager *ConnectionManager) Send(frame *Frame) error {
    return writeFrame(manager, frame)
}



