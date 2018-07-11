package logger

import (
    "log"
    "os"
)

var (
    _trace   *log.Logger
    _debug   *log.Logger
    _info    *log.Logger
    _warn    *log.Logger
    _error   *log.Logger
    _fatal   *log.Logger
)


func init() {
    _trace = log.New(os.Stdout, "TRACE - ", log.LstdFlags)

    _debug = log.New(os.Stdout, "DEBUG - ", log.LstdFlags)
    _info  = log.New(os.Stdout, "INFO  - ", log.LstdFlags)
    _warn  = log.New(os.Stdout, "WARN  - ", log.LstdFlags)
    _error = log.New(os.Stdout, "ERROR - ", log.LstdFlags)
    _fatal = log.New(os.Stdout, "FATAL - ", log.LstdFlags)
}


func Trace(o ...interface{}) {
    _trace.Println(o)
}
func Debug(o ...interface{}) {
    _debug.Println(o)
}
func Info(o ...interface{}) {
    _info.Println(o)
}
func Warn(o ...interface{}) {
    _warn.Println(o)
}
func Error(o ...interface{}) {
    _error.Println(o)
}

// !!!this will cause system exists
func Fatal(o ...interface{}) {
    _fatal.Println(o)
    os.Exit(0)
}