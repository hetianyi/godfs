package svc

import (
	"encoding/json"
	"errors"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/file"
	"github.com/hetianyi/gox/gpip"
	"github.com/hetianyi/gox/logger"
	"github.com/hetianyi/gox/uuid"
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
	validated := false
	for {
		err := pip.Receive(&common.Header{}, func(_header interface{}, bodyReader io.Reader, bodyLength int64) error {
			header := _header.(*common.Header)
			bs, _ := json.Marshal(header)
			logger.Debug("server got message:", string(bs))
			if header.Operation == common.OPERATION_CONNECT {
				h, b, l, err := authenticationHandler(header)
				if err != nil {
					return err
				}
				if h.Result != common.UNAUTHORIZED {
					return errors.New("unauthorized connection, force disconnection by server")
				}
				validated = true
				return pip.Send(h, b, l)
			} else if header.Operation == common.OPERATION_UPLOAD {
				h, b, l, err := uploadFileHandler(header)
				if err != nil {
					return err
				}
				if h.Result != common.UNAUTHORIZED {
					return errors.New("unauthorized connection, force disconnection by server")
				}
				validated = true
				return pip.Send(h, b, l)
			}
			return pip.Send(&common.Header{
				Result: common.SUCCESS,
				Msg:    "",
				Attributes: map[string]interface{}{"Name":"李四"},
			}, nil, 0)
		})
		if err != nil {
			logger.Error("error receive data:", err)
			pip.Close()
			break
		}
	}
}

func authenticationHandler(header *common.Header,
	bodyReader io.Reader,
	bodyLength int64,
	authorized bool) (*common.Header, io.Reader, int64, error) {
	if header.Attributes == nil {
		return &common.Header{
			Result: common.UNAUTHORIZED,
			Msg:    "authentication failed",
		}, nil, 0, nil
	}
	secret := header.Attributes["secret"]
	if secret != common.Config.Secret {
		return &common.Header{
			Result: common.UNAUTHORIZED,
			Msg:    "authentication failed",
		}, nil, 0, nil
	}
	return &common.Header{
		Result: common.SUCCESS,
		Msg: "authentication success",
	}, nil, 0, nil
}


func uploadFileHandler(header *common.Header,
	bodyReader io.Reader,
	bodyLength int64,
	authorized bool) (*common.Header, io.Reader, int64, error) {
	if !authorized {
		return nil, nil, 0, errors.New("unauthorized connection")
	}
	buffer := make([]byte, common.BUFFER_SIZE)
	var realRead int64 = 0
	h := util.CreateCrc32Hash()
	tmpFile := common.Config.TmpDir + "/" + uuid.UUID()
	out, err := file.CreateFile(tmpFile)
	if err != nil {
		return nil, nil, 0, err
	}
	defer func() {
		defer out.Close()
		file.Delete(tmpFile)
	}()
	for true {
		n, err := bodyReader.Read(buffer)
		realRead += int64(n)
		if err != nil && err != io.EOF {
			return nil, nil, 0, err
		}
		if n > 0 {
			_, err := h.Write(buffer[0:n])
			if err != nil {
				return nil, nil, 0, err
			}
			_, err = out.Write(buffer[0:n])
			if err != nil {
				return nil, nil, 0, err
			}
		} else {
			out.Close()
			// TODO
			return &common.Header{}, nil, 0, nil
		}
	}

	return &common.Header{}, nil, 0, nil
}


