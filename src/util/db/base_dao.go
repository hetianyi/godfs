package db

import (
    "database/sql"
    "util/logger"
    "sync"
    "app"
    "time"
    "util/common"
)

// download sqlite3 studio @
// https://sqlitestudio.pl/index.rvt?act=download


var db *sql.DB
var connMutex sync.Mutex

func InitDB() {
    logger.Debug("initial db connection")
    connMutex = *new(sync.Mutex)
    checkDb()
}


func connect() (*sql.DB, error) {
    return sql.Open("sqlite3", app.BASE_PATH + "/conf/storage.db")
}

func checkDb() error {
    connMutex.Lock()
    defer connMutex.Unlock()
    for {
        if db == nil {
            tdb, e := connect()
            if e != nil {
                logger.Error("error connect db, wait...:", app.BASE_PATH + "/data/storage.db")
                time.Sleep(time.Second * 5)
                continue
            }
            db = tdb
            logger.Debug("connect db success")
            return nil
        } else {
            return db.Ping()
        }
    }
}

func verifyConn() {
    for {
        if checkDb() != nil {
            db = nil
            time.Sleep(time.Second * 2)
        } else {
            break
        }
    }
}



// db db query
func Query(handler func(rows *sql.Rows) error, sqlString string, args ...interface{}) error {
    verifyConn()
    var rs *sql.Rows
    var e error
    if args == nil || len(args) == 0 {
        rs, e = db.Query(sqlString)
    } else {
        rs, e = db.Query(sqlString, args...)
    }
    if e != nil {
        return e
    }
    return handler(rs)
}


func DoTransaction(works func(tx *sql.Tx) error) error {
    verifyConn()
    tx, e1 := db.Begin()
    if e1 != nil {
        return e1
    }
    var globalErr error
    common.Try(func() {
        e2 := works(tx)
        if e2 != nil {
            logger.Info("roll back")
            tx.Rollback()
            globalErr = e2
        }
        if e3 := tx.Commit(); e3 != nil {
            logger.Info("roll back")
            tx.Rollback()
            globalErr = e3
        }
    }, func(i interface{}) {
        logger.Info("roll back")
        tx.Rollback()
        globalErr = i.(error)
    })
    return globalErr
}

