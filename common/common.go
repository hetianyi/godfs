package common

const (
	VERSION          = "2.0.0"
	CLIENT  BootMode = 0
	STORAGE BootMode = 1
	TRACKER BootMode = 2
)

type BootMode uint32
