package svc_test

import (
	"encoding/json"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox/gpip"
	"github.com/hetianyi/gox/logger"
	"io"
	"net"
	"strconv"
	"testing"
)

func TestSendMsg(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(6577))
	if err != nil {
		logger.Fatal("error start client:", err)
	}
	pip := &gpip.Pip{
		Conn: conn,
	}
	err = pip.Send(&common.Header{
		Operation: common.OPERATION_CONNECT,
		Attributes: map[string]interface{}{"secret":"123456"},
	}, nil, 0)
	if err != nil {
		logger.Fatal("error send data:", err)
	}
	err = pip.Receive(&common.Header{}, func(_header interface{}, bodyReader io.Reader, bodyLength int64) error {
		header := _header.(*common.Header)
		bs, _ := json.Marshal(header)
		logger.Info("client got message:", string(bs))
		return nil
	})
	if err != nil {
		logger.Error("error:", err)
	}

}
