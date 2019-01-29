package libclient

import (
	"app"
	"container/list"
	"errors"
	"fmt"
	json "github.com/json-iterator/go"
	"io"
	"libcommon"
	"libcommon/bridge"
	"libcommon/bridgev2"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"util/common"
	"util/file"
	"util/logger"
	"util/timeutil"
)

const (
	CommandNone         = 0
	CommandUpload       = 1
	CommandDownload     = 2
	CommandListConfig   = 3
	CommandUpdateConfig = 4
	CommandInspectFile  = 5
)

var (
	ConfigFile          string
	Trackers            string
	LogLevel            string
	LogRotationInterval string
	LogEnable           = true
	Secret              string
	UploadFileList      list.List
	UpdateConfigList    list.List
	InspectFileList     list.List
	Group               string
	DownloadFilePath    string
	CustomFileName      string
	SetConfig           string
)

var client *Client
var skipCheck = false

// ExecuteCommand execute client command
func ExecuteCommand(oClient *Client, command int) {
	client = oClient
	switch command {
	case CommandListConfig:
		listConfig()
	case CommandUpdateConfig:
		updateConfig()
	case CommandUpload:
		upload()
	case CommandDownload:
		download()
	case CommandInspectFile:
		inspect()
	}
}

// upload upload files.
// paths: file path to be upload
// group: file upload group, if not set, use random group
// skipCheck: whether check md5 before upload
func upload() {
	for ele := UploadFileList.Front(); ele != nil; ele = ele.Next() {
		var startTime = time.Now()
		fid, e := client.Upload(ele.Value.(string), Group, startTime, skipCheck)
		if e != nil {
			logger.Error(e)
		} else {
			now := time.Now()
			fmt.Println("[==========] 100% [" + timeutil.GetHumanReadableDuration(startTime, now) + "]\nupload success, file id:")
			fmt.Println("+-------------------------------------------+")
			fmt.Println(fid)
			fmt.Println("+-------------------------------------------+")
		}
	}
}

// download download file
func download() {
	DownloadFilePath = strings.TrimSpace(DownloadFilePath)
	if strings.Index(DownloadFilePath, "/") != 0 {
		DownloadFilePath = "/" + DownloadFilePath
	}
	filePath := ""
	var startTime time.Time
	e := client.DownloadFile(DownloadFilePath, 0, -1, func(manager *bridgev2.ConnectionManager, frame *bridgev2.Frame, resMeta *bridgev2.DownloadFileResponseMeta) (b bool, e error) {
		var fi *os.File
		if CustomFileName == "" {
			md5 := regexp.MustCompile(app.PathRegex).ReplaceAllString(DownloadFilePath, "${4}")
			CustomFileName = md5
			f, e1 := file.CreateFile(CustomFileName)
			if e1 != nil {
				return true, e1
			}
			fi = f
		} else {
			f, e1 := file.CreateFile(CustomFileName)
			if e1 != nil {
				return true, e1
			}
			fi = f
		}
		defer fi.Close()
		filePath, _ = filepath.Abs(fi.Name())
		startTime = time.Now()
		return true, writeOut(manager.Conn, frame.BodyLength, fi, startTime)
	})
	/*e := client.DownloadFile(path, 0, -1, func(realPath string, fileLen uint64, reader io.Reader) error {
		var fi *os.File
		if customDownloadFileName == "" {
			md5 := regexp.MustCompile(app.PathRegex).ReplaceAllString(realPath, "${4}")
			customDownloadFileName = md5
			f, e1 := file.CreateFile(customDownloadFileName)
			if e1 != nil {
				return e1
			}
			fi = f
		} else {
			f, e1 := file.CreateFile(customDownloadFileName)
			if e1 != nil {
				return e1
			}
			fi = f
		}
		defer fi.Close()
		filePath, _ = filepath.Abs(fi.Name())
		startTime = time.Now()
		return writeOut(reader, int64(fileLen), fi, startTime)
	})*/
	if e != nil {
		logger.Fatal("download failed:", e)
	} else {
		now := time.Now()
		fmt.Println("[==========] 100% [" + timeutil.GetHumanReadableDuration(startTime, now) + "]\ndownload success, file save as:")
		fmt.Println("+-------------------------------------------+")
		fmt.Println(filePath)
		fmt.Println("+-------------------------------------------+")
	}
}

func inspect() {
	common.WalkList(&InspectFileList, func(item interface{}) bool {
		md5 := item.(string)
		fileVO, e := client.QueryFile(md5)
		if e != nil {
			logger.Error("error inspect file", md5, "due to:", e)
		} else {
			if fileVO == nil {
				fmt.Println("file does not exist")
			} else {
				bs, _ := json.MarshalIndent(fileVO, "", "  ")
				fmt.Println(string(bs))
			}
		}
		return false
	})
}

