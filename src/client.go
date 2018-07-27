package main

import (
    "flag"
    "lib_client"
    "os"
    "path/filepath"
    "util/file"
    "util/logger"
    "validate"
    "app"
    "lib_storage"
    "lib_common"
)


var client *lib_client.Client


// 对于客户端，只提供类似于mysql的客户端，每个client与所有的tracker建立单个连接进行数据同步
// client和每个storage server最多建立一个连接
// 三方客户端可以开发成为一个连接池

func main() {

    abs, _ := filepath.Abs(os.Args[0])
    s, _ := filepath.Split(abs)
    s = file.FixPath(s)

    var uploadFile = flag.String("u", "", "the file to be uploaded")
    var downFile = flag.String("d", "", "the file to be download")
    var confPath = flag.String("c", s + string(filepath.Separator) + ".." + string(filepath.Separator) + "conf" + string(filepath.Separator) + "client.conf.template", "custom config file")
    flag.Parse()

    logger.Info("using config file:", *confPath)
    m, e := file.ReadPropFile(*confPath)
    if e == nil {
        validate.Check(m, 3)
        for k, v := range m {
            logger.Debug(k, "=", v)
        }
        app.RUN_WITH = 3
    } else {
        logger.Fatal("error read file:", e)
    }
}



