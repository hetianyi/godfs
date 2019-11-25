package util

import (
	"errors"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/file"
	"github.com/hetianyi/gox/logger"
	json "github.com/json-iterator/go"
	"regexp"
	"strings"
	"sync"
)

var storeSecretLock *sync.Mutex

func init() {
	storeSecretLock = new(sync.Mutex)
}

// ValidateStorageConfig validates storage config.
func ValidateStorageConfig(c *common.StorageConfig) error {
	if c == nil {
		return errors.New("no config provided")
	}
	// check port range
	if c.Port < 0 || c.Port > 65535 {
		return errors.New("invalid port number " +
			convert.IntToStr(c.Port) + ", port number must in the range of 0 to 65535")
	}
	// check advertise port range
	if c.AdvertisePort < 0 || c.AdvertisePort > 65535 {
		return errors.New("invalid advertise port number " +
			convert.IntToStr(c.Port) + ", port number must in the range of 0 to 65535")
	}
	// check http port range
	if c.HttpPort < 0 || c.HttpPort > 65535 {
		return errors.New("invalid http port number " +
			convert.IntToStr(c.Port) + ", port number must in the range of 0 to 65535")
	}
	// check group
	if c.Group == "" {
		c.Group = common.DEFAULT_GROUP
	}
	if m, err := regexp.MatchString(common.GROUP_PATTERN, c.Group); err != nil || !m {
		return errors.New("invalid group \"" + c.Group +
			"\", group must match pattern " + common.GROUP_PATTERN)
	}
	// check secret
	if c.Secret != "" {
		if m, err := regexp.MatchString(common.SECRET_PATTERN, c.Secret); err != nil || !m {
			return errors.New("invalid secret \"" + c.Secret +
				"\", secret must match pattern " + common.SECRET_PATTERN)
		}
	}

	// check log level
	c.LogLevel = strings.ToLower(c.LogLevel)
	if c.LogLevel != "trace" && c.LogLevel != "debug" && c.LogLevel != "info" &&
		c.LogLevel != "warn" && c.LogLevel != "error" && c.LogLevel != "fatal" {
		c.LogLevel = "info"
	}
	// check log rotation interval
	c.LogRotationInterval = strings.ToLower(c.LogRotationInterval)
	if c.LogRotationInterval != "h" && c.LogRotationInterval != "d" &&
		c.LogRotationInterval != "m" && c.LogRotationInterval != "y" {
		c.LogRotationInterval = "y"
	}
	// check rolling log file size
	if c.MaxRollingLogfileSize != 64 && c.MaxRollingLogfileSize != 128 &&
		c.MaxRollingLogfileSize != 256 && c.MaxRollingLogfileSize != 512 &&
		c.MaxRollingLogfileSize != 1024 {
		c.MaxRollingLogfileSize = 64
	}
	// prepare log directory
	if c.SaveLog2File {
		if !file.Exists(c.LogDir) {
			if err := file.CreateDirs(c.LogDir); err != nil {
				return err
			}
		}
	}
	c.DataDir = file.FixPath(c.DataDir)
	if !file.Exists(c.DataDir) {
		if err := file.CreateDirs(c.DataDir); err != nil {
			return err
		}
	}

	// initialize logger
	logConfig := &logger.Config{
		Level:              ConvertLogLevel(c.LogLevel),
		RollingPolicy:      []int{ConvertRollInterval(c.LogRotationInterval), ConvertLogFileSize(c.MaxRollingLogfileSize)},
		Write2File:         c.SaveLog2File,
		AlwaysWriteConsole: true,
		RollingFileDir:     c.LogDir,
		RollingFileName:    "godfs-storage",
	}
	logger.Init(logConfig)

	InitialConfigMap(c.DataDir + "/cfg.dat")
	c.InstanceId = LoadInstanceData()

	historySecret, err := loadHistorySecret(true, c.InstanceId, c.Secret)
	if err != nil {
		logger.Fatal(err)
	}
	c.HistorySecrets = historySecret
	GenerateDecKey(c.Secret)

	// parse tracker servers
	if c.Trackers != nil {
		if c.ParsedTrackers == nil {
			c.ParsedTrackers = make([]common.Server, len(c.Trackers))
		}
		for i, t := range c.Trackers {
			server, err := ParseServer(t)
			if err != nil {
				return err
			}
			c.ParsedTrackers[i] = *server
		}
	}
	// done!
	return nil
}

