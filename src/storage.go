package main

import (
    "util/file"
    "fmt"
    "path/filepath"
    "util/logger"
    "lib_storage"
    "validate"
)

func main() {
    s, _ := file.GetWorkDir()
    m, e := file.ReadPropFile(s + string(filepath.Separator) + "conf" + string(filepath.Separator) + "storage.conf.template")
    validate.Check(m, 1)
    if e == nil {
        for k, v := range m {
            fmt.Println(k+"="+fmt.Sprint(v))
        }
    } else {
        logger.Fatal("error read file:", e)
    }

    lib_storage.StartService(m)
}