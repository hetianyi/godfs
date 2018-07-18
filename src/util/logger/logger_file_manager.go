package logger

import (
    "os"
    "time"
    "util/timeutil"
    "util/file"
    "strconv"
    "sync"
    "app"
)

var (
    enable = false
    logLevel = 2
    out *os.File
    lastLogTime *time.Time
    logFilePrefix = "godfsLog-"
    increRollSize sync.Mutex
)


func check() {
    increRollSize.Lock()
    defer increRollSize.Unlock()

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

    if app.LOG_INTERVAL == "d" {
        if year_last != year_now || int(month_last) != int(month_now) || day_last != day_now {
            resetLogFile(now)
        }
    } else if app.LOG_INTERVAL == "h" {
        if year_last != year_now || int(month_last) != int(month_now) || day_last != day_now || hour_last != hour_now {
            resetLogFile(now)
        }
    } else if app.LOG_INTERVAL == "m" {
        if year_last != year_now || int(month_last) != int(month_now) {
            resetLogFile(now)
        }
    } else if app.LOG_INTERVAL == "y" {
        if year_last != year_now {
            resetLogFile(now)
        }
    }
}

func closeLogFile() {
    if out != nil {
        out.Close()
    }
}

func resetLogFile(now time.Time) {
    closeLogFile()
    logWith := ""
    if app.RUN_WITH == 1 {
        logWith = "storage-"
    } else {
        logWith = "tracker-"
    }
    logFileName := app.BASE_PATH + string(os.PathSeparator) + "logs" + string(os.PathSeparator) +
                    logFilePrefix + logWith + timeutil.GetLogFileName(now) + ".log"
    index := 0
    for {
        index++
        // exist file is a directory, rename to another.
        if file.Exists(logFileName) && file.IsDir(logFileName) {
            logFileName = app.BASE_PATH + string(os.PathSeparator) + "logs" + string(os.PathSeparator) +
                    logFilePrefix + timeutil.GetLogFileName(now) + "(" + strconv.Itoa(index) + ").log"
            continue
        }
        if !file.Exists(logFileName) || (file.Exists(logFileName) && file.IsFile(logFileName)) {
            tmp, e1 := file.OpenFile4Write(logFileName)
            if e1 == nil {
                out = tmp
                break
            } else {
                if index > 10  {
                    Fatal("failed create log file:", e1)
                }
            }
        }
    }
}

func SetLogLevel(level int) {
    logLevel = level
}

func SetEnable(e bool) {
    now := time.Now()
    resetLogFile(now)
    enable = e
}

