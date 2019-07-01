package common

import (
	"errors"
	"regexp"
)

// fileId=
// group/crc32[end4]/crc32[end2]/md5
//
const (
	VERSION                            = "2.0.0"
	CLIENT                    BootMode = 0
	STORAGE                   BootMode = 1
	TRACKER                   BootMode = 2
	GROUP_PATTERN                      = "^[0-9a-zA-Z-_]{1,30}$"
	SECRET_PATTERN                     = "^[^@]{1,30}$"
	SERVER_PATTERN                     = "^(([^@^,]{1,30})@)?([^@]+):([1-9][0-9]{0,5})$"
	HTTP_AUTH_PATTERN                  = "^([^:]+):([^:]+)$"
	INSTANCE_ID_PATTERN                = "^[0-9a-z-]{8}$"
	FILE_ID_PATTERN                    = "^([0-9a-zA-Z-_]{1,30})/([0-9A-F]{2})/([0-9A-F]{2})/([0-9a-f]{32})$"
	DEFAULT_STORAGE_TCP_PORT           = 9012
	DEFAULT_STORAGE_HTTP_PORT          = 8001
	DEFAULT_TRACKER_TCP_PORT           = 9022
	DEFAULT_TRACKER_HTTP_PORT          = 8011
	BUFFER_SIZE                        = 1 << 15 // 32k
	DEFAULT_GROUP                      = "G01"
)

// tcp operation codes
const (
	OPERATION_RESPONSE Operation = iota // connect
	OPERATION_CONNECT                   // response
	OPERATION_UPLOAD                    // response
	OPERATION_DOWNLOAD                  // response
	OPERATION_QUERY                     // response
)

// status codes
const (
	SUCCESS OperationResult = iota
	ERROR
	UNAUTHORIZED
	NOT_FOUND
)

var (
	NotFoundErr = errors.New("file not found")
	ServerErr   = errors.New("server internal error")
)

var (
	InitializedStorageConfiguration *StorageConfig
	InitializedClientConfiguration  *ClientConfig
	FileIdPatternRegexp             = regexp.MustCompile(FILE_ID_PATTERN)
	ServerPatternRegexp             = regexp.MustCompile(SERVER_PATTERN)
)
