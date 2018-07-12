package header

import (
	"testing"
	"strconv"
	"util/logger"
	"fmt"
)

func Test1(t *testing.T) {
	a := ""

	n, e := strconv.Atoi(a)
	if e == nil {
		fmt.Println(n)
	} else {
		logger.Error("转换失败：", e)
	}
}
