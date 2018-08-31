package db

import (
    "app"
    "util/logger"
    "testing"
)

func initParam() {
    app.BASE_PATH = "E:/WorkSpace2018/godfs"
    logger.SetLogLevel(1)
}
/*
func Test1(t *testing.T) {
    initParam()
    InitDB()

    fmt.Println(FileOrPartExists("0d3cc782c3242cf3ce4b2174e1041ed2", 2))
}
func Test2(t *testing.T) {
    initParam()
    InitDB()

    fmt.Println(AddFilePart("0d3cc782c3242cf3ce4b2174e1041ed2"))
}

func Test3(t *testing.T) {
    initParam()
    InitDB()
    var parts list.List
    parts.PushBack("0d3cc782c3242cf3ce4b2174e1041ed2")
    fmt.Println(AddFile("0d3cc782c3242cf3ce4b2174e1041ed2", &parts))
}*/

func Test4(t *testing.T) {
    /*initParam()
    InitDB()

    sqlString := "select id from files"
    err := Query(func(rows *sql.Rows) error {
        if rows != nil {
            for rows.Next() {
                var id int
                e := rows.Scan(&id)
                if e != nil {
                    return e
                }
                logger.Info(id)
            }
        }
        return nil
    }, sqlString)
    logger.Info(err)*/
}


func Test5(t *testing.T) {
    /*initParam()
    InitDB()

    sqlString1 := "select id from files"
    sqlString2 := "delete from files"

    err := DoTransaction(func(tx *sql.Tx) error {
        err := Query(func(rows *sql.Rows) error {
            if rows != nil {
                for rows.Next() {
                    var id int
                    e := rows.Scan(&id)
                    if e != nil {
                        return e
                    }
                    logger.Info(id)
                }
            }
            return nil
        }, sqlString1)
        if err != nil {
            logger.Info(err)
            return err
        }
        state, e2 := tx.Prepare(sqlString2)
        if e2 != nil {
            logger.Info(e2)
            return e2
        }
        ret, e3 := state.Exec()
        if e3 != nil {
            logger.Info(e3)
            return e3
        }
        afr, _ := ret.RowsAffected()
        logger.Info("delete success:", afr)
        panic(errors.New("test rollback"))
        return nil
    })
    logger.Info(err)*/

}


