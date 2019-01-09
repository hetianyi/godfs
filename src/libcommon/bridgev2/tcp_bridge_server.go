package bridgev2

import (
	"strconv"
	"util/logger"
	"errors"
	"net"
	"crypto/md5"
	"util/common"
)


type BridgeServer struct {
	Host string
	Port int
}

// create a new instance for bridgev2.Server
func NewServer(host string, port int) (*BridgeServer, error) {
	server := &BridgeServer {
		Host: host,
		Port: port,
	}
	return server, nil
}


// server start listening
func (server *BridgeServer) Listen() error {

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
			conn: conn,
			side: SERVER_SIDE,
		}
		if e1 != nil {
			logger.Error("accept new conn error:", e1)
			manager.Close()
		} else {
			connectionPool.Exec(func() {
				Serve(manager)
			})
		}
	}
	return nil
}


// server socket serve a single connection
func Serve(manager *ConnectionManager) {
	if manager.md == nil {
		manager.md = md5.New()
	}
	side := ""
	if manager.side == SERVER_SIDE {
		side = "server"
	} else {
		side = "client"
	}
	defer func() {
		logger.Debug("close connection from", side)
		manager.Close()
	}()
	common.Try(func() {

		for ;; {
			frame, err := manager.Receive()
			if err != nil {
				break
			}

		}


	}, func(i interface{}) {
		logger.Error("connection error:", i)
	})
}








