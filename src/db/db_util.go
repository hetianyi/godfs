package db

import (
    _ "github.com/mattn/go-sqlite3"
    "database/sql"
    "util/logger"
    "fmt"
    "util/file"
)

var db *sql.DB

func init() {
    fmt.Println(file.Exists("./storage.db"))
    openDb, e := sql.Open("sqlite3", "./storage.db")
    db = openDb
    if e != nil {
        logger.Fatal("error open db")
    }
}




func Query() {
    q := "select a.md5, a.ext from files a"
    rows, e := db.Query(q)
    if e != nil {
        logger.Fatal("error query db")
    }

    if rows != nil {
        for rows.Next() {
            var md5 string
            var ext string
            err := rows.Scan(&md5, &ext)
            if err != nil {
                logger.Error("Failed to db.Query:", err)
            }
            logger.Info("md5:", md5, "ext:", ext)
        }
    }

}



