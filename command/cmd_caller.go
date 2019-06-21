package command

func call(cmd Command) {
	switch cmd {
	case BOOT_STORAGE:
		bootStorageServer()
		break
	}
}

func bootStorageServer() {

}
