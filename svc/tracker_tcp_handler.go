package svc

import (
	"errors"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/gpip"
	"github.com/hetianyi/gox/logger"
	json "github.com/json-iterator/go"
	"github.com/logrusorgru/aurora"
	"io"
	"net"
	"time"
)

func StartTrackerTcpServer() {
	listener, err := net.Listen("tcp", common.InitializedTrackerConfiguration.BindAddress+":"+convert.IntToStr(common.InitializedTrackerConfiguration.Port))
	if err != nil {
		logger.Fatal(err)
	}
	time.Sleep(time.Millisecond * 50)
	logger.Info(" tcp server listening on ", common.InitializedTrackerConfiguration.BindAddress, ":", common.InitializedTrackerConfiguration.Port)
	logger.Info(aurora.BrightGreen("::: tracker server started :::"))
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("error accepting new connection: ", err)
			continue
		}
		logger.Debug("accept a new connection")
		go trackerClientConnHandler(conn)
	}
}

func trackerClientConnHandler(conn net.Conn) {
	pip := &gpip.Pip{
		Conn: conn,
	}
	defer pip.Close()
	validated := false
	for {
		err := pip.Receive(&common.Header{}, func(_header interface{}, bodyReader io.Reader, bodyLength int64) error {
			if _header == nil {
				return errors.New("invalid request: header is empty")
			}
			header := _header.(*common.Header)
			bs, _ := json.Marshal(header)
			logger.Debug("server got message:", string(bs))
			if header.Operation == common.OPERATION_CONNECT {
				h, b, l, err := authenticationHandler(header, common.InitializedStorageConfiguration.Secret)
				if err != nil {
					return err
				}
				if h.Result != common.SUCCESS {
					pip.Send(h, b, l)
					return errors.New("unauthorized connection, force disconnection by server")
				} else {
					validated = true
					return pip.Send(h, b, l)
				}
			} else if header.Operation == common.OPERATION_UPLOAD {
				h, b, l, err := uploadFileHandler(bodyReader, bodyLength, validated)
				if err != nil {
					return err
				}
				return pip.Send(h, b, l)
			} else if header.Operation == common.OPERATION_DOWNLOAD {
				h, b, l, err := downFileHandler(header, validated)
				if err != nil {
					return err
				}
				return pip.Send(h, b, l)
			} else if header.Operation == common.OPERATION_QUERY {
				h, b, l, err := inspectFileHandler(header, validated)
				if err != nil {
					return err
				}
				return pip.Send(h, b, l)
			}
			return pip.Send(&common.Header{
				Result:     common.SUCCESS,
				Msg:        "",
				Attributes: map[string]string{"Name": "李四"},
			}, nil, 0)
		})
		if err != nil {
			// shutdown connection error is now disabled
			/*if err != io.EOF {
				logger.Error(err)
			}*/
			pip.Close()
			break
		}
	}
}
