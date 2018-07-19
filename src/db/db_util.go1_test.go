package db

import (
    "testing"
    "app"
    "util/logger"
    "fmt"
    "container/list"
)

func initParam() {
    app.BASE_PATH = "D:/godfs"
    logger.SetLogLevel(1)
}

func Test1(t *testing.T) {
    initParam()
    InitDB()

    fmt.Println(FileOrPartExists("0d3cc782c3242cf3ce4b2174e1041ed2", 2))
}
func Test2(t *testing.T) {
    initParam()
    InitDB()

    fmt.Println(AddFilePart("0d3cc782c3242cf3ce4b2174e1041ed2"))
}

func Test3(t *testing.T) {
    initParam()
    InitDB()
    var parts list.List
    parts.PushBack("0d3cc782c3242cf3ce4b2174e1041ed2")
    fmt.Println(AddFile("0d3cc782c3242cf3ce4b2174e1041ed2", &parts))
}

