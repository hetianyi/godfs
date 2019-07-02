package reg_test

import (
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/reg"
	"github.com/hetianyi/gox/logger"
	"testing"
)

func init() {
	logger.Init(&logger.Config{
		Level: logger.DebugLevel,
	})
}

func hold() {
	h := make(chan int)
	<-h
}

func TestPut(t *testing.T) {
	s1 := &common.Instance{
		Server: common.Server{
			Host:       "192.168.1.142",
			Port:       9012,
			Secret:     "123456",
			InstanceId: "xxxxxx",
		},
		Role: common.ROLE_STORAGE,
	}
	s2 := &common.Instance{
		Server: common.Server{
			Host:       "192.168.1.143",
			Port:       9012,
			Secret:     "123456",
			InstanceId: "xxxxxx",
		},
		Role: common.ROLE_STORAGE,
	}

	logger.Error(reg.Put(s1))
	logger.Error(reg.Put(s2))
	reg.Free("xxxxxx")

	hold()
}
