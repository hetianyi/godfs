package db

import (
    _ "github.com/mattn/go-sqlite3"
    "database/sql"
    "util/logger"
    "sync"
    "time"
)

// download sqlite3 studio @
// https://sqlitestudio.pl/index.rvt?act=download

var db *sql.DB
var connMutex sync.Mutex

const (
    insertFileSQL  = "insert into files(md5, status, download_times, create_time) values(?,1,0,datetime('now','localtime'))"
    checkBeforeInsertFileSQL  = "select count(*) from files a where a.md5 = ?"
)


func init() {
    connMutex = *new(sync.Mutex)
}

func connect() (*sql.DB, error) {
    return sql.Open("sqlite3", "./storage.db")
}

func checkDb() error {
    connMutex.Lock()
    defer connMutex.Unlock()
    for {
        if db == nil {
            tdb, e := connect()
            if e != nil {
                time.Sleep(time.Second * 5)
            }
            db = tdb
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


func FileExists(md5 string) bool {
    // check connection
    verifyConn()
    rows, e := db.Query(checkBeforeInsertFileSQL, md5)
    if e != nil {
        logger.Error("db error: error check md5")
        return false
    }
    if rows != nil {
        for rows.Next() {
            var rc int
            err := rows.Scan(&rc)
            if err != nil {
                logger.Error("db error: error check md5:", err)
                return false
            }
            // file exists
            if rc > 0 {
                return true
            } else {
                return false
            }
        }
    }
    return false
}


// 新增文件
func AddFile(md5 string) error {
    if FileExists(md5) {
        return nil
    }
    if _, e1 := db.Exec(insertFileSQL, md5); e1 != nil {
        return e1
    }
    return nil
}




