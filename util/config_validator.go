package util

import (
	"errors"
	"fmt"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox/convert"
	"github.com/hetianyi/gox/file"
	"github.com/hetianyi/gox/logger"
	"regexp"
	"strings"
)

// ValidateStorageConfig validates storage config.
func ValidateStorageConfig(c *common.StorageConfig) error {
	if c == nil {
		return errors.New("no config provided")
	}

	if c.Port < 0 || c.Port > 65535 {
		return errors.New("invalid port number " + convert.IntToStr(c.Port) + ", port number must in the range of 0 to 65535")
	}

	if m, err := regexp.MatchString("[0-9a-zA-Z-_]{1,30}", c.Group); err != nil || !m {
		return errors.New("invalid group " + c.Group + ", group must match pattern " + "[0-9a-zA-Z-_]{1,30}")
	}

	c.LogLevel = strings.ToLower(c.LogLevel)
	if c.LogLevel != "trace" && c.LogLevel != "debug" && c.LogLevel != "info" &&
		c.LogLevel != "warn" && c.LogLevel != "error" && c.LogLevel != "fatal" {
		c.LogLevel = "info"
	}

	c.LogRotationInterval = strings.ToLower(c.LogRotationInterval)
	if c.LogRotationInterval != "h" && c.LogRotationInterval != "d" &&
		c.LogRotationInterval != "m" && c.LogRotationInterval != "y" {
		c.LogRotationInterval = "y"
	}

	if c.MaxRollingLogfileSize != 64 && c.MaxRollingLogfileSize != 128 && c.MaxRollingLogfileSize != 256 &&
		c.MaxRollingLogfileSize != 512 && c.MaxRollingLogfileSize != 1024 {
		c.MaxRollingLogfileSize = 64
	}

	if c.SaveLog2File {
		if !file.Exists(c.LogDir) {
			if err := file.CreateDirs(c.LogDir); err != nil {
				return err
			}
		}
	}
	logConfig := &logger.Config{
		Level: ConvertLogLevel(c.LogLevel),
	}
	logger.Init(logConfig)

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
