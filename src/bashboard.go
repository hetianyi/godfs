package main

import (
    "app"
    "libdashboard"
    "util/logger"
)

func main() {
    app.SECRET = "OASAD834jA97AAQE761=="
    logger.SetLogLevel(1)
    libdashboard.StartService(nil)

}


