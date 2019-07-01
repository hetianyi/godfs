package command

import (
	"errors"
	"fmt"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/urfave/cli"
	"os"
)

func Parse(arguments []string) {
	appFlag := cli.NewApp()
	appFlag.Version = common.VERSION
	appFlag.HideVersion = true
	appFlag.Name = "godfs"
	appFlag.Usage = "godfs"
	appFlag.HelpName = "godfs"
	// config file location
	appFlag.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "log-level",
			Value: "",
			Usage: `set log level, available options:
	(trace|debug|info|warn|error|fatal)`,
			Destination: &logLevel,
		},
		cli.IntFlag{
			Name:  "max-logfile-size",
			Value: 0,
			Usage: `rolling log file max size, available options:
	(0|64|128|256|512|1024)`,
			Destination: &maxLogfileSize,
		},
		cli.StringFlag{
			Name:        "log-rotation-interval",
			Value:       "d",
			Usage:       "log rotation interval(h|d|m|y)",
			Destination: &logRotationInterval,
		},
		cli.BoolFlag{
			Name:        "version, v",
			Usage:       `show version`,
			Destination: &showVersion,
		},
	}

	appFlag.Commands = []cli.Command{
		{
			Name:  "tracker",
			Usage: "start as tracker server",
			Action: func(c *cli.Context) error {
				finalCommand = BOOT_TRACKER
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "config, c",
					Value:       "",
					Usage:       "use custom config file",
					Destination: &configFile,
				},
				cli.StringFlag{
					Name:        "secret, s",
					Value:       "",
					Usage:       "custom global secret",
					Destination: &secret,
				}, /*
					cli.StringFlag{
						Name:        "instance-id",
						Value:       "",
						Usage:       "set instance id of server instance",
						Destination: &instanceId,
					},*/
				cli.StringFlag{
					Name:        "bind-address",
					Value:       "",
					Usage:       "bind listening address",
					Destination: &bindAddress,
				},
				cli.IntFlag{
					Name:        "port, p",
					Value:       0,
					Usage:       "server tcp port",
					Destination: &port,
				},
				cli.StringFlag{
					Name:        "advertise-address",
					Value:       "",
					Usage:       "advertise address is the broadcast address",
					Destination: &advertiseAddress,
				},
				cli.IntFlag{
					Name:        "advertise-port",
					Value:       0,
					Usage:       "advertise port is the broadcast port",
					Destination: &advertisePort,
				},
				cli.StringFlag{
					Name:        "data-dir",
					Value:       "",
					Usage:       "data directory",
					Destination: &dataDir,
				},
				cli.StringFlag{
					Name:        "preferred-network",
					Value:       "",
					Usage:       "choose preferred network interface for registering",
					Destination: &preferredNetwork,
				},
				cli.BoolTFlag{
					Name:        "enable-http",
					Usage:       "enable http server",
					Destination: &enableHttp,
				},
				cli.IntFlag{
					Name:        "http-port",
					Value:       0,
					Usage:       "http port",
					Destination: &httpPort,
				},
				cli.StringFlag{
					Name:        "http-auth",
					Value:       "",
					Usage:       "http authentication",
					Destination: &httpAuth,
				},
				cli.BoolFlag{
					Name:        "enable-mimetypes",
					Usage:       "enable http mime type",
					Destination: &enableMimetypes,
				},
				cli.StringFlag{
					Name:        "allowed-domains",
					Usage:       "allowed access domains",
					Destination: &allowedDomains,
				},
				cli.StringFlag{
					Name:  "trackers",
					Value: "",
					Usage: `set tracker servers, example:
	[<secret1>@]host1:port1,[<secret2>@]host2:port2`,
					Destination: &trackers,
				},
				cli.StringFlag{
					Name:        "logdir",
					Value:       "",
					Usage:       "set log directory",
					Destination: &logDir,
				},
				cli.BoolFlag{
					Name:        "disable-logfile",
					Usage:       "disable save log to file",
					Destination: &disableSaveLogfile,
				},
			},
		},
		{
			Name:  "storage",
			Usage: "start as storage server",
			Action: func(c *cli.Context) error {
				finalCommand = BOOT_STORAGE
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "secret, s",
					Value:       "",
					Usage:       "custom global secret",
					Destination: &secret,
				},
				cli.StringFlag{
					Name:        "group, g",
					Value:       "",
					Usage:       "set group name of storage server",
					Destination: &group,
				},
				cli.StringFlag{
					Name:        "config, c",
					Value:       "",
					Usage:       "use custom config file",
					Destination: &configFile,
				}, /*
					cli.StringFlag{
						Name:        "instance-id",
						Value:       "",
						Usage:       "set instance id of server instance",
						Destination: &instanceId,
					},*/
				cli.StringFlag{
					Name:        "bind-address",
					Value:       "",
					Usage:       "bind listening address",
					Destination: &bindAddress,
				},
				cli.IntFlag{
					Name:        "port, p",
					Value:       0,
					Usage:       "server tcp port",
					Destination: &port,
				},
				cli.StringFlag{
					Name:        "advertise-address",
					Value:       "",
					Usage:       "advertise address is the broadcast address",
					Destination: &advertiseAddress,
				},
				cli.IntFlag{
					Name:        "advertise-port",
					Value:       0,
					Usage:       "advertise port is the broadcast port",
					Destination: &advertisePort,
				},
				cli.StringFlag{
					Name:        "data-dir",
					Value:       "",
					Usage:       "data directory",
					Destination: &dataDir,
				},
				cli.StringFlag{
					Name:        "preferred-network",
					Value:       "",
					Usage:       "choose preferred network interface for registering",
					Destination: &preferredNetwork,
				},
				cli.BoolTFlag{
					Name:        "enable-http",
					Usage:       "enable http server",
					Destination: &enableHttp,
				},
				cli.IntFlag{
					Name:        "http-port",
					Value:       0,
					Usage:       "http port",
					Destination: &httpPort,
				},
				cli.StringFlag{
					Name:        "http-auth",
					Value:       "",
					Usage:       "http authentication",
					Destination: &httpAuth,
				},
				cli.BoolFlag{
					Name:        "enable-mimetypes",
					Usage:       "enable http mime type",
					Destination: &enableMimetypes,
				},
				cli.StringFlag{
					Name:        "allowed-domains",
					Usage:       "allowed access domains",
					Destination: &allowedDomains,
				},
				cli.StringFlag{
					Name:  "trackers",
					Value: "",
					Usage: `set tracker servers, example:
	[<secret1>@]host1:port1,[<secret2>@]host2:port2`,
					Destination: &trackers,
				},
				cli.StringFlag{
					Name:        "logdir",
					Value:       "",
					Usage:       "set log directory",
					Destination: &logDir,
				},
				cli.BoolFlag{
					Name:        "disable-logfile",
					Usage:       "disable save log to file",
					Destination: &disableSaveLogfile,
				},
			},
		},
		{
			Name:  "client",
			Usage: "godfs client cli",
			Action: func(c *cli.Context) error {
				if len(c.Args()) == 0 {
					cli.ShowSubcommandHelp(c)
					os.Exit(0)
				}
				return nil
			},
			Subcommands: cli.Commands{
				{
					Name:  "upload",
					Usage: "upload local files",
					Action: func(c *cli.Context) error {
						finalCommand = UPLOAD_FILE
						if len(c.Args()) == 0 {
							return errors.New(`Err: no parameters provided.
Usage: godfs upload <file1> <file2> ...`)
						}
						/*workDir, err := file.GetWorkDir()
						if err != nil {
							logger.Fatal("error get current work directory: ", err)
						}
						absPath, err := filepath.Abs(workDir)
						if err != nil {
							logger.Fatal("error get absolute work directory: ", err)
						}*/
						for i := range c.Args() {
							if !util.StringListExists(&uploadFiles, c.Args().Get(i)) {
								uploadFiles.PushBack(c.Args().Get(i))
							}
						}
						return nil
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "group, g",
							Value:       "",
							Usage:       "upload files to specific group",
							Destination: &uploadGroup,
						},
						cli.BoolFlag{
							Name:        "private, p",
							Usage:       "mark as private files",
							Destination: &privateUpload,
						},
						cli.StringFlag{
							Name:        "config, c",
							Value:       "",
							Usage:       "use custom config file",
							Destination: &configFile,
						},
						cli.StringFlag{
							Name:  "storages",
							Value: "",
							Usage: `set storage servers, example:
	[<secret1>@]host1:port1,[<secret2>@]host2:port2`,
							Destination: &storages,
						},
						cli.StringFlag{
							Name:  "trackers",
							Value: "",
							Usage: `set tracker servers, example:
	[<secret1>@]host1:port1,[<secret2>@]host2:port2`,
							Destination: &trackers,
						},
					},
				},
				{
					Name:  "download",
					Usage: "download a file through tracker servers or storage servers",
					Action: func(c *cli.Context) error {
						finalCommand = DOWNLOAD_FILE
						if len(c.Args()) == 0 {
							return errors.New(`Err: no parameters provided.
Usage: godfs download <fid1> <fid2> ...`)
						}
						for i := range c.Args() {
							if !util.StringListExists(&downloadFiles, c.Args().Get(i)) {
								downloadFiles.PushBack(c.Args().Get(i))
							}
						}
						return nil
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name, n",
							Value: "",
							Usage: `custom download filename or full path of the
	download file(only valid for single file)`,
							Destination: &customDownloadFileName,
						},
						cli.StringFlag{
							Name:        "config, c",
							Value:       "",
							Usage:       "use custom config file",
							Destination: &configFile,
						},
						cli.StringFlag{
							Name:  "storages",
							Value: "",
							Usage: `set storage servers, example:
	[<secret1>@]host1:port1,[<secret2>@]host2:port2`,
							Destination: &storages,
						},
						cli.StringFlag{
							Name:  "trackers",
							Value: "",
							Usage: `set tracker servers, example:
	[<secret1>@]host1:port1,[<secret2>@]host2:port2`,
							Destination: &trackers,
						},
					},
				},
				{
					Name:  "inspect",
					Usage: "inspect infos of some files",
					Action: func(c *cli.Context) error {
						finalCommand = INSPECT_FILE
						if len(c.Args()) == 0 {
							return errors.New(`Err: no parameters provided.
Usage: godfs inspect <fid1> <fid2> ...`)
						}
						for i := range c.Args() {
							if !util.StringListExists(&inspectFiles, c.Args().Get(i)) {
								inspectFiles.PushBack(c.Args().Get(i))
							}
						}
						return nil
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "config, c",
							Value:       "",
							Usage:       "use custom config file",
							Destination: &configFile,
						},
						cli.StringFlag{
							Name:  "storages",
							Value: "",
							Usage: `set storage servers, example:
	[<secret1>@]host1:port1,[<secret2>@]host2:port2`,
							Destination: &storages,
						},
						cli.StringFlag{
							Name:  "trackers",
							Value: "",
							Usage: `set tracker servers, example:
	[<secret1>@]host1:port1,[<secret2>@]host2:port2`,
							Destination: &trackers,
						},
					},
				},
				{ // this sub command is only used by client cli
					Name:  "config",
					Usage: "config settings operations",
					Action: func(c *cli.Context) error {
						if len(c.Args()) == 0 {
							cli.ShowSubcommandHelp(c)
							os.Exit(0)
						}
						return nil
					},
					Subcommands: []cli.Command{
						{
							Name:  "set",
							Usage: "set configs in 'k=v' form",
							Action: func(c *cli.Context) error {
								finalCommand = UPDATE_CONFIG
								if len(c.Args()) == 0 {
									return errors.New(`Err: no parameters provided.
Usage: godfs config set key=value key=value ...`)
								}
								for i := range c.Args() {
									updateConfigList.PushBack(c.Args().Get(i))
								}
								return nil
							},
						},
						{
							Name:  "ls",
							Usage: "show configs",
							Action: func(c *cli.Context) error {
								finalCommand = SHOW_CONFIG
								return nil
							},
						},
					},
				},
			},
		},
	}

	cli.AppHelpTemplate = `
Usage: {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .VisibleCommands}}

Commands:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{end}}{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}

Options:
   {{range $index, $option := .VisibleFlags}}{{if $index}}{{end}}{{$option}}
   {{end}}{{end}}
`

	cli.CommandHelpTemplate = `
Usage: {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}

{{.Usage}}{{if .VisibleFlags}}

Options:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`

	cli.SubcommandHelpTemplate = `
Usage: {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} command{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}

{{if .Description}}{{.Description}}{{else}}{{.Usage}}{{end}}

Commands:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{end}}{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{if .VisibleFlags}}

Options:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`

	appFlag.Action = func(c *cli.Context) error {
		if showVersion {
			cli.ShowVersion(c)
			os.Exit(0)
			return nil
		}
		cli.ShowAppHelp(c)
		os.Exit(0)
		return nil
	}

	err := appFlag.Run(arguments)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return
	}

	if finalCommand == SHOW_HELP {
		os.Exit(0)
	}

	call(finalCommand)
}
