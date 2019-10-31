package api_test

import (
	"fmt"
	"github.com/hetianyi/godfs/api"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox/file"
	"github.com/hetianyi/gox/logger"
	json "github.com/json-iterator/go"
	"testing"
)

var client1 api.ClientAPI

func InitTrackerTest() {
	logger.Init(&logger.Config{
		Level: logger.InfoLevel,
	})
	readyChan := make(chan int)
	client1 = api.NewClient()
	client1.SetConfig(&api.Config{
		MaxConnectionsPerServer: 1,
		SynchronizeOnce:         true,
		SynchronizeOnceCallback: readyChan,
		TrackerServers: []*common.Server{
			{
				Host:       "127.0.0.1",
				Port:       6579,
				Secret:     "123456",
				InstanceId: "123456555",
			},
			{
				Host:       "127.0.0.1",
				Port:       6575,
				Secret:     "123456",
				InstanceId: "123456555",
			},
		},
	})

	logger.Info("waiting tracker servers...")
	total := 2
	stat := 0
	for i := 0; i < total; i++ {
		stat += <-readyChan
	}
	logger.Info("all tracker servers synced, errors: ", stat, "  of total: ", total)

}

func uploadFile1() {
	fi, _ := file.GetFile("D:/tmp/123.json")
	info, _ := fi.Stat()
	ret, err := client1.Upload(fi, info.Size(), "", false)
	if err != nil {
		logger.Fatal(err)
	}
	bs, _ := json.MarshalIndent(ret, "", "  ")
	logger.Info("result is \n")
	fmt.Println(string(bs))
}

func TestUploadFile1(t *testing.T) {
	InitTrackerTest()
	for {
		uploadFile1()
	}
}
