package command

import (
	"bytes"
	"fmt"
	"github.com/hetianyi/godfs/api"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/file"
	"github.com/hetianyi/gox/logger"
	"github.com/hetianyi/gox/pg"
	json "github.com/json-iterator/go"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
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

	var readyChan chan int
	if trackerServers != nil && len(trackerServers) >= 0 {
		readyChan = make(chan int)
	}
	// init client config
	client.SetConfig(&api.Config{
		MaxConnectionsPerServer: api.DefaultMaxConnectionsPerServer,
		SynchronizeOnce:         true,
		SynchronizeOnceCallback: readyChan,
		StaticStorageServers:    staticServer,
		TrackerServers:          trackerServers,
	})

	if readyChan != nil {
		logger.Debug("synchronizing with tracker servers...")
		total := len(trackerServers)
		stat := 0
		for i := 0; i < total; i++ {
			stat += <-readyChan
		}
		logger.Debug("synchronized with all tracker servers, errors: ", stat, "  of total: ", total)
	}
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
				name := inf.Name()
				if len(name) > 20 {
					name = name[0:10] + "..." + name[len(name)-10:]
				}
				pro := pg.NewWrappedReaderProgress(inf.Size(), 50, "uploading: ["+name+"]", pg.Top, r)
				ret, err := client.Upload(r, inf.Size(), group, common.InitializedClientConfiguration.PrivateUpload)
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
			name := inf.Name()
			if len(name) > 20 {
				name = name[0:10] + "..." + name[len(name)-10:]
			}
			pro := pg.NewWrappedReaderProgress(inf.Size(), 50, "uploading: ["+name+"]", pg.Top, r)
			ret, err := client.Upload(r, inf.Size(), group, common.InitializedClientConfiguration.PrivateUpload)
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
		err := client.Download(item.(string), 0, -1, func(body io.Reader, bodyLength int64) error {
			// if download only one file and provide a custom filename.
			if downloadFiles.Len() == 1 && customDownloadFileName != "" {
				fi, err := file.CreateFile(customDownloadFileName)
				w := &pg.WrappedWriter{Writer: fi}
				// show download progressbar.
				fid := item.(string)
				if len(fid) > 20 {
					fid = fid[0:10] + "..." + fid[len(fid)-10:]
				}
				pro := pg.NewWrappedWriterProgress(bodyLength, 50, "downloading ==> ["+fid+"]", pg.Top, w)
				_, err = io.Copy(w, body)
				if err != nil {
					pro.Destroy()
				}
				return err
			}
			fileInfo, _, err := util.ParseAlias(item.(string), "")
			if err != nil {
				return err
			}
			md5 := fileInfo.Path[strings.LastIndex(fileInfo.Path, "/")+1:]
			fi, err := file.CreateFile(wd + "/" + md5)
			if err != nil {
				return err
			}
			w := &pg.WrappedWriter{Writer: fi}
			// show download progressbar.
			fid := item.(string)
			if len(fid) > 20 {
				fid = fid[0:10] + "..." + fid[len(fid)-10:]
			}
			pro := pg.NewWrappedWriterProgress(bodyLength, 50, "downloading ==> ["+fid+"]", pg.Top, w)
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

func handleGenerateToken() {
	ts := convert.Int64ToStr(gox.GetTimestamp(time.Now().Add(time.Second * time.Duration(tokenLife))))
	util.GenerateDecKey(secret)
	// the fileId must be parsed by the given secret.
	_, _, err := util.ParseAlias(tokenFileId, secret)
	if err != nil {
		fmt.Println("\nInvalid secret: cannot parse fileId using given secret")
		os.Exit(1)
	}
	token := util.GenerateToken(tokenFileId, secret, ts)
	if tokenFormat == "json" {
		ret := make(map[string]string)
		ret["token"] = token
		ret["ts"] = ts
		r, _ := json.Marshal(ret)
		fmt.Println(string(r))
	} else {
		fmt.Println("token=" + token + "&ts=" + ts)
	}
}

// handleUploadFile handles upload files by client cli.
func handleTestUploadFile() error {
	// initialize APIClient
	if err := initClient(); err != nil {
		logger.Fatal(err)
	}
	startTime := gox.GetTimestamp(time.Now())
	waitGroup := sync.WaitGroup{}
	step := common.InitializedClientConfiguration.TestScale / common.InitializedClientConfiguration.TestThread
	for i := 1; i <= common.InitializedClientConfiguration.TestThread; i++ {
		if i == common.InitializedClientConfiguration.TestThread {
			go uploadTask((i-1)*step, common.InitializedClientConfiguration.TestScale, &waitGroup)
		} else {
			go uploadTask((i-1)*step, i*step, &waitGroup)
		}
	}
	waitGroup.Add(common.InitializedClientConfiguration.TestThread)
	waitGroup.Wait()
	endTime := gox.GetTimestamp(time.Now())
	fmt.Println("[---------------------------]")
	fmt.Println("total  :", common.InitializedClientConfiguration.TestScale)
	fmt.Println("failed :", testFailed)
	fmt.Println("time   :", (endTime-startTime)/1000, "s")
	fmt.Println("average:", int64(common.InitializedClientConfiguration.TestScale)/((endTime-startTime)/1000), "/s")
	fmt.Println("[---------------------------]")
	return nil
}

func uploadTask(start int, end int, waitGroup *sync.WaitGroup) {
	for i := start; i < end; i++ {
		name := convert.IntToStr(i)
		data := []byte(name)
		size := int64(len(data))
		r := bytes.NewReader(data)
		fmt.Println(gox.GetLongLongDateString(time.Now()), "  start upload")
		ret, err := client.Upload(r, size, group, common.InitializedClientConfiguration.PrivateUpload)
		fmt.Println(gox.GetLongLongDateString(time.Now()), "  end   upload")
		if err != nil {
			logger.Error(err)
			updateTestSuccessCount()
		} else {
			bs, _ := json.MarshalIndent(ret, "", "  ")
			logger.Info("upload success:\n", string(bs))
		}
	}
	waitGroup.Done()
}

var testLock = new(sync.Mutex)
var testFailed = 0

func updateTestSuccessCount() {
	testLock.Lock()
	defer testLock.Unlock()
	testFailed++
}
