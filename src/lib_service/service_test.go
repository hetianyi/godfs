package lib_service

import (
    "testing"
    "fmt"
    "app"
    "util/logger"
    "container/list"
    "strconv"
    "math"
    "encoding/json"
    "util/db"
)

func initParam() {
    app.BASE_PATH = "E:/godfs-storage/tracker"
    logger.SetLogLevel(1)

    // 连接数据库
    SetPool(db.NewPool(app.DB_Pool_SIZE))
}

func Test1(t *testing.T) {
    initParam()
    fmt.Println(GetFileId("asd1231", nil))
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


func Test8(t *testing.T) {
    initParam()
    //fmt.Println(GetFullFileByMd5("123123a"))
    fmt.Println(GetFullFileByFid(1, 1))
}
func Test9(t *testing.T) {
    initParam()

    fmt.Println(FinishSyncTask(1))
}

func Test10(t *testing.T) {
    initParam()
    ret, _ := GetFilesBasedOnId(0)
    fmt.Println(ret.Len())
    s, _ := json.Marshal(*ret)
    fmt.Println(string(s))
}

func Test11(t *testing.T) {
    initParam()
    config, e := GetTrackerConfig("xxxx")
    if e != nil {
        logger.Error(e)
        return
    }
    fmt.Println(config)
}


func Test12(t *testing.T) {
    initParam()
    UpdateTrackerSyncId("xxxxxxxxx", 111)
}

func Test13(t *testing.T) {
    initParam()
    UpdateLocalPushId("xxxxxxxxx", 222)
}





