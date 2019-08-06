package api_test

import (
	"fmt"
	"github.com/hetianyi/godfs/api"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox/file"
	"github.com/hetianyi/gox/logger"
	json "github.com/json-iterator/go"
	"io"
	"testing"
)

var client api.ClientAPI

func init() {
	logger.Init(&logger.Config{
		Level: logger.DebugLevel,
	})

	client = api.NewClient()
	client.SetConfig(&api.Config{
		MaxConnectionsPerServer: 1,
		SynchronizeOnce:         true,
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
}

func uploadFile() {
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

func TestUploadFile(t *testing.T) {
	for {
		uploadFile()
	}
}

func TestDownload1(t *testing.T) {
	fileId := "G01/4D/99/fde67f4752cf437ec6c831111127afaa"
	err := client.Download(fileId, 0, -1, func(body io.Reader, bodyLength int64) error {
		out, err := file.CreateFile("D:/tmp/godfs-test-download.zip")
		if err != nil {
			return err
		}
		defer out.Close()
		io.Copy(out, body)
		return nil
	})
	if err != nil && err != common.NotFoundErr {
		logger.Fatal(err)
	}
	logger.Info("download file success!")

	err = client.Download(fileId, 0, -1, func(body io.Reader, bodyLength int64) error {
		out, err := file.CreateFile("D:/tmp/godfs-test-download1.zip")
		if err != nil {
			return err
		}
		defer out.Close()
		io.Copy(out, body)
		return nil
	})
	if err != nil && err != common.NotFoundErr {
		logger.Fatal(err)
	}
	logger.Info("download file success!")
}

func TestDownload2(t *testing.T) {
	fileId := "G01/4D/99/fde67f4752cf437ec6c831111127afaa"
	err := client.Download(fileId, 0, -1, func(body io.Reader, bodyLength int64) error {
		out, err := file.CreateFile("D:/tmp/godfs-test-download.zip")
		if err != nil {
			return err
		}
		defer out.Close()
		io.Copy(out, body)
		return nil
	})
	if err != nil && err != common.NotFoundErr {
		logger.Fatal(err)
	}
	if err == common.NotFoundErr {
		logger.Error(err)
	} else {
		logger.Info("download file success!")
	}
}
