package db

import (
	"database/sql"
	"sync"
	"app"
	"errors"
	"os"
	"time"
	"util/file"
	"util/logger"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/jinzhu/gorm"
)
// db write lock
// when write event happens, the sys will lock by program in case if error occurs such 'database is locked'
var dbWriteLock *sync.Mutex

func init() {
	dbWriteLock = new(sync.Mutex)
}

// download sqlite3 studio @
// https://sqlitestudio.pl/index.rvt?act=download
type IDAO interface {
	InitDB()
	connect() (*sql.DB, error)
	checkDb() error
	verifyConn()
	Query(handler func(rows *sql.Rows) error, sqlString string, args ...interface{}) error
	DoTransaction(works func(tx *sql.Tx) error) error
}

type DAO struct {
	db        *gorm.DB
	connMutex *sync.Mutex
	index     int
}

func (dao *DAO) InitDB(index int) error {
	logger.Debug("initial db connection with index:", index)
	dao.connMutex = new(sync.Mutex)
	dao.index = index
	return dao.checkDb()
}

func (dao *DAO) connect() (*gorm.DB, error) {
	logger.Debug("connect db file:", app.BASE_PATH+"/data/storage.db")
	fInfo, e := os.Stat(app.BASE_PATH + "/data/storage.db")
	// if db not exists, copy template db file to data path.
	if fInfo == nil || e != nil {
		logger.Info("no db file found, init db file from template.")
		logger.Debug("copy from", app.BASE_PATH+"/conf/storage.db to", app.BASE_PATH+"/data")
		s, e1 := file.CopyFileTo(app.BASE_PATH+"/conf/storage.db", app.BASE_PATH+"/data")
		if !s || e1 != nil {
			logger.Fatal("error prepare db file:", e1)
		}
	}
	return gorm.Open("sqlite3", app.BASE_PATH+"/data/storage.db?cache=shared&_synchronous=0")
}

func (dao *DAO) checkDb() error {
	dao.connMutex.Lock()
	defer dao.connMutex.Unlock()
	for {
		if dao.db == nil {
			tdb, e := dao.connect()
			if e != nil {
				logger.Error("error connect db, wait...:", app.BASE_PATH+"/data/storage.db")
				time.Sleep(time.Second * 1)
				continue
			}
			if app.LOG_LEVEL < 2 {
				tdb.LogMode(true)
			}
			dao.db = tdb
			logger.Debug("connect db success")
			return nil
		} else {
			return dao.db.DB().Ping()
		}
	}
}

func (dao *DAO) verifyConn() error {
	for i := 0; i < 5; i++ {
		if e := dao.checkDb(); e != nil {
			logger.Error("error check db:", e)
			dao.db.Close()
			dao.db = nil
		} else {
			return nil
		}
	}
	return errors.New("error check db: failed retry many times")
}

// do db query
func (dao *DAO) Query(handler func(*gorm.DB) error) error {
	return handler(dao.db)
}

// do db transaction
func (dao *DAO) DoTransaction(work func(*gorm.DB) error) error {
	dbWriteLock.Lock()
	defer dbWriteLock.Unlock()
	// Note the use of tx as the database handle once you are within a transaction
	tx := dao.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Error; err != nil {
		return err
	}
	e := work(tx)
	if e != nil {
		logger.Warn("roll back transaction due to:", e.Error())
		tx.Rollback()
		return e
	}
	return tx.Commit().Error
}
