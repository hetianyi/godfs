package svc

import (
	"encoding/json"
	"github.com/hetianyi/godfs/api"
	"github.com/hetianyi/godfs/binlog"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/reg"
	"io"
)

const MaxConnPerServer uint = 100

var (
	clientAPI             api.ClientAPI
	writableBinlogManager binlog.XBinlogManager
)

// InitializeClientAPI initializes client API.
func InitializeClientAPI(config *api.Config) {
	clientAPI = api.NewClient()
	clientAPI.SetConfig(config)
}

func authenticationHandler(header *common.Header, secret string) (*common.Header, *common.Instance, io.Reader, int64, error) {
	if header.Attributes == nil {
		return &common.Header{
			Result: common.UNAUTHORIZED,
			Msg:    "authentication failed",
		}, nil, nil, 0, nil
	}
	s := header.Attributes["secret"]
	if s != secret {
		return &common.Header{
			Result: common.UNAUTHORIZED,
			Msg:    "authentication failed",
		}, nil, nil, 0, nil
	}

	var instance *common.Instance
	if common.BootAs == common.BOOT_TRACKER {
		// parse instance info.
		s1 := header.Attributes["instance"]
		instance = &common.Instance{}
		if err := json.Unmarshal([]byte(s1), instance); err != nil {
			return &common.Header{
				Result: common.ERROR,
				Msg:    err.Error(),
			}, nil, nil, 0, err
		}
		if err := reg.Put(instance); err != nil {
			return &common.Header{
				Result: common.ERROR,
				Msg:    err.Error(),
			}, nil, nil, 0, err
		}
	}

	return &common.Header{
		Result: common.SUCCESS,
		Msg:    "authentication success",
	}, instance, nil, 0, nil
}
