package libservicev2

import (
	"testing"
	"util/logger"
	"libcommon"
	"app"
)

func init() {
	app.BASE_PATH = "E:\\godfs-storage\\storage1"
	SetPool(NewPool(1))
	logger.SetLogLevel(1)
}

func TestInsertFile(t *testing.T) {
	file := &libcommon.FileDO{Md5: "xxxxxx", PartNumber: 1, Group: "G02", Instance: "01", Finish: 1}
	logger.Error("file id =", file.Id)
	InsertFile(file, nil)
	logger.Error("file id =", file.Id)
}

func TestInsertFilePart(t *testing.T) {
	part := &libcommon.PartDO{Md5: "xxxxxx", Size: 1001}
	logger.Error("part id =", part.Id)
	InsertFilePart(part, nil)
	logger.Error("part id =", part.Id)
}

func TestInsertFilePartRelation(t *testing.T) {
	relation := &libcommon.FilePartRelationDO{FileId: 1, PartId: 1}
	logger.Error("relation id =", relation.Id)
	InsertFilePartRelation(relation, nil)
	logger.Error("relation id =", relation.Id)
}

func TestConfirmAppUUID(t *testing.T) {
	uuid := "aaaaa"
	logger.Info("before uuid is", uuid)
	logger.Info("after uuid is")
	logger.Info(ConfirmAppUUID(uuid))
}

