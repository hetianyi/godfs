package bridgev2

import (
	"util/common"
	"net"
	"strconv"
	"util/logger"
	"libcommon/bridge"
	"time"
)

type Server struct {
	Host string
	Port int
	Handler map[byte]OperationHandler
}

func (server *Server) Listen() {

	if server.Port <= 0 || server.Port > 65535 {
		logger.Error("invalid port range:", server.Port)
		return
	}

	for {
		common.Try(func() {
			listener, e := net.Listen("tcp", ":"+strconv.Itoa(server.Port))
			if e != nil {
				panic(e)
			} else {
				logger.Info("server listening on port:", server.Port)
				// keep accept connections.
				for {
					conn, e1 := listener.Accept()
					if e1 == nil {

					} else {
						logger.Info("accept new conn error", e1)
						if conn != nil {
							bridge.Close(conn)
						}
					}
				}
			}
		}, func(i interface{}) {
			time.Sleep(time.Second * 10)
		})
	}
}










