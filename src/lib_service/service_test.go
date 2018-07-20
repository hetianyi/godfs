package lib_service

import (
    "testing"
    "fmt"
    "app"
    "util/logger"
    "container/list"
    "strconv"
    "math"
    "lib_common/header"
)

func initParam() {
    app.BASE_PATH = "D:/godfs"
    logger.SetLogLevel(1)
}

func Test1(t *testing.T) {
    initParam()
    fmt.Println(GetFileId("asd1231"))
}
func Test2(t *testing.T) {
    initParam()
    fmt.Println(GetPartId("0d3cc782c3242cf3ce4b2174e1041ed233"))
}

func Test3(t *testing.T) {
    initParam()
    fmt.Println(AddPart("0d3cc782c3242cf3ce4b2174e1041ed233", 11))
}
func Test4(t *testing.T) {
    initParam()

    var ls = *new(list.List)
    for i := 0; i < 10; i++ {
        id, _ := AddPart("0d3cc782c3242cf3ce4b2174e1041ed23" + strconv.Itoa(i), 10)
        ls.PushBack(id)
    }

    fmt.Println(StorageAddFile("0d3cc782c3242cf3ce4b2174e1041ed2", &ls))
}

func Test5(t *testing.T) {
    fmt.Println(math.MaxInt32)
}



func Test6(t *testing.T) {
    initParam()
    //fmt.Println(AddSyncTask(6))
}

func Test7(t *testing.T) {
    initParam()
    ls, _ := GetSyncTask()
    for ele := ls.Front(); ele != nil; ele = ele.Next() {
        m := ele.Value.(*header.File)
        fmt.Println(m.Id, m.Md5, m.Instance, "(", m.PartNum, ")")
        for i := 0; i < len(m.Parts); i++ {
            fmt.Println("->", m.Parts[i].Md5, "(", m.Parts[i].FileSize, ")")
        }
    }
}

func Test8(t *testing.T) {
    initParam()
    fmt.Println(GetFullFile("f3d5a643583ed27cf865ade45698e692"))
}





