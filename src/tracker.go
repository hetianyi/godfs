package main

import (
	"app"
	"github.com/urfave/cli"
	"libclient"
	"libtracker"
	"os"
	"path/filepath"
	"runtime"
	"util/file"
	"util/logger"
	"validate"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	app.RUN_WITH = 2 // run as tracker
	abs, _ := filepath.Abs(os.Args[0]) // executable file path
	s, _ := filepath.Split(abs)
	s = file.FixPath(s)

	initTrackerFlags()

	var confPath string
	if file.IsAbsPath(libclient.ConfigFile) {
		confPath = libclient.ConfigFile
	} else {
		confPath = s + string(filepath.Separator) + libclient.ConfigFile
	}

	logger.Info("using config file:", confPath)
	m, e := file.ReadPropFile(confPath)
	if e == nil {
		validate.Check(m, app.RUN_WITH)
		for k, v := range m {
			logger.Debug(k, "=", v)
		}
		libtracker.StartService()
	} else {
		logger.Fatal("error read file:", e)
	}
}

func initTrackerFlags() {
	appFlag := cli.NewApp()
	appFlag.Version = app.APP_VERSION
	appFlag.Name = "godfs tracker"
	appFlag.Usage = ""

	// config file location
	appFlag.Flags = []cli.Flag {
		cli.StringFlag{
			Name:        "config, c",
			Value:       "../conf/tracker.conf",
			Usage:       "load config from `FILE`",
			Destination: &libclient.ConfigFile,
		},
	}
	
	appFlag.Action = func(c *cli.Context) error {
		return nil
	}

	err := appFlag.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}
}
