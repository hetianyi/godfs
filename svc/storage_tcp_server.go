package svc

import (
	"encoding/json"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/gpip"
	"github.com/hetianyi/gox/logger"
	"github.com/logrusorgru/aurora"
	"io"
	"net"
	"time"
)

func StartStorageTcpServer() {
	listener, err := net.Listen("tcp", common.Config.BindAddress+":"+convert.IntToStr(common.Config.Port))
	if err != nil {
		logger.Fatal(err)
	}
	time.Sleep(time.Millisecond * 50)
	logger.Info("  tcp server starting on port ", common.Config.Port)
	logger.Info(aurora.BrightGreen(":::server started:::"))
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("error accepting new connection: ", err)
			continue
		}
		logger.Debug("accept a new connection")
		gox.Try(func() {
			go clientConnHandler(conn)
		}, func(i interface{}) {
			logger.Error("connection error:", err)
		})
	}
}

func clientConnHandler(conn net.Conn) {
	pip := &gpip.Pip{
		Conn: conn,
	}
	defer pip.Close()
	for {
		err := pip.Receive(&common.Header{}, func(_header interface{}, bodyReader io.Reader, bodyLength int64) error {
			header := _header.(*common.Header)
			bs, _ := json.Marshal(header)
			logger.Debug("server got message:", string(bs))
			if header.Operation == common.OPERATION_CONNECT {
				return pip.Send(authenticationHandler(header), nil, 0)
			}
			return pip.Send(&common.Header{
				Result: common.SUCCESS,
				Msg:    "",
				Attributes: map[string]interface{}{"Name":"李四"},
			}, nil, 0)
		})
		if err != nil {
			logger.Error("error receive data:", err)
			break
		}
	}
}

func authenticationHandler(header *common.Header) *common.Header {
	if header.Attributes == nil {
		return &common.Header{
			Result: common.UNAUTHORIZED,
			Msg:    "authentication failed",
		}
	}
	secret := header.Attributes["secret"]
	if secret != common.Config.Secret {
		return &common.Header{
			Result: common.UNAUTHORIZED,
			Msg:    "authentication failed",
		}
	}
	return &common.Header{
		Result: common.SUCCESS,
		Msg: "authentication success",
	}
}