// ValidateTrackerConfig validates tracker config.
func ValidateTrackerConfig(c *common.TrackerConfig) error {
	if c == nil {
		return errors.New("no config provided")
	}
	// check port range
	if c.Port < 0 || c.Port > 65535 {
		return errors.New("invalid port number " +
			convert.IntToStr(c.Port) + ", port number must in the range of 0 to 65535")
	}
	// check advertise port range
	if c.AdvertisePort < 0 || c.AdvertisePort > 65535 {
		return errors.New("invalid advertise port number " +
			convert.IntToStr(c.Port) + ", port number must in the range of 0 to 65535")
	}
	// check http port range
	if c.HttpPort < 0 || c.HttpPort > 65535 {
		return errors.New("invalid http port number " +
			convert.IntToStr(c.Port) + ", port number must in the range of 0 to 65535")
	}
	// check secret
	if c.Secret != "" {
		if m, err := regexp.MatchString(common.SECRET_PATTERN, c.Secret); err != nil || !m {
			return errors.New("invalid secret \"" + c.Secret +
				"\", secret must match pattern " + common.SECRET_PATTERN)
		}
	}

	// check log level
	c.LogLevel = strings.ToLower(c.LogLevel)
	if c.LogLevel != "trace" && c.LogLevel != "debug" && c.LogLevel != "info" &&
		c.LogLevel != "warn" && c.LogLevel != "error" && c.LogLevel != "fatal" {
		c.LogLevel = "info"
	}
	// check log rotation interval
	c.LogRotationInterval = strings.ToLower(c.LogRotationInterval)
	if c.LogRotationInterval != "h" && c.LogRotationInterval != "d" &&
		c.LogRotationInterval != "m" && c.LogRotationInterval != "y" {
		c.LogRotationInterval = "y"
	}
	// check rolling log file size
	if c.MaxRollingLogfileSize != 64 && c.MaxRollingLogfileSize != 128 &&
		c.MaxRollingLogfileSize != 256 && c.MaxRollingLogfileSize != 512 &&
		c.MaxRollingLogfileSize != 1024 {
		c.MaxRollingLogfileSize = 64
	}
	// prepare log directory
	if c.SaveLog2File {
		if !file.Exists(c.LogDir) {
			if err := file.CreateDirs(c.LogDir); err != nil {
				return err
			}
		}
	}
	c.DataDir = file.FixPath(c.DataDir)
	if !file.Exists(c.DataDir) {
		if err := file.CreateDirs(c.DataDir); err != nil {
			return err
		}
	}

	// initialize logger
	logConfig := &logger.Config{
		Level:              ConvertLogLevel(c.LogLevel),
		RollingPolicy:      []int{ConvertRollInterval(c.LogRotationInterval), ConvertLogFileSize(c.MaxRollingLogfileSize)},
		Write2File:         c.SaveLog2File,
		AlwaysWriteConsole: true,
		RollingFileDir:     c.LogDir,
		RollingFileName:    "godfs-storage",
	}
	logger.Init(logConfig)

	InitialConfigMap(c.DataDir + "/cfg.dat")
	c.InstanceId = LoadInstanceData()

	historySecret, err := loadHistorySecret(false, c.InstanceId, c.Secret)
	if err != nil {
		logger.Fatal(err)
	}
	c.HistorySecrets = historySecret
	GenerateDecKey(c.Secret)

	// parse tracker servers
	if c.Trackers != nil {
		if c.ParsedTrackers == nil {
			c.ParsedTrackers = make([]common.Server, len(c.Trackers))
		}
		for i, t := range c.Trackers {
			server, err := ParseServer(t)
			if err != nil {
				return err
			}
			c.ParsedTrackers[i] = *server
		}
	}
	// done!
	return nil
}

// ValidateStorageConfig validates storage config.
func ValidateClientConfig(c *common.ClientConfig) error {
	if c == nil {
		return errors.New("no config provided")
	}
	// check secret
	if c.Secret != "" {
		if m, err := regexp.MatchString(common.SECRET_PATTERN, c.Secret); err != nil || !m {
			return errors.New("invalid secret \"" + c.Secret +
				"\", secret must match pattern " + common.SECRET_PATTERN)
		}
	}
	c.LogLevel = strings.ToLower(c.LogLevel)
	if c.LogLevel != "trace" && c.LogLevel != "debug" && c.LogLevel != "info" &&
		c.LogLevel != "warn" && c.LogLevel != "error" && c.LogLevel != "fatal" {
		c.LogLevel = "info"
	}
	// initialize logger
	logConfig := &logger.Config{
		Level:              ConvertLogLevel(c.LogLevel),
		Write2File:         false,
		AlwaysWriteConsole: true,
		Formatter:          &logger.NoneTextFormatter{},
	}
	logger.Init(logConfig)

	// TODO Extract public parts
	// parse tracker servers
	if c.Trackers != nil {
		if c.ParsedTrackers == nil {
			c.ParsedTrackers = make([]common.Server, len(c.Trackers))
		}
		for i, t := range c.Trackers {
			server, err := ParseServer(t)
			if err != nil {
				return err
			}
			c.ParsedTrackers[i] = *server
		}
	}
	// done!
	return nil
}

