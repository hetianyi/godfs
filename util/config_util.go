package util

import (
	"bytes"
	"encoding/json"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/file"
	"github.com/mitchellh/go-homedir"
	"io"
	"runtime"
)

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
func DefaultDataDir() string {
	user, err := homedir.Dir()
	if err != nil {
		return "/tmp/godfs"
	}
	if runtime.GOOS == "windows" {
		return user + "\\AppData\\Local\\godfs\\Data"
	}
	return user + "/godfs/data"
}

func DefaultAdvertiseAddress() {
	gox.GetMyAddress("")
}
