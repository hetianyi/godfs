package main

import (
	"app"
	"github.com/urfave/cli"
	"libclient"
	"libdashboard"
	"os"
	"path/filepath"
	"util/file"
	"util/logger"
	"validate"
)

// show in dashboard(in minutes):
// --------------------------------
// cpu          record one month
// storage io   record one month
// disk usage   record once
// network io   record one month
// --------------------------------
// tracker hosts 10 caches of each storage,
// delete one cache once it was send successfully to web manager.
//
func main() {
	// set client type
	app.RUN_WITH = 4
	app.CLIENT_TYPE = 3
	app.UUID = "DASHBOARD-CLIENT"

	abs, _ := filepath.Abs(os.Args[0])
	s, _ := filepath.Split(abs)
	s = file.FixPath(s)

	initStorageFlags()

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
		libdashboard.StartService(m)
	} else {
		logger.Fatal("error read file:", e)
	}

}


func initDashboardFlags() {
	appFlag := cli.NewApp()
	appFlag.Version = app.APP_VERSION
	appFlag.Name = "godfs dashboard"
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

