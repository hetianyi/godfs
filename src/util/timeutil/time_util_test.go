package timeutil

import (
    "testing"
    "fmt"
    "time"
    "strings"
    "container/list"
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
    fmt.Println(GetLongLongDateString(tm))
}
func Test3(t *testing.T) {
    tm := time.Now()
    fmt.Println(GetLogFileName(tm))
    fmt.Println(GetLogFileName(tm))
    fmt.Println(GetLogFileName(tm))
    fmt.Println(GetLogFileName(tm))
    fmt.Println(GetLogFileName(tm))
    fmt.Println(GetLogFileName(tm))
    fmt.Println(GetLogFileName(tm))
    fmt.Println(GetLogFileName(tm))
    fmt.Println(string('z'))
    fmt.Println(strings.ToUpper("az2597de0e65eedbc867620e744a35ad"[0:2]))
}


func Test4(t *testing.T) {
    var ls list.List
    ls.PushBack("1")
    ls.PushBack("2")
    ls.PushBack("3")
    for ele := ls.Front(); ele != nil; {
        next := ele.Next()
        ls.Remove(ele)
        ele = next
    }
    fmt.Println(ls.Len())
}

func Test5(t *testing.T) {
    start := time.Now()
    time.Sleep(time.Second * 5)
    fmt.Println(GetHumanReadableDuration(start, time.Now()))
}