package validate

import (
    "os"
    "strings"
    "util/logger"
)

// adapt to docker environment
// if you don't want conf file, you can specific parameter in system env which is same with conf file
// such as:
// docker run -d -e bas_path="/var/godfs" -e trackers="192.168.1.100:1022,192.168.1.101:1022" ...

var (
    //map[param]
    allParams = map[string]int{
        "group":1,
        "instance_id":1,
        "bind_address":1,
        "port":2,
        "base_path":2,
        "trackers":1,
        "secret":2,
        "log_rotation_interval":2,
        "assign_disk_space":1,
        "slice_size":1,
        "http_enable":1,
        "enable_mime_types":1,
        "http_port":1,
    }
)



func GetEnv(key string) string {
    return os.Getenv(key)
}

// if these param exist in system env , then replace it with system env
func replaceParams(m map[string]string) {
    for k := range m {
        envVal := strings.TrimSpace(GetEnv(k))
        if envVal != "" {
            logger.Warn("parameter", k, "using env value:", envVal)
            m[k] = envVal
        }
    }
}





