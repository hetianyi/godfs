package api_test

import (
	"encoding/json"
	"fmt"
	"github.com/hetianyi/godfs/api"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox/file"
	"github.com/hetianyi/gox/logger"
	"testing"
)

func init() {
	logger.Init(&logger.Config{
		Level: logger.DebugLevel,
	})
}

func TestClientAPIImpl_Upload(t *testing.T) {
	client := api.NewClient()
	client.Init(&api.Config{
		MaxConnectionsPerServer: 1,
		StaticStorageServers: []*common.StorageServer{
			{
				Server: common.Server{
					Host:       "127.0.0.1",
					Port:       3456,
					Secret:     "kasd3123",
					InstanceId: "123",
				},
				Group: "G1",
			},
		},
	})

	for {
		fi, _ := file.GetFile("D:/tmp/js.zip")
		info, _ := fi.Stat()
		ret, err := client.Upload(fi, info.Size(), "")
		if err != nil {
			logger.Fatal(err)
		}
		bs, _ := json.MarshalIndent(ret, "", "  ")
		logger.Info("result is \n")
		fmt.Println(string(bs))
	}
}
