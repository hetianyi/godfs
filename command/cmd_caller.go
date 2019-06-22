package command

import (
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/svc"
)

func call(cmd Command) {
	switch cmd {
	case BOOT_STORAGE:
		svc.BootStorageServer(ConfigAssembly(common.STORAGE).(*common.StorageConfig))
		break
	}
}
