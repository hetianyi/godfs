package main

import (
	"github.com/hetianyi/godfs/command"
	"os"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	args := os.Args[1:]
	for _, arg := range args {
		if arg == "-v" || arg == "--version" {
			args = []string{"--version"}
		}
	}
	newArgs := append([]string{os.Args[0]}, args...)
	command.Parse(newArgs)
}
