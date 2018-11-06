package libclient

import (
    "libcommon/bridge"
    "util/logger"
    "encoding/json"
)

var (
    managedStatistic = make(map[string][]bridge.ServerStatistic)
)

func updateStatistic(tracker string, statistic []bridge.ServerStatistic) {
    ret, _ := json.Marshal(statistic)
    logger.Info("update statistic info:( ", string(ret), ")")
    managedStatistic[tracker] = statistic
}