func writeOut(in io.Reader, offset int64, out io.Writer, startTime time.Time) error {
	buffer, _ := bridge.MakeBytes(app.BufferSize, false, 0, false)
	defer bridge.RecycleBytes(buffer)
	var finish, total int64
	var stopFlag = false
	defer func() { stopFlag = true }()
	total = offset
	finish = 0
	go libcommon.ShowPercent(&total, &finish, &stopFlag, startTime)

	// total read bytes
	var readBodySize int64 = 0
	// next time bytes to read
	var nextReadSize int
	for {
		// left bytes is more than a buffer
		if (offset-readBodySize)/int64(len(buffer)) >= 1 {
			nextReadSize = len(buffer)
		} else { // left bytes less than a buffer
			nextReadSize = int(offset - readBodySize)
		}
		if nextReadSize == 0 {
			break
		}
		len, e2 := in.Read(buffer[0:nextReadSize])
		if e2 == nil {
			wl, e5 := out.Write(buffer[0:len])
			if e5 != nil || wl != len {
				return errors.New("error write out")
			}
			finish += int64(len)
			readBodySize += int64(len)
			logger.Trace("write:", readBodySize)
		} else {
			if e2 == io.EOF {
				return nil
			}
			return e2
		}
	}
	return nil
}

func listConfig() {
	configTemp, err := ReadConf()
	common.TOperation(err == nil, func() interface{} {
		bs, _ := json.MarshalIndent(configTemp, "", "  ")
		fmt.Println(string(bs))
		return nil
	}, func() interface{} {
		logger.Fatal(err)
		return nil
	})
}

func updateConfig() {
	configTemp, err := ReadConf()
	common.TOperation(err == nil, func() interface{} {
		common.WalkList(&UpdateConfigList, func(item interface{}) bool {
			set := item.(string)
			k, v := parseConfigItem(set)
			if k == "trackers" {
				configTemp.Trackers = strings.Split(v, ",")
			} else if k == "log_level" {
				if app.LogLevelSet[v] == 0 {
					logger.Error("value of config key log_level must be one of trace|debug|info|warm|error|fatal")
				} else {
					configTemp.LogLevel = v
				}
			} else if k == "log_rotation_interval" {
				if app.LogRotationSet[v] == 0 {
					logger.Error("value of config key log_rotation_interval must be one of h|d|m|y")
				} else {
					configTemp.LogRotationInterval = v
				}
			} else if k == "secret" {
				configTemp.Secret = v
			} else {
				logger.Error("unknown config key:", k)
			}
			return false
		})
		if err := WriteConf(configTemp); err != nil {
			logger.Fatal("cannot write config file:", err)
		} else {
			logger.Info("update config success!")
		}
		return nil
	}, func() interface{} {
		fmt.Println(err)
		return nil
	})

}

// write client config to file.
func WriteConf(clientConfig *app.ClientConfig) error {
	fi, e := file.CreateFile(app.BasePath + string(os.PathSeparator) + "config.json")
	if e != nil {
		return e
	}
	defer fi.Close()
	bs, e1 := json.MarshalIndent(clientConfig, "", "  ")
	if e1 != nil {
		return e1
	}
	fi.Write(bs)
	return nil
}

// read client config
func ReadConf() (*app.ClientConfig, error) {
	configFile, e := file.GetFile(app.BasePath + string(os.PathSeparator) + "config.json")
	if e != nil {
		return nil, e
	}
	defer configFile.Close()

	fi, e1 := configFile.Stat()
	if e1 != nil {
		return nil, e1
	}

	buffer, e2 := bridgev2.MakeBytes(fi.Size(), true, 10240, true)
	if e2 != nil {
		return nil, e2
	}
	_, e3 := configFile.Read(buffer)
	if e3 != nil {
		return nil, e3
	}
	var config = &app.ClientConfig{}
	e4 := json.Unmarshal(buffer, config)
	if e4 != nil {
		return nil, e4
	}
	return config, nil
}

func parseConfigItem(item string) (string, string) {
	firstEQ := strings.Index(item, "=")
	if firstEQ == -1 {
		return item, ""
	} else {
		return item[0:firstEQ], item[firstEQ+1:]
	}
	/*var k, v string
	if len(kv) == 1 {
		k = strings.TrimSpace(kv[0])
	} else if len(kv) > 1 {
		k = strings.TrimSpace(kv[0])
		for i := range kv {
			if i == 0 {
				continue
			}
			v += strings.TrimSpace(kv[i])
		}
	}

	if k == "secret" {
		m[k] = v
	} else if k == "trackers" {
		m[k] = v
	} else if k == "log_enable" {
		if v == "" || (v != "true" && v != "false") {
			v = "true"
		}
		m[k] = v
	} else if k == "log_level" {
		if v != "trace" && v != "debug" && v != "info" && v != "warn" &&
			v != "error" && v != "fatal" {
			v = "info"
		}
		m[k] = v
	} else if k == "log_rotation_interval" {
		if v != "h" && v != "d" &&
			v != "m" && v != "y" {
			v = "d"
		}
		m[k] = v
	} else {
		return errors.New("unknown parameter: \"" + k + "\"")
	}
	logger.Info("set", k, "to", "\""+v+"\"")
	return writeConf(m["trackers"], m["secret"], m["log_enable"], m["log_level"], m["log_rotation_interval"])*/
}
