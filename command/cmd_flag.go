package command

import (
	"errors"
	"fmt"
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/godfs/util"
	"github.com/urfave/cli"
	"os"
)

// Parse parses command flags using `github.com/urfave/cli`
func Parse(arguments []string) {
	appFlag := cli.NewApp()
	appFlag.Version = common.VERSION
	appFlag.HideVersion = true
	appFlag.Name = "godfs"
	appFlag.Usage = "godfs"
	appFlag.HelpName = "godfs"
	// config file location
	appFlag.Flags = []cli.Flag{
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
				finalCommand = common.CMD_BOOT_TRACKER
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "log-level",
					Value: "",
					Usage: `set log level, available options:
	(trace|debug|info|warn|error|fatal)`,
					Destination: &logLevel,
				},
				cli.StringFlag{
					Name:        "secret, s",
					Value:       "",
					Usage:       "custom global secret",
					Destination: &secret,
				},
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
				cli.BoolFlag{
					Name:        "disable-http",
					Usage:       "disable http server",
					Destination: &disableHttp,
				},
				cli.IntFlag{
					Name:        "http-port",
					Value:       0,
					Usage:       "http port",
					Destination: &httpPort,
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
					Name:        "log-dir",
					Value:       "",
					Usage:       "set log directory",
					Destination: &logDir,
				},
				cli.IntFlag{
					Name:  "max-logfile-size",
					Value: 0,
					Usage: `rolling log file max size, options:
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
				finalCommand = common.CMD_BOOT_STORAGE
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "log-level",
					Value: "",
					Usage: `set log level, available options:
	(trace|debug|info|warn|error|fatal)`,
					Destination: &logLevel,
				},
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
					Name:  "access-mode, m",
					Value: "private",
					Usage: `the access mode used by uploaded files by default,
	options:(private|public)`,
					Destination: &defaultAccessMode,
				},
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
				cli.BoolFlag{
					Name:        "disable-http",
					Usage:       "disable http server",
					Destination: &disableHttp,
				},
				cli.IntFlag{
					Name:        "http-port",
					Value:       0,
					Usage:       "http port",
					Destination: &httpPort,
				},
				cli.BoolTFlag{
					Name:        "enable-mimetypes",
					Usage:       "enable http mime type",
					Destination: &enableMimetypes,
				},
				cli.StringFlag{
					Name:        "allowed-hosts",
					Usage:       "allowed access hosts",
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
					Name:        "log-dir",
					Value:       "",
					Usage:       "set log directory",
					Destination: &logDir,
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
						finalCommand = common.CMD_UPLOAD_FILE
						if len(c.Args()) == 0 {
							return errors.New(`Err: no parameters provided.
Usage: godfs client upload <file1> <file2> ...`)
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
							Name:        "public, p",
							Usage:       "mark as public files",
							Destination: &publicUpload,
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
						cli.StringFlag{
							Name:  "log-level",
							Value: "",
							Usage: `set log level, available options:
	(trace|debug|info|warn|error|fatal)`,
							Destination: &logLevel,
						},
					},
				},
				{
					Name:  "download",
					Usage: "download a file through tracker servers or storage servers",
					Action: func(c *cli.Context) error {
						finalCommand = common.CMD_DOWNLOAD_FILE
						if len(c.Args()) == 0 {
							return errors.New(`Err: no parameters provided.
Usage: godfs client download <fid1> <fid2> ...`)
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
						cli.StringFlag{
							Name:  "log-level",
							Value: "",
							Usage: `set log level, available options:
	(trace|debug|info|warn|error|fatal)`,
							Destination: &logLevel,
						},
					},
				},
				{
					Name:  "inspect",
					Usage: "inspect infos of some files",
					Action: func(c *cli.Context) error {
						finalCommand = common.CMD_INSPECT_FILE
						if len(c.Args()) == 0 {
							return errors.New(`Err: no parameters provided.
Usage: godfs client inspect <fid1> <fid2> ...`)
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
						cli.StringFlag{
							Name:  "log-level",
							Value: "",
							Usage: `set log level, available options:
	(trace|debug|info|warn|error|fatal)`,
							Destination: &logLevel,
						},
					},
				},
				{
					Name:  "token",
					Usage: "generate file access token",
					Action: func(c *cli.Context) error {
						finalCommand = common.CMD_GENERATE_TOKEN
						if len(c.Args()) == 0 {
							return errors.New(`Err: no parameters provided.
Usage: godfs client token <fid1> <fid2> ...`)
						}
						tokenFileId = c.Args()[0]
						return nil
					},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:        "secret, s",
							Value:       "",
							Usage:       "secret used for generating token",
							Destination: &secret,
						},
						cli.IntFlag{
							Name:        "life, l",
							Value:       3600,
							Usage:       "token life(in seconds)",
							Destination: &tokenLife,
						},
						cli.StringFlag{
							Name:        "format, f",
							Value:       "url",
							Usage:       "token format:json|url",
							Destination: &tokenFormat,
						},
					},
				},
				{
					Name:  "test",
					Usage: "running benchmark",
					Action: func(c *cli.Context) error {
						finalCommand = common.CMD_TEST_UPLOAD
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
							Name:        "public, p",
							Usage:       "mark as public files",
							Destination: &publicUpload,
						},
						cli.IntFlag{
							Name:        "scale, s",
							Usage:       "test scale",
							Value:       10000,
							Destination: &testScale,
						},
						cli.IntFlag{
							Name:        "thread, t",
							Usage:       "test thread size",
							Value:       5,
							Destination: &testThread,
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
						cli.StringFlag{
							Name:  "log-level",
							Value: "",
							Usage: `set log level, available options:
	(trace|debug|info|warn|error|fatal)`,
							Destination: &logLevel,
						},
					},
				},
			},
		},
	}

	cli.AppHelpTemplate = `
Usage: {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .VisibleCommands}}

Commands:{{range .VisibleCategories}}
{{if .Name}}
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

Commands:
{{range .VisibleCategories}}{{if .Name}}
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

	if finalCommand == common.CMD_SHOW_HELP {
		os.Exit(0)
	}

	call(finalCommand)
}
