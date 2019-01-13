package common

import (
	"container/list"
	"strings"
	"strconv"
)

func ConvertBoolFromInt(input int) bool {
	if input <= 0 {
		return false
	}
	return true
}

// convert list to array
func List2Array(ls *list.List) []interface{} {
	if ls == nil {
		return nil
	}
	arr := make([]interface{}, ls.Len())
	index := 0
	for ele := ls.Front(); ele != nil; ele = ele.Next() {
		arr[index] = ele.Value
		index++
	}
	return arr
}

// parse host and port from connection string
func ParseHostPortFromConnStr(connStr string) (string, int) {
	host := strings.Split(connStr, ":")[0]
	port, _ := strconv.Atoi(strings.Split(connStr, ":")[1])
	return host, port
}
