package lib_storage

import (
    "time"
    "app"
    "util/logger"
)

func startSyncSerivce() {
    timer := time.NewTicker(time.Second * app.SYNC_INTERVAL)
    for {
        <-timer.C
        logger.Debug("fetch sync tasks")

    }
}
