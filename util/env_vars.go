package util

import (
	"github.com/hetianyi/gox/logger"
	"os"
	"strings"
)

func GetEnv(key string) string {
	return os.Getenv(key)
}

// if these param exist in system env , then replace it with system env
func ExchangeEnvValue(key string, then func(envValue string)) {
	envVal := strings.TrimSpace(GetEnv(key))
	if envVal != "" {
		logger.Warn("config property \"", key, "\" load from environment")
		then(envVal)
	}
}
