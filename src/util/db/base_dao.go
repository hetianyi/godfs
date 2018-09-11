package db

import (
    "database/sql"
    "sync"
)

import (
    _ "github.com/mattn/go-sqlite3"
    "util/logger"
    "app"
    "time"
    "util/common"
    "os"
    "util/file"
    "errors"
)

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
    db *sql.DB
    connMutex *sync.Mutex
    index int
}


func (dao *DAO) InitDB(index int) error {
    logger.Debug("initial db connection with index:", index)
    dao.connMutex = new(sync.Mutex)
    dao.index = index
    return dao.checkDb()
}

func (dao *DAO) connect() (*sql.DB, error) {
    logger.Debug("connect db file:", app.BASE_PATH + "/data/storage.db")
    fInfo, e := os.Stat(app.BASE_PATH + "/data/storage.db")
    // if db not exists, copy template db file to data path.
    if fInfo == nil || e != nil {
        logger.Info("no db file found, init db file from template.")
        logger.Debug("copy from", app.BASE_PATH + "/conf/storage.db to", app.BASE_PATH + "/data")
        s, e1 := file.CopyFileTo(app.BASE_PATH + "/conf/storage.db", app.BASE_PATH + "/data")
        if !s || e1 != nil {
            logger.Fatal("error prepare db file:", e1)
        }
    }
    return sql.Open("sqlite3", app.BASE_PATH + "/data/storage.db")
}

func (dao *DAO) checkDb() error {
    dao.connMutex.Lock()
    defer dao.connMutex.Unlock()
    for {
        if dao.db == nil {
            tdb, e := dao.connect()
            if e != nil {
                logger.Error("error connect db, wait...:", app.BASE_PATH + "/data/storage.db")
                time.Sleep(time.Second * 1)
                continue
            }
            dao.db = tdb
            logger.Debug("connect db success")
            return nil
        } else {
            return dao.db.Ping()
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



// db db query
func (dao *DAO) Query(handler func(rows *sql.Rows) error, sqlString string, args ...interface{}) error {
    te := dao.verifyConn()
    if te != nil {
        return te
    }
    var rs *sql.Rows
    var e error
    logger.Debug("exec SQL:\n\t" + sqlString)
    if args == nil || len(args) == 0 {
        rs, e = dao.db.Query(sqlString)
    } else {
        rs, e = dao.db.Query(sqlString, args...)
    }
    if e != nil {
        return e
    }
    return handler(rs)
}


func (dao *DAO) DoTransaction(works func(tx *sql.Tx) error) error {
    te := dao.verifyConn()
    if te != nil {
        return te
    }
    tx, e1 := dao.db.Begin()
    if e1 != nil {
        return e1
    }
    var globalErr error
    common.Try(func() {
        e2 := works(tx)
        if e2 != nil {
            logger.Debug("roll back")
            tx.Rollback()
            globalErr = e2
        } else {
            if e3 := tx.Commit(); e3 != nil {
                logger.Debug("roll back")
                tx.Rollback()
                globalErr = e3
            }
        }
    }, func(i interface{}) {
        logger.Debug("roll back")
        tx.Rollback()
        globalErr = i.(error)
    })
    return globalErr
}

