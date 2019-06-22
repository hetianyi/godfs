package common

const (
	VERSION                            = "2.0.0"
	CLIENT                    BootMode = 0
	STORAGE                   BootMode = 1
	TRACKER                   BootMode = 2
	GROUP_PATTERN                      = "^[0-9a-zA-Z-_]{1,30}$"
	SECRET_PATTERN                     = "^[^@]{1,30}$"
	SERVER_PATTERN                     = "^(([^@^,]{1,30})@)?([^@]+):([1-9][0-9]{0,5})$"
	HTTP_AUTH_PATTERN                  = "^([^:]+):([^:]+)$"
	DEFAULT_STORAGE_TCP_PORT           = 3389
	DEFAULT_STORAGE_HTTP_PORT          = 8001
	DEFAULT_TRACKER_TCP_PORT           = 3390
	DEFAULT_TRACKER_HTTP_PORT          = 8002
)
