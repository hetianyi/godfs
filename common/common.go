package common

import (
	"errors"
	"regexp"
	"time"
)

const (
	VERSION = "2.0.0-dev"
	//
	BOOT_CLIENT  BootMode = 0
	BOOT_STORAGE BootMode = 1
	BOOT_TRACKER BootMode = 2
	BOOT_AGENT   BootMode = 3
	//
	GROUP_PATTERN       = "^[0-9a-zA-Z-_]{1,30}$"
	SECRET_PATTERN      = "^[^@]{1,30}$"
	SERVER_PATTERN      = "^(([^@^,]{1,30})@)?([^@]+):([1-9][0-9]{0,5})$"
	HTTP_AUTH_PATTERN   = "^([^:]+):([^:]+)$"
	INSTANCE_ID_PATTERN = "^[0-9a-z-]{8}$"
	FILE_META_PATTERN   = "^([0-9a-zA-Z-_]{1,30})/([0-9A-F]{2})/([0-9A-F]{2})/([0-9a-f]{32})$"
	//
	DEFAULT_STORAGE_TCP_PORT  = 10706
	DEFAULT_STORAGE_HTTP_PORT = 11222
	DEFAULT_TRACKER_TCP_PORT  = 11706
	DEFAULT_TRACKER_HTTP_PORT = 12222
	BUFFER_SIZE               = 1 << 15 // 32k
	DEFAULT_GROUP             = "G01"
	//
	OPERATION_RESPONSE       Operation = 0
	OPERATION_CONNECT        Operation = 1
	OPERATION_UPLOAD         Operation = 2
	OPERATION_DOWNLOAD       Operation = 3
	OPERATION_QUERY          Operation = 4
	OPERATION_SYNC_INSTANCES Operation = 5
	OPERATION_PUSH_BINLOGS   Operation = 6
	OPERATION_SYNC_BINLOGS   Operation = 7
	//
	SUCCESS           OperationResult = 0
	ERROR             OperationResult = 1
	UNAUTHORIZED      OperationResult = 2
	NOT_FOUND         OperationResult = 3
	UNKNOWN_OPERATION OperationResult = 4
	//
	CMD_SHOW_HELP      Command = 0
	CMD_SHOW_VERSION   Command = 1
	CMD_UPDATE_CONFIG  Command = 2
	CMD_SHOW_CONFIG    Command = 3
	CMD_UPLOAD_FILE    Command = 4
	CMD_DOWNLOAD_FILE  Command = 5
	CMD_INSPECT_FILE   Command = 6
	CMD_BOOT_TRACKER   Command = 7
	CMD_BOOT_STORAGE   Command = 8
	CMD_TEST_UPLOAD    Command = 9
	CMD_GENERATE_TOKEN Command = 10
	CMD_BOOT_AGENT     Command = 11
	//
	ROLE_TRACKER Role = 1
	ROLE_STORAGE Role = 2
	ROLE_PROXY   Role = 3
	ROLE_CLIENT  Role = 4
	ROLE_ANY     Role = 5
	//
	REGISTER_HOLD RegisterState = 1
	REGISTER_FREE RegisterState = 2
	//
	REGISTER_INTERVAL    = time.Second * 30
	SYNCHRONIZE_INTERVAL = time.Second * 45

	FILE_ID_SIZE = 86

	BUCKET_KEY_CONFIGMAP         = "configMap"
	BUCKET_KEY_FAILED_BINLOG_POS = "failedBinlogPos"
	BUCKET_KEY_FILEID            = "fileIds"
)

var (
	NotFoundErr                     = errors.New("file not found")
	ServerErr                       = errors.New("server internal error")
	InitializedTrackerConfiguration *TrackerConfig
	InitializedStorageConfiguration *StorageConfig
	InitializedAgentConfiguration   *AgentConfig
	InitializedClientConfiguration  *ClientConfig
	FileMetaPatternRegexp           = regexp.MustCompile(FILE_META_PATTERN)
	ServerPatternRegexp             = regexp.MustCompile(SERVER_PATTERN)
	BootAs                          BootMode
	configMap                       *ConfigMap
	CusterSecret                    = make(map[string]string)
)

func SetConfigMap(config *ConfigMap) {
	configMap = config
}

func GetConfigMap() *ConfigMap {
	return configMap
}

func AddSecret(instanceId string, secret ...string) {
	if secret == nil {
		return
	}
	for _, s := range secret {
		CusterSecret[s] = instanceId
	}
}

func GetSecret(secret string) (instance string) {
	return CusterSecret[secret]
}
