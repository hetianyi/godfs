package logger

import (
    "log"
    "os"
    "fmt"
    "sync"
)

var (
    _trace   *log.Logger    // 0
    _debug   *log.Logger    // 1
    _info    *log.Logger    // 2
    _warn    *log.Logger    // 3
    _error   *log.Logger    // 4
    _fatal   *log.Logger    // 5
    increRollSize sync.Mutex
)


func init() {
    //fmt.Println("初始化Logger//////")
    _trace = log.New(os.Stdout, "TRACE - ", log.LstdFlags)
    _debug = log.New(os.Stdout, "DEBUG - ", log.LstdFlags)
    _info  = log.New(os.Stdout, "INFO  - ", log.LstdFlags)
    _warn  = log.New(os.Stdout, "WARN  - ", log.LstdFlags)
    _error = log.New(os.Stdout, "ERROR - ", log.LstdFlags)
    _fatal = log.New(os.Stdout, "FATAL - ", log.LstdFlags)
}

func doIncreRollSize(size int) {
    increRollSize.Lock()
    defer increRollSize.Unlock()
    if nil != trigger {
        trigger.check()
    }

}


func Trace(o ...interface{}) {
    if logLevel  > 0 {
        return
    }
    line := fmt.Sprintln(o)
    _trace.Println(line)
    if out != nil {
        doIncreRollSize(len([]byte(line)))
        out.Write([]byte(line))
    }
}

func Debug(o ...interface{}) {
    if logLevel  > 1 {
        return
    }
    line := fmt.Sprintln(o)
    _debug.Println(line)
    if out != nil {
        doIncreRollSize(len([]byte(line)))
        out.Write([]byte(line))
    }
}

func Info(o ...interface{}) {
    if logLevel  > 2 {
        return
    }
    line := fmt.Sprintln(o)
    _info.Print(line)
    if out != nil {
        doIncreRollSize(len([]byte(line)))
        out.Write([]byte(line))
    }
}

func Warn(o ...interface{}) {
    if logLevel  > 3 {
        return
    }
    line := fmt.Sprintln(o)
    _warn.Println(line)
    if out != nil {
        out.Write([]byte(line))
        doIncreRollSize(len([]byte(line)))
    }
}

func Error(o ...interface{}) {
    if logLevel  > 4 {
        return
    }
    line := fmt.Sprintln(o)
    _error.Println(line)
    if out != nil {
        doIncreRollSize(len([]byte(line)))
        out.Write([]byte(line))
    }
}

// !!!this will cause system exists
func Fatal(o ...interface{}) {
    if logLevel  > 5 {
        return
    }
    line := fmt.Sprintln(o)
    _fatal.Println(line)
    if out != nil {
        doIncreRollSize(len([]byte(line)))
        out.Write([]byte(line))
    }
    os.Exit(0)
}


