package command

import (
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/svc"
)

// call calls handler function due to command.
func call(cmd common.Command) {
	switch cmd {
	case common.CMD_BOOT_STORAGE:
		common.BootAs = common.BOOT_STORAGE
		ConfigAssembly(common.BOOT_STORAGE)
		svc.BootStorageServer()
		break
	case common.CMD_BOOT_TRACKER:
		common.BootAs = common.BOOT_TRACKER
		ConfigAssembly(common.BOOT_TRACKER)
		svc.BootTrackerServer()
		break
	case common.CMD_UPLOAD_FILE:
		common.BootAs = common.BOOT_CLIENT
		ConfigAssembly(common.BOOT_CLIENT)
		handleUploadFile()
		break
	case common.CMD_DOWNLOAD_FILE:
		common.BootAs = common.BOOT_CLIENT
		ConfigAssembly(common.BOOT_CLIENT)
		handleDownloadFile()
		break
	case common.CMD_INSPECT_FILE:
		common.BootAs = common.BOOT_CLIENT
		ConfigAssembly(common.BOOT_CLIENT)
		handleInspectFile()
		break
	}
}
