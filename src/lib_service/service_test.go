package lib_service

import (
    "testing"
    "fmt"
    "app"
    "util/logger"
    "container/list"
    "strconv"
)

func initParam() {
    app.BASE_PATH = "E:/WorkSpace2018/godfs"
    logger.SetLogLevel(1)
}

func Test1(t *testing.T) {
    initParam()
    fmt.Println(GetFileId("asd1231"))
}
func Test2(t *testing.T) {
    initParam()
    fmt.Println(GetPartId("0d3cc782c3242cf3ce4b2174e1041ed2"))
}

func Test3(t *testing.T) {
    initParam()
    fmt.Println(AddPart("0d3cc782c3242cf3ce4b2174e1041ed233"))
}
func Test4(t *testing.T) {
    initParam()

    var ls = *new(list.List)
    for i := 0; i < 10; i++ {
        id, _ := AddPart("0d3cc782c3242cf3ce4b2174e1041ed23" + strconv.Itoa(i))
        ls.PushBack(id)
    }

    fmt.Println(AddFile("0d3cc782c3242cf3ce4b2174e1041ed2", &ls))
}
