package command

import (
	"github.com/hetianyi/godfs/common"
	"github.com/hetianyi/gox/logger"
	"github.com/urfave/cli"
)

func Parse(arguments []string) {

	appFlag := cli.NewApp()
	appFlag.Version = common.VERSION
	appFlag.HideVersion = true
	appFlag.Name = "godfs"
	appFlag.Usage = "godfs"

	// config file location
	appFlag.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "tracker",
			Usage:       "boot as tracker server",
			Destination: &trackerModel,
		},
		cli.BoolFlag{
			Name:        "storage",
			Usage:       "boot as storage server",
			Destination: &storageModel,
		},
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
		},
		cli.StringFlag{
			Name:  "logLevel, ll",
			Value: "",
			Usage: `set log level
	available options(trace|debug|info|warn|error|fatal)
`,
			Destination: &configFile,
		},
		cli.StringFlag{
			Name:  "trackerServers, ts",
			Value: "",
			Usage: `set tracker servers, example:
	[<secret1>@]host1:port1,[<secret2>@]host2:port2
`,
			Destination: &trackers,
		},
		cli.StringFlag{
			Name:  "storageServers, ss",
			Value: "",
			Usage: `set storage servers, example:
	[<secret1>@]host1:port1,[<secret2>@]host2:port2
`,
			Destination: &storages,
		},
	}

	// sub command 'upload'
	appFlag.Commands = []cli.Command{
		{
			Name:  "upload",
			Usage: "upload local files",
			Action: func(c *cli.Context) error {
				finalCommand = UPLOAD_FILE
				/*workDir, err := file.GetWorkDir()
				if err != nil {
					logger.Fatal("error get current work directory: ", err)
				}
				absPath, err := filepath.Abs(workDir)
				if err != nil {
					logger.Fatal("error get absolute work directory: ", err)
				}*/
				for i := range c.Args() {
					uploadFiles.PushBack(c.Args().Get(i))
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
			},
		},
		{
			Name:  "download",
			Usage: "download a file through tracker servers or storage servers",
			Action: func(c *cli.Context) error {
				finalCommand = DOWNLOAD_FILE
				for i := range c.Args() {
					uploadFiles.PushBack(c.Args().Get(i))
				}
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "name, n",
					Value:       "",
					Usage:       "custom filename of the download file",
					Destination: &customDownloadFileName,
				},
			},
		},
		{
			Name:  "inspect",
			Usage: "inspect infos of some files",
			Action: func(c *cli.Context) error {
				finalCommand = INSPECT_FILE
				for i := range c.Args() {
					inspectFiles.PushBack(c.Args().Get(i))
				}
				return nil
			},
		},
		{ // this sub command is only used by client cli
			Name:  "config",
			Usage: "config settings command",
			Action: func(c *cli.Context) error {
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:  "set",
					Usage: "set configs in 'k=v' form",
					Action: func(c *cli.Context) error {
						finalCommand = UPDATE_CONFIG
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
	}
	cli.AppHelpTemplate = `
{{if .Usage}} - {{.Usage}}{{end}}Usage: {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .Version}}{{if not .HideVersion}}
Version: {{.Version}}{{end}}{{end}}{{if .Description}}
Description: {{.Description}}{{end}}{{if len .Authors}}
Author{{with $length := len .Authors}}{{if ne 1 $length}}S{{end}}{{end}}:
   {{range $index, $author := .Authors}}{{if $index}}
   {{end}}{{$author}}{{end}}{{end}}{{if .VisibleCommands}}

Commands: {{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{end}}{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}

Options:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}{{if .Copyright}}
Copyright:
   {{.Copyright}}{{end}}

`

	cli.CommandHelpTemplate = `
Usage:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .Category}}

{{.Usage}}

Category:
   {{.Category}}{{end}}{{if .VisibleFlags}}

Options:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`

	cli.SubcommandHelpTemplate = `
Usage: {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} command{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}

{{if .Description}}{{.Description}}{{else}}{{.Usage}}{{end}}

Commands:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{end}}{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}
{{end}}{{if .VisibleFlags}}
Options:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`

	err := appFlag.Run(arguments)
	if err != nil {
		logger.Fatal(err)
	}
}
