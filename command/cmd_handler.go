package command

var handlers = make(map[Command]func())

func Register(cmd Command, handler func()) {
	handlers[cmd] = handler
}

func GetHandler(cmd Command) func() {
	return handlers[cmd]
}
