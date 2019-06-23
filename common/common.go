package common

// fileId=
// fidVersion(1-9a-z)[1] + instanceID[8] + fileLen[8] + crc32[8] + rand[3]
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
	DEFAULT_STORAGE_TCP_PORT           = 3389
	DEFAULT_STORAGE_HTTP_PORT          = 8001
	DEFAULT_TRACKER_TCP_PORT           = 3390
	DEFAULT_TRACKER_HTTP_PORT          = 8002
	BUFFER_SIZE                        = 1 << 15 // 32k
)

// tcp operation codes
const (
	OPERATION_RESPONSE Operation = iota // connect
	OPERATION_CONNECT                   // response
	OPERATION_UPLOAD                    // response
)

// status codes
const (
	SUCCESS OperationResult = iota
	ERROR
	UNAUTHORIZED
)

var (
	Config *StorageConfig
)
