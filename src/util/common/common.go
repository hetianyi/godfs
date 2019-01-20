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

// ternary operation
func TOperation(condition bool, trueOperation func() interface{}, falseOperation func() interface{}) interface{}{
	if condition {
		if trueOperation == nil {
			return nil
		}
		return trueOperation()
	}
	if falseOperation == nil {
		return nil
	}
	return falseOperation()
}

// ternary operation
func TValue(condition bool, trueValue interface{}, falseValue interface{}) interface{}{
	if condition {
		return trueValue
	}
	return falseValue
}

// walk a list
// walker return value as break signal
// if it is true, break walking
func WalkList(ls *list.List, walker func(item interface{}) bool) {
	if ls == nil {
		return
	}
	for ele := ls.Front(); ele != nil; ele = ele.Next() {
		breakWalk := walker(ele.Value)
		if breakWalk {
			break
		}
	}
}