package lib_service

import (
    "testing"
    "fmt"
    "app"
    "util/logger"
)

func initParam() {
    app.BASE_PATH = "E:/WorkSpace2018/godfs"
    logger.SetLogLevel(1)
}

func Test1(t *testing.T) {
    initParam()
    fmt.Println(GetFileId("asd1231"))
}
