package main

import (
    "app"
    "libdashboard"
    "path/filepath"
    "os"
    "util/file"
    "flag"
    "util/logger"
    "validate"
)

func main() {
    // set client type
    app.RUN_WITH = 4
    app.CLIENT_TYPE = 3
    app.UUID = "DASHBOARD-CLIENT"

    abs, _ := filepath.Abs(os.Args[0])
    s, _ := filepath.Split(abs)
    s = file.FixPath(s)
    var confPath = flag.String("c", s+string(filepath.Separator)+".."+string(filepath.Separator)+"conf"+string(filepath.Separator)+"dashboard.conf", "custom config file")
    flag.Parse()
    logger.Info("using config file:", *confPath)
    m, e := file.ReadPropFile(*confPath)
    if e == nil {
        validate.Check(m, app.RUN_WITH)
        for k, v := range m {
            logger.Debug(k, "=", v)
        }
        libdashboard.StartService(m)
    } else {
        logger.Fatal("error read file:", e)
    }

}


