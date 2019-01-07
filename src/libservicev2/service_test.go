package libservicev2

import (
	"testing"
	"util/logger"
	"libcommon"
	"app"
	"container/list"
	"strconv"
	"fmt"
	"encoding/json"
	"time"
)

func init() {
	app.BASE_PATH = "E:\\godfs-storage\\storage1"
	SetPool(NewPool(1))
	logger.SetLogLevel(1)
}

func PrintResult(result... interface{}) {
	fmt.Println("\n\n+++~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~+++")
	if result != nil {
		for i := range result {
			obj := result[i]
			bs, _ := json.Marshal(obj)
			fmt.Println(string(bs))
		}
	}
	fmt.Println("+++~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~+++")
}

func TestInsertFile(t *testing.T) {
	file := &libcommon.FileVO{Md5: "eeeeee", PartNumber: 1, Group: "G01", Instance: "01", Finish: 1}
	ls := list.New()
	for i := 0; i < 3; i++ {
		part := &libcommon.PartDO{Md5: "rrrr_" + strconv.Itoa(i), Size: int64(1000+i)}
		ls.PushBack(part)
	}
	for ele := ls.Front(); ele != nil; ele = ele.Next() {
		fmt.Println(ele.Value.(*libcommon.PartDO).Md5)
	}
	file.SetParts(ls)
	InsertFile(file, nil)
	s, _ := json.Marshal(file)
	logger.Info(string(s))
}


func TestConfirmAppUUID(t *testing.T) {
	uuid := "aaaaa"
	logger.Info("before uuid is", uuid)
	logger.Info("after uuid is")
	logger.Info(ConfirmAppUUID(uuid))
}


func TestGetFileIdByMd5(t *testing.T) {
	logger.Error(GetFileIdByMd5("xxxxxx", nil))
}

func TestGetPartIdByMd5(t *testing.T) {
	logger.Error(GetPartIdByMd5("xxxxxx", nil))
}

func TestUpdateTrackerInfo(t *testing.T) {
	logger.Error(UpdateTrackerInfo(&libcommon.TrackerDO{Uuid: "xxxxxx", TrackerSyncId: 12, LastRegTime: time.Now(), LocalPushId: 11}))
}

func TestGetTrackerInfo(t *testing.T) {
	logger.Error(GetTrackerInfo("xxxxxx"))
}

func TestGetReadyPushFiles(t *testing.T) {
	ret, e := GetReadyPushFiles("xxxxxx")
	if e != nil {
		logger.Error(e)
	} else {
		for fileEle := ret.Front(); fileEle != nil; fileEle = fileEle.Next() {
			bs, _ := json.Marshal(fileEle.Value.(*libcommon.FileVO))
			fmt.Println(string(bs))
		}

	}
}

func TestGetFullFileByMd5(t *testing.T) {
	PrintResult(GetFullFileByMd5("eeeeee", 0))
}

func TestGetFullFileById(t *testing.T) {
	PrintResult(GetFullFileById(55800000004, 2))
}

func TestUpdateFileFinishStatus(t *testing.T) {
	PrintResult(UpdateFileFinishStatus(55800000003, 0, nil))
}

func TestGetFullFilesFromId(t *testing.T) {
	PrintResult(GetFullFilesFromId(4, false, "G01", 10))
}

func TestGetClientByUUID(t *testing.T) {
	PrintResult(GetStorageClientByUUID("123"))
}
func TestSaveStorageClient(t *testing.T) {
	SaveStorageClient(&libcommon.StorageDO{"123", "123", 1, 1, 123, 1, "123", "123", 12, true, 0, 12, 123, 13, true, 1, 1, 1})
}

func TestQuerySystemStatistic(t *testing.T) {
	PrintResult(QuerySystemStatistic())
}

func TestInsertWebTracker(t *testing.T) {
	storage := &libcommon.WebTrackerDO{
		Uuid: "xxxxxx",
		Host: "xxx",
		Port: 1024,
		Status: 1,
		Secret: "123456",
		TotalFiles: 3,
		Remark: "asdasd",
		AddTime: time.Now(),
	}
	PrintResult(InsertWebTracker(storage, nil))
}


