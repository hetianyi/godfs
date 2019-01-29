package bridgev2

import (
	"app"
	"errors"
	"hash"
	"net"
	"strconv"
	"util/pool"
)

const (
	ServerSide = 1
	ClientSide = 2

	StateNotConnect   = 0
	StateConnected    = 1
	StateValidated    = 2
	StateDisconnected = 3
)

var connPool *pool.ClientConnectionPool

func init() {
	connPool = &pool.ClientConnectionPool{}
	connPool.Init(50)
}

// ConnectionManager common connection manager
// server connection dont't has server info
type ConnectionManager struct {
	// storage server info
	server *app.ServerInfo
	Conn   net.Conn // connection that being managed
	// represent this connection is server side(1) or client side(2)
	Side int
	Md   hash.Hash
	// connect state
	// 0: not connect
	// 1: connected but not validate
	// 2: validated
	// 3: disconnected
	State int
	UUID  string // storage uuid, this field is used by server side.
}

// Close close manager and return connection to pool.
func (manager *ConnectionManager) Close() {
	if manager.Conn != nil {
		connPool.ReturnConnBridge(manager.server, manager.Conn)
	}
}

// Destroy close manager and close connection.
func (manager *ConnectionManager) Destroy() {
	if manager.server == nil {
		if manager.Conn != nil {
			manager.Conn.Close()
		}
		return
	}
	if manager.Conn != nil {
		connPool.ReturnBrokenConnBridge(manager.server, manager.Conn)
	}
}

// Receive receive data frame from server/client
func (manager *ConnectionManager) Receive() (*Frame, error) {
	return readFrame(manager)
}

// Send send data to from server/client
func (manager *ConnectionManager) Send(frame *Frame) error {
	return writeFrame(manager, frame)
}

// RequireStatus assert status.
func (manager *ConnectionManager) RequireStatus(requiredState int) error {
	if manager.State < requiredState {
		panic(errors.New("connect state not satisfied, expect " + strconv.Itoa(requiredState) + ", now is " + strconv.Itoa(manager.State)))
	}
	return nil
}
