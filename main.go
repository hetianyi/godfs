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
	args := os.Args[1:]
	for _, arg := range args {
		if arg == "-v" || arg == "--version" {
			args = []string{"--version"}
		}
	}
	newArgs := append([]string{os.Args[0]}, args...)
	command.Parse(newArgs)
}
