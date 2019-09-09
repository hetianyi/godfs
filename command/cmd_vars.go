package command

import (
	"container/list"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/hetianyi/gox"
	"strings"
)

// var sets
var (
	privateUpload          bool      // upload private file
	showVersion            bool      // show app version
	uploadGroup            string    // upload group
	customDownloadFileName string    // custom fi                le download location and filename
	inspectFiles           list.List // custom file download location and filename
	updateConfigList       list.List // configs to be update
	configFile             string    // specified config file to be use
	trackers               string    // tracker servers(used by client and storage mode)
	storages               string    // storage servers(used by client mode)
	logLevel               string    // log level(trace, debug, info, warn, error, fatal)
	secret                 string    // secret of this instance
	uploadFiles            list.List // files to be uploaded
	downloadFiles          list.List // files to be downloaded
	group                  string
	instanceId             string
	bindAddress            string
	port                   int
	advertiseAddress       string
	advertisePort          int
	dataDir                string
	preferredNetwork       string
	maxLogfileSize         int
	logRotationInterval    string
	disableHttp            bool
	httpPort               int
	httpAuth               string
	enableMimetypes        bool
	allowedDomains         string
	logDir                 string
	disableSaveLogfile     bool
	finalCommand           common.Command
)

//
func ConfigAssembly(bm common.BootMode) interface{} {
	if bm == common.BOOT_STORAGE {
		c := &common.StorageConfig{}
		c.Port = gox.TValue(port <= 0, common.DEFAULT_STORAGE_TCP_PORT, port).(int)
		c.AdvertisePort = gox.TValue(advertisePort <= 0, c.Port, advertisePort).(int)
		c.HttpPort = gox.TValue(httpPort <= 0, common.DEFAULT_STORAGE_HTTP_PORT, httpPort).(int)
		c.Group = group
		c.Secret = secret
		c.HttpAuth = httpAuth
		c.LogLevel = logLevel
		c.LogRotationInterval = logRotationInterval
		c.MaxRollingLogfileSize = maxLogfileSize
		c.SaveLog2File = !disableSaveLogfile

		if logDir == "" {
			logDir = util.DefaultLogDir()
		}
		c.LogDir = logDir

		if dataDir == "" {
			dataDir = util.DefaultDataDir()
		}
		c.DataDir = dataDir
		c.TmpDir = dataDir + "/tmp"

		if advertisePort == 0 {
			advertisePort = c.Port
		}
		if advertiseAddress == "" {
			c.AdvertiseAddress = gox.GetMyAddress(preferredNetwork)
		}
		c.PreferredNetworks = preferredNetwork

		if bindAddress == "" {
			bindAddress = "127.0.0.1"
		}
		c.BindAddress = bindAddress
		c.EnableMimeTypes = enableMimetypes
		c.EnableHttp = !disableHttp

		if trackers != "" {
			c.Trackers = strings.Split(trackers, ",")
		}
		if allowedDomains != "" {
			c.AllowedDomains = strings.Split(allowedDomains, ",")
		}
		common.InitializedStorageConfiguration = c
		return c
	} else if bm == common.BOOT_TRACKER {
		c := &common.TrackerConfig{}
		c.Port = gox.TValue(port <= 0, common.DEFAULT_TRACKER_TCP_PORT, port).(int)
		c.AdvertisePort = gox.TValue(advertisePort <= 0, c.Port, advertisePort).(int)
		c.HttpPort = gox.TValue(httpPort <= 0, common.DEFAULT_TRACKER_HTTP_PORT, httpPort).(int)
		c.Secret = secret
		c.HttpAuth = httpAuth
		c.LogLevel = logLevel
		c.LogRotationInterval = logRotationInterval
		c.MaxRollingLogfileSize = maxLogfileSize
		c.SaveLog2File = !disableSaveLogfile

		if logDir == "" {
			logDir = util.DefaultLogDir()
		}
		c.LogDir = logDir

		if dataDir == "" {
			dataDir = util.DefaultDataDir()
		}
		c.DataDir = dataDir

		if advertisePort == 0 {
			advertisePort = c.Port
		}
		if advertiseAddress == "" {
			c.AdvertiseAddress = gox.GetMyAddress(preferredNetwork)
		}
		c.PreferredNetworks = preferredNetwork

		if bindAddress == "" {
			bindAddress = "127.0.0.1"
		}
		c.BindAddress = bindAddress
		c.EnableHttp = !disableHttp

		if trackers != "" {
			c.Trackers = strings.Split(trackers, ",")
		}
		common.InitializedTrackerConfiguration = c
		return c
	} else if bm == common.BOOT_CLIENT {
		c := &common.ClientConfig{}
		c.Secret = secret
		c.LogLevel = logLevel
		if trackers != "" {
			c.Trackers = strings.Split(trackers, ",")
		}
		c.PrivateUpload = privateUpload
		common.InitializedClientConfiguration = c
		return c
	}
	return nil
}
