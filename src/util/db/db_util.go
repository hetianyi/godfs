package db

import (
    _ "github.com/mattn/go-sqlite3"
)


/*
// which:
//      1:file
//      2:part
func FileOrPartExists(md5 string, which int) int {
    // check connection
    verifyConn()

    var q string
    if which == 1 {
        q = fileExistsSQL
    } else {
        q = partExistsSQL
    }

    rows, e := db.Query(q, md5)
    if e != nil {
        logger.Error("db error: error check md5")
        return 0
    }
    if rows != nil {
        for rows.Next() {
            var id int
            err := rows.Scan(&id)
            if err != nil {
                logger.Error("db error: error check md5:", err)
                return 0
            }
            // file exists
            return id
        }
    }
    return 0
}


func AddFilePart(md5 string) (int64, error) {
    ret, e1 := db.Exec(insertPartSQL, md5)
    if  e1 != nil {
        return 0, e1
    }
    lastId, _ := ret.LastInsertId()
    return lastId, nil
}

// 新增文件
func AddFile(md5 string, parts *list.List) error {

    tx, e := db.Begin()
    if e != nil {
        return e
    }

    var err error
    common.Try(func() {

        fId := FileOrPartExists(md5, 1)
        if fId == 0 {
            state, ee := tx.Prepare(insertFileSQL)
            defer state.Close()
            if ee != nil {
                panic(ee)
            }
            ret, e1 := state.Exec(md5, parts.Len())
            if  e1 != nil {
                panic(e1)
            }
            fId1, _ := ret.LastInsertId()
            fId = int(fId1)
        }

        for ele := parts.Front(); ele != nil; ele = ele.Next() {
            pId := FileOrPartExists(ele.Value.(string), 2)
            if pId == 0 {
                state, ee := tx.Prepare(insertPartSQL)
                if ee != nil {
                    panic(ee)
                }
                ret, e1 := state.Exec(ele.Value.(string))
                if  e1 != nil {
                    state.Close()
                    panic(e1)
                }
                lastId, _ := ret.LastInsertId()
                pId = int(lastId)
            }
            state, ee := tx.Prepare(insertRelationSQL)
            if ee != nil {
                panic(ee)
            }
            _, e2 := state.Exec(fId, pId)
            if  e2 != nil {
                state.Close()
                panic(e2)
            }
        }
        if e3 := tx.Commit(); e3 != nil {
            panic(e3)
        }

    }, func(i interface{}) {
        logger.Error(i)
        err = i.(error)
        tx.Rollback()
    })
    return err
}
*/



