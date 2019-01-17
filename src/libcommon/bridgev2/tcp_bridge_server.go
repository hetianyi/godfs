package bridgev2

import (
	"strconv"
	"util/logger"
	"errors"
	"net"
	"crypto/md5"
	"util/common"
	"util/pool"
)

// max client connection set to 1000
var connectionPool, _ = pool.NewPool(1000, 0)

type TcpBridgeServer struct {
	Host string
	Port int
}

// create a new instance for bridgev2.Server
func NewServer(host string, port int) *TcpBridgeServer {
	server := &TcpBridgeServer {
		Host: host,
		Port: port,
	}
	return server
}


// server start listening.
// callback func will called when a server connection is closed by server/client.
func (server *TcpBridgeServer) Listen(callbacks... func(manager *ConnectionManager)) error {

	if server.Port <= 0 || server.Port > 65535 {
		return errors.New("invalid port range: " + strconv.Itoa(server.Port))
	}

	listener, e1 := net.Listen("tcp", ":"+strconv.Itoa(server.Port))
	if e1 != nil {
		panic(e1)
		return nil
	}
	logger.Info("server listening on port:", server.Port)
	// keep accept connections.
	for {
		conn, e1 := listener.Accept()
		manager := &ConnectionManager{
			Conn: conn,
			Side: SERVER_SIDE,
		}
		if e1 != nil {
			logger.Error("accept new conn error:", e1)
			manager.Destroy()
		} else {
			logger.Debug("accept a new connection from remote addr:", conn.RemoteAddr().String())
			connectionPool.Exec(func() {
				manager.State = STATE_CONNECTED
				common.Try(func() {
					Serve(manager, callbacks...)
				}, func(i interface{}) {
				})
			})
		}
	}
	return nil
}


// server socket serve a single connection
func Serve(manager *ConnectionManager, callbacks... func(manager *ConnectionManager)) {
	if manager.Md == nil {
		manager.Md = md5.New()
	}
	side := ""
	if manager.Side == SERVER_SIDE {
		side = "server"
	} else {
		side = "client"
	}
	defer func() {
		logger.Debug("close connection from", side)
		// call callback functions
		if callbacks != nil {
			for i := range callbacks {
				fun := callbacks[i]
				common.Try(func() {
					fun(manager)
				}, func(i interface{}) {
					logger.Error(i)
				})
			}
		}
		manager.Destroy()
	}()
	common.Try(func() {
		logger.Debug("ready for client request event")
		for {
			frame, err := manager.Receive()
			if err != nil {
				panic(err)
				break
			}
			handler := GetOperationHandler(frame.GetOperation())
			if handler == nil || handler.Handler == nil {
				panic(errors.New("no handler for operation: " + strconv.Itoa(int(frame.GetOperation()))))
				break
			}
			logger.Debug("receive a new request from remote client, operation:", frame.GetOperation())
			if e2 := handler.Handler(manager, frame) ; e2 != nil {
				panic(e2)
				break
			}
		}
	}, func(i interface{}) {
		logger.Error("server serve error:", i)
	})
}








