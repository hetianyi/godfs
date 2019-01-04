package libservicev2

import (
	"libcommon"
	"github.com/jinzhu/gorm"
)

var dbPool *DbConnPool

func SetPool(pool *DbConnPool) {
	dbPool = pool
}


// insert new file to table file,
// if file exists, file id will replaced by existing id.
func InsertFile(file *libcommon.FileDO, dao *DAO) error {
	if dao == nil {
		dao1, ef := dbPool.GetDB()
		if ef != nil {
			return ef
		}
		dao = dao1
		defer dbPool.ReturnDB(dao)
	}
	return dao.Query(func(db *gorm.DB) error {
		db.Where(libcommon.FileDO{Md5: file.Md5}).FirstOrCreate(file)
		return nil
	})
}

// insert new file to table part,
// if part exists, part id will replaced by existing id.
func InsertFilePart(part *libcommon.PartDO, dao *DAO) error {
	if dao == nil {
		dao1, ef := dbPool.GetDB()
		if ef != nil {
			return ef
		}
		dao = dao1
		defer dbPool.ReturnDB(dao)
	}
	return dao.Query(func(db *gorm.DB) error {
		db.Where(libcommon.PartDO{Md5: part.Md5}).FirstOrCreate(part)
		return nil
	})
}

// insert new file_part relation to table relation_file_part,
// if relation exists, relation id will replaced by existing id.
func InsertFilePartRelation(relation *libcommon.FilePartRelationDO, dao *DAO) error {
	if dao == nil {
		dao1, ef := dbPool.GetDB()
		if ef != nil {
			return ef
		}
		dao = dao1
		defer dbPool.ReturnDB(dao)
	}
	return dao.Query(func(db *gorm.DB) error {
		db.Where(libcommon.FilePartRelationDO{FileId: relation.FileId, PartId: relation.PartId}).FirstOrCreate(relation)
		return nil
	})
}

// save current app uuid to table sys,
// if uuid already exists, skip
func ConfirmAppUUID(uuid string) (ret string, e error) {
	dao, ef := dbPool.GetDB()
	if ef != nil {
		return "", ef
	}
	defer dbPool.ReturnDB(dao)
	uuidDO := &libcommon.SysDO{Key: "uuid", Value: uuid}
	return uuidDO.Value, dao.Query(func(db *gorm.DB) error {
		db.Where(libcommon.SysDO{Key: "uuid"}).FirstOrCreate(uuidDO)
		return nil
	})
}






