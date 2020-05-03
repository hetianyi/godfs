package util

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/file"
	"github.com/hetianyi/gox/logger"
	"github.com/hetianyi/gox/uuid"
	json "github.com/json-iterator/go"
	"github.com/mitchellh/go-homedir"
	"io"
	"runtime"
	"strings"
)

var (
	az09 = []rune{'A', 'B', 'C', 'D', 'E', 'F', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
)

// LoadInstanceData load old instance id from data dir.
// If there is no old instance id before, create a new instance id for this data dir.
func LoadInstanceData() string {
	configMap := common.GetConfigMap()
	key := "instanceId"
	bInsId, err := configMap.GetConfig(key)
	if err != nil {
		logger.Fatal("cannot load instanceId")
	}
	if bInsId == nil {
		bInsId = []byte(uuid.UUID()[0:8])
		configMap.PutConfig(key, bInsId)
	}
	logger.Debug("instance id:", string(bInsId))
	return string(bInsId)
}

func PrepareDirs(tmpDir string) error {
	file.DeleteAll(tmpDir)
	// tmp dir
	if !file.Exists(tmpDir) {
		return file.CreateDirs(tmpDir)
	}
	return nil
}

// LoadConfig loads config from config file.
func LoadConfig(c string, mode common.BootMode, container interface{}) error {
	cf, err := file.GetFile(c)
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, cf)
	if err != nil {
		return err
	}
	return json.Unmarshal(buffer.Bytes(), container)
}

// WriteConfig writes config to file.
func WriteConfig(c string, container interface{}) error {
	cf, err := file.CreateFile(c)
	if err != nil {
		return err
	}
	bs, err := json.MarshalIndent(container, " ", "")
	if err != nil {
		return err
	}
	_, err = cf.Write(bs)
	if err != nil {
		return err
	}
	return nil
}

// DefaultLogDir returns default system log directory.
func DefaultLogDir() string {
	if runtime.GOOS == "windows" {
		user, err := homedir.Dir()
		if err != nil {
			return "/var/log/godfs"
		}
		return user + "\\AppData\\Local\\godfs\\Logs"
	}
	return "/var/log/godfs"
}

// DefaultLogDir returns default system log directory.
func DefaultDataDir(mode common.BootMode) string {
	user, err := homedir.Dir()
	var middlePath = ""
	if mode == common.BOOT_STORAGE {
		middlePath = "Storage"
	} else if mode == common.BOOT_TRACKER {
		middlePath = "Tracker"
	} else if mode == common.BOOT_CLIENT {
		middlePath = "Client"
	} else if mode == common.BOOT_AGENT {
		middlePath = "Agent"
	}
	if err != nil {
		fmt.Println("cannot get home dir, use tmp dir: " + err.Error())
		return "/tmp/godfs"
	}
	if runtime.GOOS == "windows" {
		return user + "\\AppData\\Local\\godfs\\" + middlePath
	}
	return user + "/godfs/" + strings.ToLower(middlePath)
}

func DefaultAdvertiseAddress() {
	gox.GetMyAddress("")
}

// ParseServers parses server infos from a string.
func ParseServers(servers string) ([]*common.Server, error) {
	var registeredServers []*common.Server
	// parse tracker servers
	if servers != "" {
		ss := strings.Split(servers, ",")
		registeredServers = make([]*common.Server, len(ss))
		for i, s := range ss {
			s = strings.TrimSpace(s)
			s, err := ParseServer(s)
			if err != nil {
				return nil, err
			}
			registeredServers[i] = s
		}
	}
	return registeredServers, nil
}

func ParseServer(s string) (*common.Server, error) {
	if common.ServerPatternRegexp.MatchString(s) {
		secret := common.ServerPatternRegexp.ReplaceAllString(s, "$2")
		host := common.ServerPatternRegexp.ReplaceAllString(s, "$3")
		port, _ := convert.StrToUint16(common.ServerPatternRegexp.ReplaceAllString(s, "$4"))
		return &common.Server{
			Host:   host,
			Port:   port,
			Secret: secret,
		}, nil
	} else {
		return nil, errors.New("invalid server string, format must be the pattern of [<secret>@]<host>:<port>")
	}
}
