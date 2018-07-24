package lib_service

import (
    "container/list"
    "database/sql"
    "util/db"
    "app"
    "lib_common/bridge"
)

const (
    insertFileSQL  = "insert into files(md5, parts_num, instance, finish) values(?,?,?,?)"
    insertPartSQL  = "insert into parts(md5, size) values(?,?)"
    insertRelationSQL  = "insert into parts_relation(fid, pid) values(?, ?)"
    fileExistsSQL  = "select id from files a where a.md5 = ?"
    partExistsSQL  = "select id from parts a where a.md5 = ?"

    addSyncTaskSQL  = "insert into task(fid, type, status) values(?,?,?)"
    finishSyncTaskSQL  = "update task set status=0 where fid=?"
    getSyncTaskSQL  = `select fid, type from task where status=1 limit ?`
    getFullFileSQL1  = `select b.id, b.md5, b.instance from files b where b.md5=? `
    getFullFileSQL11  = `select b.id, b.md5, b.instance from files b where b.id=? `
    getFullFileSQL2  = `select d.md5, d.size
                        from files b
                        left join parts_relation c on b.id = c.fid
                        left join parts d on c.pid = d.id where b.md5=? order by d.id`
    getFullFileSQL21  = `select d.md5, d.size
                        from files b
                        left join parts_relation c on b.id = c.fid
                        left join parts d on c.pid = d.id where b.id=? order by d.id`

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
        ret, e3 := state.Exec(md5, parts.Len(), app.INSTANCE_ID, 1)
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
        return AddSyncTask(fid, app.TASK_REPORT_FILE, tx)
    })
}


// tracker add file
// parts is map[md5]partsize
func TrackerAddFile(meta *bridge.OperationRegisterFileRequest) (int, error) {
    var fid int
    var e error
    fid, e = GetFileId(meta.File.Md5)
    if e != nil {
        return 0, e
    }
    if fid > 0 {
        return fid, nil
    }
    err := db.DoTransaction(func(tx *sql.Tx) error {
        fi := meta.File
        parts := fi.Parts
        instance := fi.Instance
        state, e2 := tx.Prepare(insertFileSQL)
        if e2 != nil {
            return e2
        }
        ret, e3 := state.Exec(fi.Md5, fi.PartNum, instance, 1)
        if e3 != nil {
            return e3
        }
        lastId, e4 := ret.LastInsertId()
        if e4 != nil {
            return e4
        }
        id := int(lastId)
        fid = id
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
    if err != nil {
        return 0, err
    }
    return fid, nil
}


// 将同步任务写到task表中，然后由定时任务读取，同步给tracker，再由tracker服务器广播
// task的type暂定位1，标识要同步到tracker服务器
// status, 1:有效，0:无效
func AddSyncTask(fid int, taskType int, tx *sql.Tx) error {
    state, e2 := tx.Prepare(addSyncTaskSQL)
    if e2 != nil {
        return e2
    }
    ret, e3 := state.Exec(fid, taskType, 1)
    if e3 != nil {
        return e3
    }
    _, e4 := ret.LastInsertId()
    if e4 != nil {
        return e4
    }
    return nil
}

func FinishSyncTask(taskId int) error {
    return db.DoTransaction(func(tx *sql.Tx) error {
        state, e2 := tx.Prepare(finishSyncTaskSQL)
        if e2 != nil {
            return e2
        }
        ret, e3 := state.Exec(taskId)
        if e3 != nil {
            return e3
        }
        _, e4 := ret.LastInsertId()
        if e4 != nil {
            return e4
        }
        return nil
    })
}

func GetSyncTask() (*list.List, error) {
    var ls list.List
    e := db.Query(func(rows *sql.Rows) error {
        if rows != nil {
            for rows.Next() {
                var fid, taskType int
                e1 := rows.Scan(&fid, &taskType)
                if e1 != nil {
                    return e1
                }
                ret := &bridge.Task{FileId: fid, TaskType: taskType}
                ls.PushBack(ret)
            }
        }
        return nil
    }, getSyncTaskSQL, 10)
    if e != nil {
        return nil, e
    }
    return &ls, nil
}

func GetFullFileByMd5(md5 string) (*bridge.File, error) {
    var fi *bridge.File
    // query file
    e1 := db.Query(func(rows *sql.Rows) error {
        if rows != nil {
            for rows.Next() {
                var id int
                var md5, instance string
                e1 := rows.Scan(&id, &md5, &instance)
                if e1 != nil {
                    return e1
                } else {
                    fi = &bridge.File{Id: id, Md5: md5, Instance: instance}
                }
            }
        }
        return nil
    }, getFullFileSQL1, md5)

    if e1 != nil {
        return nil, e1
    }
    // not exists
    if fi == nil {
        return nil, nil
    }

    e2 := db.Query(func(rows *sql.Rows) error {
        if rows != nil {
            var tparsList list.List
            for rows.Next() {
                var size int64
                var md5 string
                e1 := rows.Scan(&md5, &size)
                if e1 != nil {
                    return e1
                } else {
                    var part = &bridge.FilePart{Md5: md5, FileSize: size}
                    tparsList.PushBack(part)
                }
            }
            var parsList = make([]bridge.FilePart, tparsList.Len())
            index := 0
            for ele := tparsList.Front(); ele != nil; ele = ele.Next() {
                parsList[index] = *ele.Value.(*bridge.FilePart)
                index++
            }
            fi.PartNum = tparsList.Len()
            fi.Parts = parsList
        }
        return nil
    }, getFullFileSQL2, md5)

    if e2 != nil {
        return nil, e2
    }
    return fi, nil
}


func GetFullFileByFid(fid int) (*bridge.File, error) {
    var fi *bridge.File
    // query file
    e1 := db.Query(func(rows *sql.Rows) error {
        if rows != nil {
            for rows.Next() {
                var id int
                var md5, instance string
                e1 := rows.Scan(&id, &md5, &instance)
                if e1 != nil {
                    return e1
                } else {
                    fi = &bridge.File{Id: id, Md5: md5, Instance: instance}
                }
            }
        }
        return nil
    }, getFullFileSQL11, fid)

    if e1 != nil {
        return nil, e1
    }
    // not exists
    if fi == nil {
        return nil, nil
    }

    e2 := db.Query(func(rows *sql.Rows) error {
        if rows != nil {
            var tparsList list.List
            for rows.Next() {
                var size int64
                var md5 string
                e1 := rows.Scan(&md5, &size)
                if e1 != nil {
                    return e1
                } else {
                    var part = &bridge.FilePart{Md5: md5, FileSize: size}
                    tparsList.PushBack(part)
                }
            }
            var parsList = make([]bridge.FilePart, tparsList.Len())
            index := 0
            for ele := tparsList.Front(); ele != nil; ele = ele.Next() {
                parsList[index] = *ele.Value.(*bridge.FilePart)
                index++
            }
            fi.PartNum = tparsList.Len()
            fi.Parts = parsList
        }
        return nil
    }, getFullFileSQL21, fid)

    if e2 != nil {
        return nil, e2
    }
    return fi, nil
}




