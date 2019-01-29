package timeutil

import (
	"app"
	"bytes"
	"strconv"
	"sync"
	"time"
)

var (
	lock      = *new(sync.Mutex)
	increment = 0
)

// GetDateString 获取短日期格式：2018-11-11
func GetDateString(t time.Time) string {
	var buff bytes.Buffer
	buff.WriteString(strconv.Itoa(GetYear(t)))
	buff.WriteString("-")
	buff.WriteString(format2(GetMonth(t)))
	buff.WriteString("-")
	buff.WriteString(format2(GetDay(t)))
	return buff.String()
}

// GetLongDateString 获取长日期格式：2018-11-11 12:12:12
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

// GetShortDateString 获取长日期格式：2018-11-11 12:12:12
func GetShortDateString(t time.Time) string {
	var buff bytes.Buffer
	buff.WriteString(format2(GetHour(t)))
	buff.WriteString(":")
	buff.WriteString(format2(GetMinute(t)))
	buff.WriteString(":")
	buff.WriteString(format2(GetSecond(t)))
	return buff.String()
}

// GetLongLongDateString 获取长日期格式：2018-11-11 12:12:12,233
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

// GetLogFileName 获取日志文件格式
// till:
//      1: 到年，如2018
//      2: 到月，如2018-01
//      3: 到日，如2018-12-12
//      4: 到时，如2018-12-12_09
//      5: 到分，如2018-12-12_09_11
func GetLogFileName(t time.Time) string {
	var buff bytes.Buffer
	buff.WriteString(strconv.Itoa(GetYear(t)))
	if app.LogInterval == "y" {
		return buff.String()
	}
	buff.WriteString("-")
	buff.WriteString(format2(GetMonth(t)))
	if app.LogInterval == "m" {
		return buff.String()
	}
	buff.WriteString("-")
	buff.WriteString(format2(GetDay(t)))
	if app.LogInterval == "d" {
		return buff.String()
	}
	buff.WriteString("_")
	buff.WriteString(format2(GetHour(t)))
	if app.LogInterval == "h" {
		return buff.String()
	}
	buff.WriteString("_")
	buff.WriteString(format2(GetMinute(t)))
	return buff.String()
}

// GetTimestamp get current timestamp in milliseconds.
func GetTimestamp(t time.Time) int64 {
	return t.UnixNano() / 1e6
}

func CreateTime(millis int64) time.Time {
	return time.Unix(millis, 0)
}

// GetNanosecond get current timestamp in Nanosecond.
func GetNanosecond(t time.Time) int64 {
	return t.UnixNano()
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

func getGroupIncrement() int {
	lock.Lock()
	defer lock.Unlock()
	increment++
	if increment > 100000 {
		increment = 0
	}
	return increment
}

func GetUUID() string {
	return "tmp_" + strconv.FormatInt(GetNanosecond(time.Now()), 10) + "_" + strconv.Itoa(getGroupIncrement())
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

func GetHumanReadableDuration(start time.Time, end time.Time) string {
	v := GetTimestamp(end)/1000 - GetTimestamp(start)/1000 // seconds
	h := v / 3600
	m := v % 3600 / 60
	s := v % 60
	return format2(int(h)) + ":" + format2(int(m)) + ":" + format2(int(s))
}

func GetLongHumanReadableDuration(start time.Time, end time.Time) string {
	v := int(GetTimestamp(end)/1000 - GetTimestamp(start)/1000) // seconds
	return strconv.Itoa(v/86400) + "d " + strconv.Itoa(v%86400/3600) + "h " + strconv.Itoa(v%3600/60) + "m " + strconv.Itoa(v%60) + "s"
}
