package common

import (
	"errors"
	"regexp"
)

const (
	VERSION                             = "2.0.0"
	BOOT_CLIENT               BootMode  = 0
	BOOT_STORAGE              BootMode  = 1
	BOOT_TRACKER              BootMode  = 2
	GROUP_PATTERN                       = "^[0-9a-zA-Z-_]{1,30}$"
	SECRET_PATTERN                      = "^[^@]{1,30}$"
	SERVER_PATTERN                      = "^(([^@^,]{1,30})@)?([^@]+):([1-9][0-9]{0,5})$"
	HTTP_AUTH_PATTERN                   = "^([^:]+):([^:]+)$"
	INSTANCE_ID_PATTERN                 = "^[0-9a-z-]{8}$"
	FILE_ID_PATTERN                     = "^([0-9a-zA-Z-_]{1,30})/([0-9A-F]{2})/([0-9A-F]{2})/([0-9a-f]{32})$"
	DEFAULT_STORAGE_TCP_PORT            = 9012
	DEFAULT_STORAGE_HTTP_PORT           = 8001
	DEFAULT_TRACKER_TCP_PORT            = 9022
	DEFAULT_TRACKER_HTTP_PORT           = 8011
	BUFFER_SIZE                         = 1 << 15 // 32k
	DEFAULT_GROUP                       = "G01"
	OPERATION_RESPONSE        Operation = iota
	OPERATION_CONNECT
	OPERATION_UPLOAD
	OPERATION_DOWNLOAD
	OPERATION_QUERY
	SUCCESS OperationResult = iota
	ERROR
	UNAUTHORIZED
	NOT_FOUND
	CMD_SHOW_HELP Command = iota
	CMD_SHOW_VERSION
	CMD_UPDATE_CONFIG
	CMD_SHOW_CONFIG
	CMD_UPLOAD_FILE
	CMD_DOWNLOAD_FILE
	CMD_INSPECT_FILE
	CMD_BOOT_TRACKER
	CMD_BOOT_STORAGE
	ROLE_TRACKER Role = iota
	ROLE_STORAGE
	ROLE_PROXY
	REGISTER_HOLD RegisterState = iota
	REGISTER_FREE
)

var (
	NotFoundErr                     = errors.New("file not found")
	ServerErr                       = errors.New("server internal error")
	InitializedTrackerConfiguration *TrackerConfig
	InitializedStorageConfiguration *StorageConfig
	InitializedClientConfiguration  *ClientConfig
	FileIdPatternRegexp             = regexp.MustCompile(FILE_ID_PATTERN)
	ServerPatternRegexp             = regexp.MustCompile(SERVER_PATTERN)
)
