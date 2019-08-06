package command

import (
	"github.com/hetianyi/godfs/api"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/file"
	"github.com/hetianyi/gox/logger"
	"github.com/hetianyi/gox/pg"
	json "github.com/json-iterator/go"
	"io"
	"path/filepath"
)

var client api.ClientAPI

// initClient initializes APIClient.
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
	staticServer := make([]*common.StorageServer, gox.TValue(_staticServer == nil, 0, len(_staticServer)).(int))
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
		SynchronizeOnce:         true,
		StaticStorageServers:    staticServer,
		TrackerServers:          trackerServers,
	})
	return nil
}

// handleUploadFile handles upload files by client cli.
func handleUploadFile() error {
	// initialize APIClient
	if err := initClient(); err != nil {
		logger.Fatal(err)
	}
	wd, err := file.GetWorkDir()
	if err != nil {
		return err
	}
	total := 0   // total files
	success := 0 // success files
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
				r := &pg.WrappedReader{Reader: fi}
				// show upload progressbar.
				pro := pg.NewWrappedReaderProgress(inf.Size(), 50, "uploading "+inf.Name(), pg.Top, r)
				ret, err := client.Upload(r, inf.Size(), group)
				fi.Close()
				if err != nil {
					pro.Destroy()
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
			r := &pg.WrappedReader{Reader: fi}
			// show upload progressbar.
			pro := pg.NewWrappedReaderProgress(inf.Size(), 50, "uploading "+inf.Name(), pg.Top, r)
			ret, err := client.Upload(r, inf.Size(), group)
			if err != nil {
				pro.Destroy()
				logger.Error(err)
				return false
			}
			success++
			bs, _ := json.MarshalIndent(ret, "", "  ")
			logger.Info("upload success: ", item.(string), "\n", string(bs))
			return false
		})
	}
	logger.Info("upload finish, success ", success, " of total ", total)
	return nil
}

// handleDownloadFile handles download files by client cli.
func handleDownloadFile() error {
	// initialize APIClient
	if err := initClient(); err != nil {
		logger.Fatal(err)
	}
	wd, err := file.GetWorkDir()
	if err != nil {
		return err
	}
	if downloadFiles.Len() == 0 {
		return nil
	}
	// generate full file path of custom download file name.
	if customDownloadFileName != "" && !file.IsAbsPath(customDownloadFileName) {
		absPath, err := file.AbsPath(customDownloadFileName)
		if err == nil {
			customDownloadFileName = absPath
		} else {
			customDownloadFileName = ""
		}
	}
	// create directory for download files.
	if downloadFiles.Len() == 1 && customDownloadFileName != "" {
		gox.Try(func() {
			if !file.Exists(customDownloadFileName) || (file.Exists(customDownloadFileName) && file.IsFile1(customDownloadFileName)) {
				parent := filepath.Dir(customDownloadFileName)
				if !file.Exists(parent) {
					err := file.CreateDirs(parent)
					if err != nil {
						panic(err)
					}
				}
			} else {
				customDownloadFileName = ""
			}
		}, func(e interface{}) {
			logger.Error("cannot custom download filename: ", e)
			customDownloadFileName = ""
		})
	}
	total := 0   // total files
	success := 0 // success files
	// checking download fileIds.
	gox.WalkList(&downloadFiles, func(item interface{}) bool {
		total++
		if !common.FileIdPatternRegexp.Match([]byte(item.(string))) {
			logger.Warn("invalid format fileId: ", item.(string))
			return false
		}
		err := client.Download(item.(string), 0, -1, func(body io.Reader, bodyLength int64) error {
			// if download only one file and provide a custom filename.
			if downloadFiles.Len() == 1 && customDownloadFileName != "" {
				logger.Info("downloading ", item.(string), " to ", customDownloadFileName)
				fi, err := file.CreateFile(customDownloadFileName)
				w := &pg.WrappedWriter{Writer: fi}
				// show download progressbar.
				pro := pg.NewWrappedWriterProgress(bodyLength, 50, "downloading "+item.(string), pg.Top, w)
				_, err = io.Copy(w, body)
				if err != nil {
					pro.Destroy()
				}
				return err
			}
			md5 := common.FileIdPatternRegexp.ReplaceAllString(item.(string), "$4")
			fi, err := file.CreateFile(wd + "/" + md5)
			logger.Info("downloading ", item.(string), " to ", fi.Name())
			if err != nil {
				return err
			}
			w := &pg.WrappedWriter{Writer: fi}
			// show download progressbar.
			pro := pg.NewWrappedWriterProgress(bodyLength, 50, "downloading "+item.(string), pg.Top, w)
			_, err = io.Copy(w, body)
			if err != nil {
				pro.Destroy()
			}
			return err
		})
		if err == nil {
			success++
		} else {
			logger.Error("error downloading file ", item.(string), ": ", err)
		}
		return false
	})
	logger.Info("download finish, success ", success, " of total ", total)
	return nil
}

// handleInspectFile handles query file information by client cli.
func handleInspectFile() error {
	// initialize APIClient
	if err := initClient(); err != nil {
		logger.Fatal(err)
	}
	if inspectFiles.Len() == 0 {
		return nil
	}
	resultMap := make(map[string]*common.FileInfo)
	total := 0
	success := 0
	// checking download fileIds.
	gox.WalkList(&inspectFiles, func(item interface{}) bool {
		total++
		if !common.FileIdPatternRegexp.Match([]byte(item.(string))) {
			logger.Warn("invalid format fileId: ", item.(string))
			return false
		}

		info, err := client.Query(item.(string))
		resultMap[item.(string)] = info
		if err == nil {
			success++
		} else {
			logger.Error("error inspect file ", item.(string), ": ", err)
		}
		return false
	})
	bs, err := json.MarshalIndent(resultMap, "", "  ")
	if err != nil {
		logger.Error(err)
	} else {
		logger.Info("inspect result:\n", string(bs))
	}
	logger.Info("inspect finish, success ", success, " of total ", total)
	return nil
}
