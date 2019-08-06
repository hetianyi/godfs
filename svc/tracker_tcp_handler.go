package svc

import (
	"errors"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/reg"
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
	authorized := false
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
					authorized = true
					return pip.Send(h, b, l)
				}
			}
			if !authorized {
				pip.Send(&common.Header{
					Result: common.UNAUTHORIZED,
					Msg:    "authentication failed",
				}, nil, 0)
				return errors.New("unauthorized connection, force disconnection by server")
			}
			if header.Operation == common.OPERATION_REGISTER {
				h, b, l, err := registerHandler(header)
				if err != nil {
					return err
				}
				return pip.Send(h, b, l)
			} else if header.Operation == common.OPERATION_SYNC_INSTANCES {
				h, b, l, err := synchronizeInstancesHandler(header)
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
			pip.Close()
			break
		}
	}
}

// inspectFileHandler inspects file's information
func registerHandler(header *common.Header) (*common.Header, io.Reader, int64, error) {
	if header.Attributes == nil {
		return &common.Header{
			Result: common.ERROR,
			Msg:    "no message provided",
		}, nil, 0, nil
	}
	s1 := header.Attributes["instance"]
	instance := &common.Instance{}
	if err := json.Unmarshal([]byte(s1), instance); err != nil {
		return &common.Header{
			Result: common.ERROR,
			Msg:    err.Error(),
		}, nil, 0, err
	}
	if err := reg.Put(instance); err != nil {
		return &common.Header{
			Result: common.ERROR,
			Msg:    err.Error(),
		}, nil, 0, err
	}
	return &common.Header{
		Result: common.SUCCESS,
	}, nil, 0, nil
}

// inspectFileHandler inspects file's information
func synchronizeInstancesHandler(header *common.Header) (*common.Header, io.Reader, int64, error) {
	snapshot := reg.InstanceSetSnapshot()
	ret, _ := json.Marshal(snapshot)
	return &common.Header{
		Result: common.SUCCESS,
		Attributes: map[string]string{
			"instances": string(ret),
		},
	}, nil, 0, nil
}
