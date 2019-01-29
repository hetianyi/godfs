package logger

import (
	"app"
	"os"
	"strconv"
	"sync"
	"time"
	"util/file"
	"util/timeutil"
)

var (
	enable        = false
	logLevel      = 2
	out           *os.File
	lastLogTime   *time.Time
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

	yearLast, monthLast, dayLast := lastLogTime.Date()
	hourLast := timeutil.GetHour(*lastLogTime)

	yearNow, monthNow, dayNow := now.Date()
	hourNow := timeutil.GetHour(now)

	if app.LogInterval == "d" {
		if yearLast != yearNow || int(monthLast) != int(monthNow) || dayLast != dayNow {
			resetLogFile(now)
		}
	} else if app.LogInterval == "h" {
		if yearLast != yearNow || int(monthLast) != int(monthNow) || dayLast != dayNow || hourLast != hourNow {
			resetLogFile(now)
		}
	} else if app.LogInterval == "m" {
		if yearLast != yearNow || int(monthLast) != int(monthNow) {
			resetLogFile(now)
		}
	} else if app.LogInterval == "y" {
		if yearLast != yearNow {
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
	if app.RunWith == 1 {
		logWith = "storage-"
	} else if app.RunWith == 2 {
		logWith = "tracker-"
	} else if app.RunWith == 3 {
		logWith = "client-"
	}
	logFileName := app.BasePath + string(os.PathSeparator) + "logs" + string(os.PathSeparator) +
		logFilePrefix + logWith + timeutil.GetLogFileName(now) + ".log"
	index := 0
	for {
		index++
		// exist file is a directory, rename to another.
		if file.Exists(logFileName) && file.IsDir(logFileName) {
			logFileName = app.BasePath + string(os.PathSeparator) + "logs" + string(os.PathSeparator) +
				logFilePrefix + timeutil.GetLogFileName(now) + "(" + strconv.Itoa(index) + ").log"
			continue
		}
		if !file.Exists(logFileName) || (file.Exists(logFileName) && file.IsFile(logFileName)) {
			tmp, e1 := file.OpenFile4Write(logFileName)
			if e1 == nil {
				out = tmp
				break
			} else {
				if index > 10 {
					Fatal("failed create log file:", e1)
				}
			}
		}
	}
}

func SetLogLevel(level int) {
	logLevel = level
	app.LogLevel = logLevel
}

// SetEnable enable log
func SetEnable(e bool) {
	now := time.Now()
	resetLogFile(now)
	enable = e
}
