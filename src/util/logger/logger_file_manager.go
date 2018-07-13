package logger

import (
    "os"
    "time"
    "util/timeutil"
    "lib_common"
    "util/file"
)

var (
    enable = false
    logLevel = 2
    out *os.File
    trigger *Trigger //当前轮转日志达到最大时的触发器
    lastLogTime *time.Time
    logFilePrefix = "godfsLog-"
)

type ITrigger interface {
    check(rollSize int64)
}

type Trigger struct {
}

func (trigger *Trigger) check() {
    if !enable {
        return
    }
    now := time.Now()
    if lastLogTime == nil {
        lastLogTime = &now
    }

    year_last, month_last, day_last := lastLogTime.Date()
    hour_last := timeutil.GetHour(*lastLogTime)

    year_now, month_now, day_now := now.Date()
    hour_now := timeutil.GetHour(now)

    if lib_common.LOG_INTERVAL == "d" {
        if year_last != year_now || int(month_last) != int(month_now) || day_last != day_now {
            logFileName := logFilePrefix + timeutil.GetLogFileName(now) + ".log"
            file.CreateFile(lib_common.BASE_PATH)
        }
    } else if lib_common.LOG_INTERVAL == "h" {

    } else if lib_common.LOG_INTERVAL == "m" {

    } else if lib_common.LOG_INTERVAL == "y" {

    }


}




// enable write into log file.
// till: 参考 timeutil.GetLogFileName()
func EnableLogFile(_maxRollSize int64, _logLevel int, till int) {
    enable = true
    maxRollSize = _maxRollSize
    if _logLevel < 0 || _logLevel > 5 {
        _logLevel = 2
    }
    logLevel = _logLevel

}


func SetLogLevel(level int) {
    logLevel = level
}

