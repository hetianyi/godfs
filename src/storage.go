package main

import (
    "util/file"
    "path/filepath"
    "util/logger"
    "lib_storage"
    "validate"
    "flag"
    "fmt"
    "os"
)

// TODO 文件保存分为group和Initial memeber ID(文件初始成员ID)
// 当客户端下载文件的时候，如果文件尚未在组内全部同步完成，
// 并且恰好访问到没有同步完成的机器时，客户端会将请求重定向到文件原始服务器
// exp: /G001(组)/M01(原始服务器)/{MD5}
func main() {
    fmt.Println(os.Args)
    s, _ := file.GetWorkDir()
    fmt.Println(s)
    s = file.FixPath(s)
    var confPath = flag.String("c", s + string(filepath.Separator) + ".." + string(filepath.Separator) + "conf" + string(filepath.Separator) + "storage.conf.template", "custom config file")
    flag.Parse()

    m, e := file.ReadPropFile(*confPath)
    if e == nil {
        validate.Check(m, 1)
        for k, v := range m {
            logger.Debug(k, "=", v)
        }
        lib_storage.StartService(m)
    } else {
        logger.Fatal("error read file:", e)
    }
}
