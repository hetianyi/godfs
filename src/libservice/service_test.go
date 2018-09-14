package libservice

import (
	"app"
	"container/list"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"sync"
	"testing"
	"time"
	"util/db"
	"util/logger"
)

func initParam() {
	app.BASE_PATH = "E:/godfs-storage/storage1"
	logger.SetLogLevel(2)

	// 连接数据库
	SetPool(db.NewPool(3))
}

func Test1(t *testing.T) {
	initParam()
	fmt.Println(GetFileId("asd1231", nil))
}
func Test2(t *testing.T) {
	initParam()
	fmt.Println(GetPartId("0d3cc782c3242cf3ce4b2174e1041ed233", nil))
}

func Test3(t *testing.T) {
	initParam()
	fmt.Println(AddPart("0d3cc782c3242cf3ce4b2174e1041ed233", 11))
}
func Test4(t *testing.T) {
	initParam()

	var ls = *new(list.List)
	for i := 0; i < 10; i++ {
		id, _ := AddPart("0d3cc782c3242cf3ce4b2174e1041ed23"+strconv.Itoa(i), 10)
		ls.PushBack(id)
	}

	fmt.Println(StorageAddFile("0d3cc782c3242cf3ce4b2174e1041ed2", "G001", &ls))
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
	//UpdateTrackerSyncId("xxxxxxxxx", 111)
}

func Test13(t *testing.T) {
	initParam()
	//UpdateLocalPushId("xxxxxxxxx", 222)
}

func Test14(t *testing.T) {
	initParam()
	time.Sleep(time.Second)
	go forTest(0)
	//go forTest(1)
	//go forTest(2)
	//go forTest(3)
	//go forTest(4)
	//time.Sleep(time.Second*100000000)
	timer1()
}

var total = 0
var lock1 = new(sync.Mutex)

func increase() {
	lock1.Lock()
	defer lock1.Unlock()
	total++
}
func timer1() {
	timer := time.NewTicker(time.Second)
	for {
		fmt.Println("avg:", total, "/s")
		total = 0
		<-timer.C
	}
}

func forTest(init int) {
	i := init
	s := "xxxxxxx"
	var ls list.List
	for {
		i += 1
		logger.Info(i, StorageAddFile(s+"_"+strconv.Itoa(i), "G01", &ls))
		increase()
	}
}

func Test141(t *testing.T) {
	var ls list.List
	ls.PushBack(1)
	ls.Remove(ls.Front())
	ls.Remove(ls.Front())
	ls.PushBack(2)
}
