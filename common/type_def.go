package common

import (
	"github.com/hetianyi/gox/convert"
	"strings"
)

type Command uint32
type Operation byte
type OperationResult byte
type BootMode uint32
type Role byte
type RegisterState byte

type StorageConfig struct {
	Trackers              []string `json:"trackers"`
	Secret                string   `json:"secret"`
	Group                 string   `json:"group"`
	InstanceId            string
	BindAddress           string `json:"bindAddress"`
	Port                  int    `json:"port"`
	AdvertiseAddress      string `json:"advertiseAddress"`
	AdvertisePort         int    `json:"advertisePort"`
	DataDir               string `json:"dataDir"`
	TmpDir                string
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
	InstanceId            string
	BindAddress           string `json:"bindAddress"`
	Port                  int    `json:"port"`
	AdvertiseAddress      string `json:"advertiseAddress"`
	AdvertisePort         int    `json:"advertisePort"`
	DataDir               string `json:"dataDir"`
	PreferredNetworks     string `json:"preferredNetworks"`
	LogLevel              string `json:"logLevel"`
	LogDir                string `json:"logDir"`
	SaveLog2File          bool   `json:"saveLog2File"`
	MaxRollingLogfileSize int    `json:"maxRollingLogfileSize"`
	LogRotationInterval   string `json:"logRotationInterval"`
	EnableHttp            bool   `json:"enableHttp"`
	HttpPort              int    `json:"httpPort"`
	HttpAuth              string `json:"httpAuth"`
	ParsedTrackers        []Server
}

type ClientConfig struct {
	Trackers       []string `json:"trackers"`
	Storages       []string `json:"storages"`
	LogLevel       string   `json:"logLevel"`
	Secret         string   `json:"secret"`
	ParsedTrackers []Server
}

type Server struct {
	Host       string `json:"host"`
	Port       uint16 `json:"port"`
	Secret     string `json:"secret"`
	InstanceId string `json:"instanceId"`
}

type StorageServer struct {
	Server
	Group string `json:"group"`
}

func (s *Server) ConnectionString() string {
	return strings.Join([]string{s.Host, convert.Uint16ToStr(s.Port)}, ":")
}

// GetHost returns server's host.
func (s *Server) GetHost() string {
	return s.Host
}

// GetPort returns server's port.
func (s *Server) GetPort() uint16 {
	return s.Port
}

func (s *StorageServer) ToServer() *Server {
	return &Server{
		Host:       s.Host,
		Port:       s.Port,
		InstanceId: s.InstanceId,
	}
}

type Header struct {
	Operation  Operation         `json:"op"`
	Result     OperationResult   `json:"ret"`
	Msg        string            `json:"msg"`
	Attributes map[string]string `json:"ats"`
}

type UploadResult struct {
	Group  string `json:"group"`
	Node   string `json:"node"`
	FileId string `json:"fileId"`
}

type FileInfo struct {
	Group      string `json:"group"`
	FileId     string `json:"fileId"`
	FileLength int64  `json:"size"`
	CreateTime int64  `json:"createTime"`
}

type Instance struct {
	Server
	Role         Role              `json:"role"`
	Attributes   map[string]string `json:"ats"`
	RegisterTime int64             `json:"ts"`
	State        RegisterState     `json:"state"`
}

type BingLog struct {
	Id        [8]byte  // offset int64
	md5       [32]byte // string 32
	Timestamp [8]byte  // int64
}
