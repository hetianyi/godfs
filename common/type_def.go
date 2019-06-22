package common

type BootMode uint32

type StorageConfig struct {
	Trackers              []string `json:"trackers"`
	Secret                string   `json:"secret"`
	Group                 string   `json:"group"`
	InstanceId            string   `json:"instanceId"`
	BindAddress           string   `json:"bindAddress"`
	Port                  int      `json:"port"`
	AdvertiseAddress      string   `json:"advertiseAddress"`
	AdvertisePort         int      `json:"advertisePort"`
	DataDir               string   `json:"dataDir"`
	PreferredNetworks     string   `json:"preferredNetworks"`
	LogLevel              string   `json:"logLevel"`
	LogDir                string   `json:"logDir"`
	SaveLog2File          bool     `json:"saveLog2File"`
	MaxRollingLogfileSize int      `json:"maxRollingLogfileSize"`
	LogRotationInterval   string   `json:"logRotationInterval"`
	EnableHttp            bool     `json:"enableHttp"`
	HttpPort              int      `json:"httpPort"`
	HttpAuth              string   `json:"httpAuth"`
	EnableMimeTypes       bool     `json:"enableMimeTypes"`
	AllowedDomains        []string `json:"allowedDomains"`
	ParsedTrackers        []Server
}

type TrackerConfig struct {
	Trackers              []string `json:"trackers"`
	Secret                string   `json:"secret"`
	InstanceId            string   `json:"instanceId"`
	BindAddress           string   `json:"bindAddress"`
	Port                  int      `json:"port"`
	AdvertiseAddress      string   `json:"advertiseAddress"`
	AdvertisePort         int      `json:"advertisePort"`
	DataDir               string   `json:"dataDir"`
	PreferredNetworks     string   `json:"preferredNetworks"`
	LogLevel              string   `json:"logLevel"`
	LogDir                string   `json:"logDir"`
	SaveLog2File          bool     `json:"saveLog2File"`
	MaxRollingLogfileSize int      `json:"maxRollingLogfileSize"`
	LogRotationInterval   string   `json:"logRotationInterval"`
	EnableHttp            bool     `json:"enableHttp"`
	HttpPort              int      `json:"httpPort"`
	HttpAuth              string   `json:"httpAuth"`
	EnableMimeTypes       bool     `json:"enableMimeTypes"`
	AllowedDomains        []string `json:"allowedDomains"`
}

type ClientConfig struct {
	Trackers []string `json:"trackers"`
	Storages []string `json:"storages"`
	LogLevel string   `json:"logLevel"`
	Secret   string   `json:"secret"`
}

type Server struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Secret string `json:"secret"`
}

type Operation byte

type Header struct {
	Operation  Operation              `json:"op"`
	Result     byte                   `json:"ret"`
	Msg        string                 `json:"msg"`
	Attributes map[string]interface{} `json:"ats"`
}
