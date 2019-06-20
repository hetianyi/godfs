package main

import (
	"github.com/hetianyi/godfs/command"
	"github.com/hetianyi/gox/logger"
	"os"
)

func init() {
	logger.Init(nil)
}

func main() {
	command.Parse(os.Args)
}
