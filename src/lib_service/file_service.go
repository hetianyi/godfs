package lib_service

import (
    "container/list"
    "database/sql"
    "util/db"
    "lib_common/header"
    "encoding/json"
    "app"
)

const (
    insertFileSQL  = "insert into files(md5, parts_num, instance) values(?,?,?)"
    insertPartSQL  = "insert into parts(md5, size) values(?,?)"
    insertRelationSQL  = "insert into parts_relation(fid, pid) values(?, ?)"
    fileExistsSQL  = "select id from files a where a.md5 = ?"
    partExistsSQL  = "select id from parts a where a.md5 = ?"

    addSyncTaskSQL  = "insert into task(type, fid, status) values(?,?,?)"
    getSyncTaskSQL  = `select b.id, b.md5, b.instance, '['||group_concat('{"md5":"'||d.md5||'","size":'||d.size||'}')||']' as parts from task a 
                        left join files b on a.fid = b.id 
                        left join parts_relation c on b.id = c.fid
                        left join parts d on c.pid = d.id
                        where a.type = 1 and a.status = 1 limit 10  `

)

// get file id by md5
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

// get part id by md5
func GetPartId(md5 string) (int, error) {
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
    }, partExistsSQL, md5)
    if e != nil {
        return 0, e
    }
    return id, nil
}


func AddPart(md5 string, size int64) (int, error) {
    pid, ee := GetPartId(md5)
    if ee != nil {
        return 0, ee
    }
    // file exists, will skip
    if pid != 0 {
        return pid, nil
    }
    var id int
    err := db.DoTransaction(func(tx *sql.Tx) error {
        state, e2 := tx.Prepare(insertPartSQL)
        if e2 != nil {
            return e2
        }
        ret, e3 := state.Exec(md5, size)
        if e3 != nil {
            return e3
        }
        lastId, e4 := ret.LastInsertId()
        if e4 != nil {
            return e4
        }
        id = int(lastId)
        return nil
    })
    return id, err
}

// storage add file and add new sync task
// parts is part id list
func StorageAddFile(md5 string, parts *list.List) error {
    fid, ee := GetFileId(md5)
    if ee != nil {
        return ee
    }
    // file exists, will skip
    if fid != 0 {
        return nil
    }
    return db.DoTransaction(func(tx *sql.Tx) error {
        state, e2 := tx.Prepare(insertFileSQL)
        if e2 != nil {
            return e2
        }
        ret, e3 := state.Exec(md5, parts.Len(), app.INSTANCE_ID)
        if e3 != nil {
            return e3
        }
        lastId, e4 := ret.LastInsertId()
        if e4 != nil {
            return e4
        }
        fid := int(lastId)
        for ele := parts.Front(); ele != nil; ele = ele.Next() {
            state, e2 := tx.Prepare(insertRelationSQL)
            if e2 != nil {
                return e2
            }
            _, e3 := state.Exec(fid, ele.Value)
            if e3 != nil {
                return e3
            }
        }
        return AddSyncTask(fid, tx)
    })
}


// tracker add file
// parts is map[md5]partsize
func TrackerAddFile(meta *header.CommunicationRegisterFileRequestMeta) error {
    return db.DoTransaction(func(tx *sql.Tx) error {
        fi := meta.File
        parts := fi.Parts
        instance := fi.Instance
        state, e2 := tx.Prepare(insertFileSQL)
        if e2 != nil {
            return e2
        }
        ret, e3 := state.Exec(fi.Md5, fi.PartNum, instance)
        if e3 != nil {
            return e3
        }
        lastId, e4 := ret.LastInsertId()
        if e4 != nil {
            return e4
        }
        id := int(lastId)

        for i := range parts {
            state, e2 := tx.Prepare(insertPartSQL)
            if e2 != nil {
                return e2
            }
            ret, e3 := state.Exec(parts[i].Md5, parts[i].FileSize)
            if e3 != nil {
                return e3
            }
            lastPid, e5 := ret.LastInsertId()
            if e5 != nil {
                return e5
            }

            state1, e6 := tx.Prepare(insertRelationSQL)
            if e6 != nil {
                return e6
            }
            ret1, e7 := state1.Exec(id, lastPid)
            if e7 != nil {
                return e7
            }
            _, e8 := ret1.LastInsertId()
            if e8 != nil {
                return e8
            }
        }
        return nil
    })
}


// 将同步任务写到task表中，然后由定时任务读取，同步给tracker，再由tracker服务器广播
// task的type暂定位1，标识要同步到tracker服务器
// status, 1:有效，0:无效
func AddSyncTask(fid int, tx *sql.Tx) error {
    state, e2 := tx.Prepare(addSyncTaskSQL)
    if e2 != nil {
        return e2
    }
    ret, e3 := state.Exec(1, fid, 1)
    if e3 != nil {
        return e3
    }
    _, e4 := ret.LastInsertId()
    if e4 != nil {
        return e4
    }
    return nil
}

func GetSyncTask() (*list.List, error) {
    var ls list.List
    e := db.Query(func(rows *sql.Rows) error {
        if rows != nil {
            for rows.Next() {
                var id int
                var md5, instance, parts string
                e1 := rows.Scan(&id, &md5, &instance, &parts)
                if e1 != nil {
                    return e1
                }
                var parsList []header.FilePart
                e2 := json.Unmarshal([]byte(parts), &parsList)
                if e2 != nil {
                    return e2
                }
                ret := &header.File{Id: id, Md5: md5, Instance: instance, PartNum: len(parsList), Parts: parsList}
                ls.PushBack(ret)
            }
        }
        return nil
    }, getSyncTaskSQL)
    if e != nil {
        return nil, e
    }
    return &ls, nil
}



