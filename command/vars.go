package command

import "container/list"

const (
	SHOW_HELP Command = iota
	SHOW_VERSION
	UPDATE_CONFIG
	SHOW_CONFIG
	UPLOAD_FILE
	DOWNLOAD_FILE
	INSPECT_FILE
	BOOT_TRACKER
	BOOT_STORAGE
)

type Command uint32

// var sets
var (
	clientModel            bool      // flag boot as client mode
	storageModel           bool      // flag boot as storage mode
	trackerModel           bool      // flag boot as tracker mode
	privateUpload          bool      // upload private file
	uploadGroup            string    // upload group
	downloadFid            string    // fid to be downloaded
	customDownloadFileName string    // custom file download location and filename
	inspectFiles           list.List // custom file download location and filename
	updateConfigList       list.List // configs to be update
	configFile             string    // specified config file to be use

	trackers    string    // tracker servers(used by client and storage mode)
	storages    string    // storage servers(used by storage mode)
	logLevel    string    // log level(trace, debug, info, warn, error, fatal)
	secret      string    // secret of this instance
	uploadFiles list.List // files to be uploaded
)

var finalCommand Command
