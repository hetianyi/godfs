package util

import (
	"container/list"
	"github.com/hetianyi/gox"
)

// StringListExists checks if a string list contains the string.
func StringListExists(list *list.List, ele string) bool {
	exists := false
	gox.WalkList(list, func(item interface{}) bool {
		if item.(string) == ele {
			exists = true
			return true
		}
		return false
	})
	return exists
}
