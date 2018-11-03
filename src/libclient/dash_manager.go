package libclient

import (
    "libcommon/bridge"
    "util/logger"
)

var (
    managedStatistic = make(map[string][]bridge.ServerStatistic)
)

func updateStatistic(tracker string, statistic []bridge.ServerStatistic) {
    logger.Info("update statistic info:", statistic)
    managedStatistic[tracker] = statistic
}


