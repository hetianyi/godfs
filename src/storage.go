package main

import (
    "util/file"
    "path/filepath"
    "util/logger"
    "lib_storage"
    "validate"
    "flag"
    "fmt"
)

func main() {
    s, _ := file.GetWorkDir()
    s = file.FixPath(s)
    var confPath = flag.String("c", s + string(filepath.Separator) + ".." + string(filepath.Separator) + "conf" + string(filepath.Separator) + "storage.conf.template", "custom config file")

    flag.Parse()
    fmt.Println("confPath ", *confPath)

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
