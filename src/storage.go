package main

import (
	"app"
	"github.com/urfave/cli"
	"libclient"
	"libstorage"
	"os"
	"path/filepath"
	"runtime"
	"util/file"
	"util/logger"
	"validate"
)

// 当客户端下载文件的时候，如果文件尚未在组内全部同步完成，
// 并且恰好访问到没有同步完成的机器时，客户端会将请求重定向到文件原始服务器
// exp: /G001(组)/01(原始服务器实例ID)/M[S](单片or多片)/{MD5}[.ext]
// 文件的原始名称需要客户端自行记录（可能未来加上服务端记录功能）
// TODO support detect total file size for assigning
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	app.RunWith = 1
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
		validate.Check(m, app.RunWith)
		for k, v := range m {
			logger.Debug(k, "=", v)
		}
		libstorage.StartService(m)
	} else {
		logger.Fatal("error read file:", e)
	}
}


func initStorageFlags() {
	appFlag := cli.NewApp()
	appFlag.Version = app.Version
	appFlag.Name = "godfs storage"
	appFlag.Usage = ""

	// config file location
	appFlag.Flags = []cli.Flag {
		cli.StringFlag{
			Name:        "config, c",
			Value:       "../conf/storage.conf",
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
