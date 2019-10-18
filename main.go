package main

import (
	"github.com/hetianyi/godfs/command"
	"os"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	/*// Create a context to use with the Goodbye library's functions.
	ctx := context.Background()

	// Always defer `goodbye.Exit` as early as possible since it is
	// safe to execute no matter what.
	defer goodbye.Exit(ctx, -1)

	// Invoke `goodbye.Notify` to begin trapping signals this process
	// might receive. The Notify function can specify which signals are
	// trapped, but if none are specified then a default list is used.
	// The default set is platform dependent. See the files
	// "goodbye_GOOS.go" for more information.
	goodbye.Notify(ctx)

	goodbye.RegisterWithPriority(func(ctx context.Context, sig os.Signal) {
		fmt.Println("system exit", sig)
		logger.Sync()
	}, -1)*/

	args := os.Args[1:]
	for _, arg := range args {
		if arg == "-v" || arg == "--version" {
			args = []string{"--version"}
		}
	}
	newArgs := append([]string{os.Args[0]}, args...)
	command.Parse(newArgs)
}
