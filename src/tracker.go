package main

import (
    "util/file"
    "path/filepath"
    "util/logger"
    "app"
    "os"
    "flag"
    "validate"
    "lib_tracker"
    "runtime"
)

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())
    abs, _ := filepath.Abs(os.Args[0])
    s, _ := filepath.Split(abs)
    s = file.FixPath(s)
    var confPath = flag.String("c", s + string(filepath.Separator) + ".." + string(filepath.Separator) + "conf" + string(filepath.Separator) + "tracker.conf", "custom config file")
    flag.Parse()
    logger.Info("using config file:", *confPath)
    m, e := file.ReadPropFile(*confPath)
    if e == nil {
        validate.Check(m, 2)
        for k, v := range m {
            logger.Debug(k, "=", v)
        }
        app.RUN_WITH = 2
        lib_tracker.StartService(m)
    } else {
        logger.Fatal("error read file:", e)
    }
}