func InitialConfigMap(path string) {
	logger.Debug("initial config map: ", path)
	configMap, err := common.NewConfigMap(path)
	if err != nil {
		logger.Fatal("cannot initialize configMap file")
	}
	common.SetConfigMap(configMap)
}

func ConvertLogLevel(levelString string) logger.Level {
	levelString = strings.ToLower(levelString)
	switch levelString {
	case "trace":
		return logger.TraceLevel
	case "debug":
		return logger.DebugLevel
	case "info":
		return logger.InfoLevel
	case "warn":
		return logger.WarnLevel
	case "error":
		return logger.ErrorLevel
	case "fatal":
		return logger.FatalLevel
	default:
		return logger.InfoLevel
	}
}

func ConvertRollInterval(rollString string) int {
	rollString = strings.ToLower(rollString)
	switch rollString {
	case "h":
		return logger.HOUR
	case "d":
		return logger.DAY
	case "m":
		return logger.MONTH
	case "y":
		return logger.YEAR
	default:
		return logger.YEAR
	}
}

func ConvertLogFileSize(s int) int {
	switch s {
	case 64:
		return logger.MB64
	case 128:
		return logger.MB128
	case 256:
		return logger.MB256
	case 512:
		return logger.MB512
	case 1024:
		return logger.MB1024
	default:
		return logger.SIZE_NO_LIMIT
	}
}

// loadHistorySecret loads history secrets from db.
func loadHistorySecret(store bool, instanceId, secret string) (map[string]string, error) {
	if store {
		if err := StoreSecrets(instanceId, secret); err != nil {
			return nil, err
		}
	}
	sm, err := GetSecrets()
	if err != nil {
		return nil, err
	}
	for k, _ := range sm {
		AddSecretEncryptKeys(k)
	}
	return sm, nil
}

func GetSecrets() (map[string]string, error) {
	retMap := make(map[string]string)
	configKey := "secrets"
	configMap := common.GetConfigMap()
	ret, err := configMap.GetConfig(configKey)
	if err != nil {
		return nil, err
	}
	if ret == nil || len(ret) == 0 {
		return nil, nil
	}
	err = json.Unmarshal(ret, &retMap)
	if err != nil {
		return nil, err
	}
	return retMap, nil
}

// StoreSecrets stores secrets to config bucket.
func StoreSecrets(instanceId string, secret ...string) error {
	storeSecretLock.Lock()
	defer storeSecretLock.Unlock()

	logger.Debug("store secrets for instance ", instanceId, ":", secret)
	if secret == nil {
		return nil
	}
	newSecrets := make(map[string]string)
	for _, s := range secret {
		if common.GetSecret(s) != "" {
			continue
		}
		newSecrets[s] = instanceId
		common.AddSecret(instanceId, s)
		AddSecretEncryptKeys(s)
	}

	if len(newSecrets) == 0 {
		logger.Debug("now new secrets to store")
		return nil
	}

	retMap := make(map[string]string)
	configKey := "secrets"
	configMap := common.GetConfigMap()
	ret, err := configMap.GetConfig(configKey)
	if err != nil {
		return err
	}
	if ret != nil && len(ret) > 0 {
		err = json.Unmarshal(ret, &retMap)
		if err != nil {
			return err
		}
	}
	for k, v := range newSecrets {
		retMap[k] = v
	}

	ret, err = json.Marshal(retMap)
	if err != nil {
		return err
	}
	if err = configMap.PutConfig(configKey, ret); err != nil {
		return err
	}
	return nil
}

func CollectMapKeys(m map[string]string) []string {
	if m == nil {
		return nil
	}
	ret := make([]string, len(m))
	index := 0
	for i, _ := range m {
		ret[index] = i
		index++
	}
	return ret
}
