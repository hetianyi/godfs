package main

import (
    "util/file"
    "path/filepath"
    "fmt"
    "util/logger"
    "lib_tracker"
)

func main() {
    s, _ := file.GetWorkDir()
    m, e := file.ReadPropFile(s + string(filepath.Separator) + "conf" + string(filepath.Separator) + "tracker.conf.template")
    if e == nil {
        for k, v := range m {
            fmt.Println(k+"="+fmt.Sprint(v))
        }
    } else {
        logger.Fatal("error read file:", e)
    }
    lib_tracker.StartService(m)
}