package command

import (
	"encoding/json"
	"github.com/hetianyi/godfs/api"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/file"
	"github.com/hetianyi/gox/logger"
)

var client api.ClientAPI

func initClient() error {
	util.ValidateClientConfig(common.InitializedClientConfiguration)
	client = api.NewClient()
	// parse servers
	trackerServers, err := util.ParseServers(trackers)
	if err != nil {
		return err
	}
	_staticServer, err := util.ParseServers(storages)
	if err != nil {
		return err
	}
	staticServer := make([]*common.StorageServer, gox.TValue(_staticServer == nil , 0, len(_staticServer)).(int))
	if _staticServer != nil {
		for i, s := range _staticServer {
			staticServer[i] = &common.StorageServer{
				Server: *s,
			}
		}
	}
	// init client config
	client.SetConfig(&api.Config{
		MaxConnectionsPerServer: api.DefaultMaxConnectionsPerServer,
		StaticStorageServers: staticServer,
		TrackerServers: trackerServers,
	})
	return nil
}

func handleUploadFile() error {
	logger.Debug("")
	// initialize APIClient
	if err := initClient(); err != nil {
		logger.Fatal(err)
	}
	wd, err := file.GetWorkDir()
	if err != nil {
		return err
	}
	total := 0
	success := 0
	// upload all files in work dir.
	if uploadFiles.Len() == 1 && uploadFiles.Front().Value.(string) == "*" {
		files, err := file.ListFiles(wd)
		if err != nil {
			return err
		}
		if files != nil && len(files) > 0 {
			for _, inf := range files {
				if inf.IsDir() {
					continue
				}
				total++
				fi, err := file.GetFile(inf.Name())
				if err != nil {
					logger.Error(err)
				}
				ret, err := client.Upload(fi, inf.Size(), group)
				if err != nil {
					logger.Error(err)
				}
				success++
				bs, _ := json.MarshalIndent(ret, "", "  ")
				logger.Info("\n", string(bs))
			}
		}
	} else { // upload specified files.
		gox.WalkList(&uploadFiles, func(item interface{}) bool {
			total++
			fi, err := file.GetFile(item.(string))
			if err != nil {
				logger.Error(err)
				return false
			}
			inf, err := fi.Stat()
			if err != nil {
				logger.Error(err)
				return false
			}
			total++
			ret, err := client.Upload(fi, inf.Size(), group)
			if err != nil {
				logger.Error(err)
				return false
			}
			success++
			bs, _ := json.MarshalIndent(ret, "", "  ")
			logger.Info("upload success: ", item.(string), "\n", string(bs))
			return false
		})
	}
	return nil
}
