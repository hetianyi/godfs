package common

import "container/list"

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