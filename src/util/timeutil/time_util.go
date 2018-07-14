package timeutil

import (
    "time"
    "strconv"
    "bytes"
    "lib_common"
)


// 获取短日期格式：2018-11-11
func GetDateString(t time.Time) string {
    var buff bytes.Buffer
    buff.WriteString(strconv.Itoa(GetYear(t)))
    buff.WriteString("-")
    buff.WriteString(format2(GetMonth(t)))
    buff.WriteString("-")
    buff.WriteString(format2(GetDay(t)))
    return buff.String()
}
// 获取长日期格式：2018-11-11 12:12:12
func GetLongDateString(t time.Time) string {
    var buff bytes.Buffer
    buff.WriteString(strconv.Itoa(GetYear(t)))
    buff.WriteString("-")
    buff.WriteString(format2(GetMonth(t)))
    buff.WriteString("-")
    buff.WriteString(format2(GetDay(t)))
    buff.WriteString(" ")
    buff.WriteString(format2(GetHour(t)))
    buff.WriteString(":")
    buff.WriteString(format2(GetMinute(t)))
    buff.WriteString(":")
    buff.WriteString(format2(GetSecond(t)))
    return buff.String()
}
// 获取长日期格式：2018-11-11 12:12:12,233
func GetLongLongDateString(t time.Time) string {
    var buff bytes.Buffer
    buff.WriteString(strconv.Itoa(GetYear(t)))
    buff.WriteString("-")
    buff.WriteString(format2(GetMonth(t)))
    buff.WriteString("-")
    buff.WriteString(format2(GetDay(t)))
    buff.WriteString(" ")
    buff.WriteString(format2(GetHour(t)))
    buff.WriteString(":")
    buff.WriteString(format2(GetMinute(t)))
    buff.WriteString(":")
    buff.WriteString(format2(GetSecond(t)))
    buff.WriteString(",")
    buff.WriteString(format3(GetMillionSecond(t)))
    return buff.String()
}

// 获取日志文件格式
// till:
//      1: 到年，如2018
//      2: 到月，如2018-01
//      3: 到日，如2018-12-12
//      4: 到时，如2018-12-12_09
//      5: 到分，如2018-12-12_09_11
func GetLogFileName(t time.Time) string {
    var buff bytes.Buffer
    buff.WriteString(strconv.Itoa(GetYear(t)))
    if lib_common.LOG_INTERVAL == "y" {
        return buff.String()
    }
    buff.WriteString("-")
    buff.WriteString(format2(GetMonth(t)))
    if lib_common.LOG_INTERVAL == "m" {
        return buff.String()
    }
    buff.WriteString("-")
    buff.WriteString(format2(GetDay(t)))
    if lib_common.LOG_INTERVAL == "d" {
        return buff.String()
    }
    buff.WriteString("_")
    buff.WriteString(format2(GetHour(t)))
    if lib_common.LOG_INTERVAL == "h" {
        return buff.String()
    }
    buff.WriteString("_")
    buff.WriteString(format2(GetMinute(t)))
    return buff.String()
}


// get current timestamp in milliseconds.
func GetTimestamp(t time.Time) int64 {
    return t.UnixNano() / 1e6
}

func GetYear(t time.Time) int {
    return t.Year()
}
func GetMonth(t time.Time) int {
    return int(t.Month())
}
func GetDay(t time.Time) int {
    return t.Day()
}
func GetHour(t time.Time) int {
    return t.Hour()
}

func GetMinute(t time.Time) int {
    return t.Minute()
}
func GetSecond(t time.Time) int {
    return t.Second()
}
func GetMillionSecond(t time.Time) int {
    return t.Nanosecond() / 1e6
}

func format2(input int) string {
    if input < 10 {
        return "0" + strconv.Itoa(input)
    }
    return strconv.Itoa(input)
}
func format3(input int) string {
    if input < 10 {
        return "00" + strconv.Itoa(input)
    }
    if input < 100 {
        return "0" + strconv.Itoa(input)
    }
    return strconv.Itoa(input)
}

