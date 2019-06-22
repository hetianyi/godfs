package command

import (
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/svc"
)

func call(cmd Command) {
	switch cmd {
	case BOOT_STORAGE:
		ConfigAssembly(common.STORAGE)
		svc.BootStorageServer()
		break
	}
}
