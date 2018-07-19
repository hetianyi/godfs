package lib_service

import (
    "container/list"
    "database/sql"
    "util/db"
)

const (
    insertFileSQL  = "insert into files(md5, parts_num) values(?,?)"
    insertPartSQL  = "insert into parts(md5) values(?)"
    insertRelationSQL  = "insert into parts_relation(fid, pid) values(?, ?)"
    fileExistsSQL  = "select id from files a where a.md5 = ?"
    partExistsSQL  = "select id from parts a where a.md5 = ?"
)


func ExistsPart(md5 string) bool {
    return false
}

func ExistsFile(md5 string) bool {
    return false
}

func GetFileId(md5 string) (int, error) {
    var id = 0
    e := db.Query(func(rows *sql.Rows) error {
        if rows != nil {
            for rows.Next() {
                e := rows.Scan(&id)
                if e != nil {
                    return e
                }
            }
        }
        return nil
    }, fileExistsSQL, md5)
    if e != nil {
        return 0, e
    }
    return id, nil
}

func GetPartId(md5 string) (int, error) {
    return 0,nil
}


func AddPart(md5 string) (int, error) {
    return 0,nil
}

func AddFile(md5 string, parts *list.List) {

}



