package main

import (
    "util/file"
    "path/filepath"
    "util/logger"
    "lib_storage"
    "validate"
)

func main() {
    s, _ := file.GetWorkDir()
    m, e := file.ReadPropFile(s + string(filepath.Separator) + "conf" + string(filepath.Separator) + "storage.conf.template")
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
