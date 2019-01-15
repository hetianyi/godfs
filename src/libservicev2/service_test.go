package libservicev2

import (
	"testing"
	"util/db"
	"util/logger"
	"app"
	"container/list"
	"strconv"
	"fmt"
	"encoding/json"
	"time"
	"util/timeutil"
)

func init() {
	logger.SetLogLevel(1)
	app.BASE_PATH = "E:\\godfs-storage\\storage1"
	SetPool(db.NewPool(1))
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
	file := &app.FileVO{Md5: "eeeeee", PartNumber: 1, Group: "G01", Instance: "01", Finish: 1}
	ls := list.New()
	for i := 0; i < 3; i++ {
		part := &app.PartDO{Md5: "rrrr_" + strconv.Itoa(i), Size: int64(1000+i)}
		ls.PushBack(part)
	}
	for ele := ls.Front(); ele != nil; ele = ele.Next() {
		fmt.Println(ele.Value.(*app.PartDO).Md5)
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

func TestGetTrackerInfo(t *testing.T) {
	logger.Error(GetTracker("xxxxxx"))
}

func TestGetReadyPushFiles(t *testing.T) {
	ret, e := GetReadyPushFiles("xxxxxx")
	if e != nil {
		logger.Error(e)
	} else {
		for fileEle := ret.Front(); fileEle != nil; fileEle = fileEle.Next() {
			bs, _ := json.Marshal(fileEle.Value.(*app.FileVO))
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


func TestSaveTrackerInfo(t *testing.T) {
	logger.Error(SaveTracker(&app.TrackerDO{
		Uuid: "xxx",
		TrackerSyncId: 1,
		LastRegTime: timeutil.GetTimestamp(time.Now()),
		LocalPushId: 1 ,
		Host: "456",
		Port: 111,
		Status: 1,
		Secret: "123456",
		TotalFiles: 1,
		Remark: "asd",
		AddTime: timeutil.GetTimestamp(time.Now()),
	}))
}


func TestGetStorageByUUID(t *testing.T) {
	PrintResult(GetStorageByUUID("123"))
}
func TestExistsStorage(t *testing.T) {
	PrintResult(ExistsStorage("789"))
}
func TestSaveStorage(t *testing.T) {
	storage := &app.StorageDO{
		Uuid: "123",
		Host: "",
		Port: 0,
		Status: app.STATUS_ENABLED,
		TotalFiles: 0,
		Group: "",
		InstanceId: "",
		HttpPort: 0,
		HttpEnable: false,
		StartTime: 0,
		Download: 0,
		Upload: 0,
		Disk: 0,
		ReadOnly: false,
		Finish: 0,
		IOin: 0,
		IOout: 0,
	}
	SaveStorage("", *storage)
}

func TestQuerySystemStatistic(t *testing.T) {
	PrintResult(QuerySystemStatistic())
}

func TestGetAllWebTrackers(t *testing.T) {
	PrintResult(GetAllTrackers())
}

func TestUpdateTrackerStatus(t *testing.T) {
	PrintResult(UpdateTrackerStatus("xxx", 0, nil))
}

func TestInsertStorageStatisticLog(t *testing.T) {
	webStorage  := &app.StorageStatisticLogDO {
		StorageUuid: "ssssss",
		LogTime: timeutil.GetTimestamp(time.Now()),
		IOin: 1,
		IOout: 1,
		Disk: 1,
		Memory: 1,
		Download: 1,
		Upload: 1,
	}
	PrintResult(InsertStorageStatisticLog(webStorage, nil))
}

func TestGetFileCount(t *testing.T) {
	total := 0
	for ;; {
		fmt.Println(GetFileCount(), "   ", total)
		total++
	}
}

func TestGetIndexStatistic(t *testing.T) {
	PrintResult(GetIndexStatistic())
}

func TestInsertPulledTrackerFiles(t *testing.T) {
	file := &app.FileVO{
		Id: 2,
		Md5: "xxxx",
		PartNumber: 3,
		Group: "G01",
		Instance: "001",
		Finish: 0,
	}
	part := &app.PartDO{
		Id: 1,
		Md5: "123",
		Size: 1024,
	}
	parts := make([]app.PartDO, 1)
	file.Parts = parts
	parts[0] = *part
	files := make([]app.FileVO, 1)
	files[0] = *file
	PrintResult(InsertPulledTrackerFiles("xxx", files, nil))
}

func TestInsertRegisteredFiles(t *testing.T) {
	file := &app.FileVO{
		Id: 3,
		Md5: "xxxx111",
		PartNumber: 3,
		Group: "G01",
		Instance: "001",
		Finish: 0,
	}
	part := &app.PartDO{
		Id: 4,
		Md5: "123",
		Size: 1024,
	}
	parts := make([]app.PartDO, 1)
	file.Parts = parts
	parts[0] = *part
	files := make([]app.FileVO, 1)
	files[0] = *file
	PrintResult(InsertRegisteredFiles(files))
}

func TestGetReadyDownloadFiles(t *testing.T) {
	PrintResult(GetReadyDownloadFiles(1))
}