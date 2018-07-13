package timeutil

import (
    "testing"
    "fmt"
    "time"
)

func Test1(t *testing.T) {
    tm := time.Now()
    fmt.Println(GetTimestamp(tm))
    fmt.Println(GetYear(tm))
    fmt.Println(GetMonth(tm))
    fmt.Println(GetDay(tm))
    fmt.Println(GetHour(tm))
    fmt.Println(GetMinute(tm))
    fmt.Println(GetSecond(tm))
}

func Test2(t *testing.T) {
    tm := time.Now()
    fmt.Println(GetDateString(tm))
    fmt.Println(GetLongDateString(tm))
}
func Test3(t *testing.T) {
    tm := time.Now()
    fmt.Println(GetLogFileName(tm, 1))
    fmt.Println(GetLogFileName(tm, 2))
    fmt.Println(GetLogFileName(tm, 3))
    fmt.Println(GetLogFileName(tm, 4))
    fmt.Println(GetLogFileName(tm, 5))
    fmt.Println(GetLogFileName(tm, 6))
    fmt.Println(GetLogFileName(tm, 7))
    fmt.Println(GetLogFileName(tm, 0))
}
