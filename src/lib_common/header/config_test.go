package header

import (
	"testing"
	"strconv"
	"util/logger"
	"fmt"
	"regexp"
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


func Test2(t *testing.T) {
	fmt.Println(regexp.Match("^[0-9a-zA-Z_]+$", []byte("001-")))
}


type A struct {
	Id int
}

type B struct {
	A
	name string
}


func Test3(t *testing.T) {
	m := make(map[int]string)
	m[2] = "xxxx"
	fmt.Println(m[1])
	fmt.Println(m[2])
	var b = &B{
	}
	fmt.Println(b.Id)


	operationHeadMap := make(map[int][]byte)
	operationHeadMap[1] = []byte{1,1}
	fmt.Println(operationHeadMap[2]==nil)

}