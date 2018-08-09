package lib_service

import (
    "container/list"
    "database/sql"
    "util/db"
    "app"
    "lib_common/bridge"
    "strconv"
    "bytes"
    "util/logger"
)

const (
    insertFileSQL  = "insert into files(md5, parts_num, instance, finish) values(?,?,?,?)"
    updateFileStatusSQL  = "update files set finish=1 where id=?"
    insertPartSQL  = "insert into parts(md5, size) values(?,?)"
    insertRelationSQL  = "insert into parts_relation(fid, pid) values(?, ?)"
    fileExistsSQL  = "select id from files a where a.md5 = ? "
    partExistsSQL  = "select id from parts a where a.md5 = ?"

    addSyncTaskSQL  = "insert into task(fid, type, status) values(?,?,?)"
    finishSyncTaskSQL  = "update task set status=0 where fid=?"
    getSyncTaskSQL  = `select fid, type from task a
                        left join files b on a.fid=b.id where status=1 and type=? and b.instance in (?) limit ?`
    getFullFileSQL1  = `select b.id, b.md5, b.instance, parts_num from files b where b.md5=? `
    getFullFileSQL11  = `select b.id, b.md5, b.instance, parts_num from files b where b.id=? `
    getFullFileSQL12  = `select b.id, b.md5, b.instance, parts_num from files b where b.id > ? limit 10`
    getFullFileSQL2  = `select d.md5, d.size
                        from files b
                        left join parts_relation c on b.id = c.fid
                        left join parts d on c.pid = d.id where b.md5=?`
    getFullFileSQL21  = `select d.md5, d.size
                        from files b
                        left join parts_relation c on b.id = c.fid
                        left join parts d on c.pid = d.id where b.id=?`
    getFullFileSQL22  = `select b.id, d.md5, d.size
                        from files b
                        left join parts_relation c on b.id = c.fid
                        left join parts d on c.pid = d.id where b.id in(`

    updateSyncId  = `replace into sys(id, master_sync_id) values(1, ?)`
    getSyncId  = `select master_sync_id from sys where id=1`


    updateTrackerSyncId = `replace into trackers(uuid, master_sync_id, last_reg_time, local_push_id)
                            values(?, ?, datetime('now','localtime'), (select local_push_id from trackers where uuid = ?))`
    updateLocalPushId = `replace into trackers(uuid, master_sync_id, last_reg_time, local_push_id)
                            values(?, 
                            (select master_sync_id from trackers where uuid = ?),
                            (select last_reg_time from trackers where uuid = ?), ?)`

    getTrackerConfig = `select master_sync_id, local_push_id from trackers where uuid=?`

)

var dbPool *db.DbConnPool

func SetPool(pool *db.DbConnPool) {
    dbPool = pool
}

// get file id by md5
func GetFileId(md5 string) (int, error) {
    var id = 0
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    e := dao.Query(func(rows *sql.Rows) error {
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
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    e := dao.Query(func(rows *sql.Rows) error {
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
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    err := dao.DoTransaction(func(tx *sql.Tx) error {
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
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    return dao.DoTransaction(func(tx *sql.Tx) error {
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



// storage add file which is not exist at local
func StorageAddRemoteFile(fi *bridge.File) error {
    fid, ee := GetFileId(fi.Md5)
    if ee != nil {
        return ee
    }
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    // file exists, will skip
    if fid != 0 {
        return dao.DoTransaction(func(tx *sql.Tx) error {
            return UpdateSyncId(fi.Id, tx)
        })
    }

    return dao.DoTransaction(func(tx *sql.Tx) error {
        state, e2 := tx.Prepare(insertFileSQL)
        if e2 != nil {
            return e2
        }
        ret, e3 := state.Exec(fi.Md5, fi.PartNum, fi.Instance, 0)
        if e3 != nil {
            return e3
        }
        lastId, e4 := ret.LastInsertId()
        if e4 != nil {
            return e4
        }
        fid := int(lastId)
        for i := range fi.Parts {
            state1, e4 := tx.Prepare(insertPartSQL)
            if e4 != nil {
                return e4
            }
            ret1, e5 := state1.Exec(fi.Parts[i].Md5, fi.Parts[i].FileSize)
            if e5 != nil {
                return e5
            }
            lastPartId, e6 := ret1.LastInsertId()
            if e6 != nil {
                return e6
            }
            pid := int(lastPartId)
            state2, e2 := tx.Prepare(insertRelationSQL)
            if e2 != nil {
                return e2
            }
            _, e3 := state2.Exec(fid, pid)
            if e3 != nil {
                return e3
            }
        }
        if ee := UpdateSyncId(fi.Id, tx); ee != nil {
            return ee
        }
        return AddSyncTask(fid, app.TASK_DOWNLOAD_FILE, tx)
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
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    err := dao.DoTransaction(func(tx *sql.Tx) error {
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
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    return dao.DoTransaction(func(tx *sql.Tx) error {
        state, e2 := tx.Prepare(finishSyncTaskSQL)
        if e2 != nil {
            return e2
        }
        _, e3 := state.Exec(taskId)
        if e3 != nil {
            return e3
        }
        return nil
    })
}

func GetTask(tasType int, instanceIds string) (*list.List, error) {
    var ls list.List
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    e := dao.Query(func(rows *sql.Rows) error {
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
    }, getSyncTaskSQL, tasType, instanceIds, 10)
    if e != nil {
        return nil, e
    }
    return &ls, nil
}

// finishSync=1 只查已在本地同步完成的文件
func GetFullFileByMd5(md5 string, finishFlag int) (*bridge.File, error) {

    var addOn = ""
    if finishFlag == 1 {
        addOn = " and finish=1"
    } else if finishFlag == 0 {
        addOn = " and finish=0"
    }
    var fi *bridge.File
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    // query file
    e1 := dao.Query(func(rows *sql.Rows) error {
        if rows != nil {
            for rows.Next() {
                var id, partNu int
                var md5, instance string
                e1 := rows.Scan(&id, &md5, &instance, &partNu)
                if e1 != nil {
                    return e1
                } else {
                    fi = &bridge.File{Id: id, Md5: md5, Instance: instance, PartNum: partNu}
                }
            }
        }
        return nil
    }, getFullFileSQL1 + addOn, md5)

    if e1 != nil {
        return nil, e1
    }
    // not exists
    if fi == nil {
        return nil, nil
    }

    e2 := dao.Query(func(rows *sql.Rows) error {
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
    }, getFullFileSQL2 + addOn, md5)

    if e2 != nil {
        return nil, e2
    }
    return fi, nil
}


func GetFullFileByFid(fid int, finishFlag int) (*bridge.File, error) {
    var addOn = ""
    if finishFlag == 1 {
        addOn = " and finish=1"
    } else if finishFlag == 0 {
        addOn = " and finish=0"
    }
    var fi *bridge.File
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    // query file
    e1 := dao.Query(func(rows *sql.Rows) error {
        if rows != nil {
            for rows.Next() {
                var id, partNu int
                var md5, instance string
                e1 := rows.Scan(&id, &md5, &instance, &partNu)
                if e1 != nil {
                    return e1
                } else {
                    fi = &bridge.File{Id: id, Md5: md5, Instance: instance, PartNum: partNu}
                }
            }
        }
        return nil
    }, getFullFileSQL11 + addOn, fid)

    if e1 != nil {
        return nil, e1
    }
    // not exists
    if fi == nil {
        return nil, nil
    }

    e2 := dao.Query(func(rows *sql.Rows) error {
        if rows != nil {
            var partsList list.List
            for rows.Next() {
                var size int64
                var md5 string
                e1 := rows.Scan(&md5, &size)
                if e1 != nil {
                    return e1
                } else {
                    var part = &bridge.FilePart{Md5: md5, FileSize: size}
                    partsList.PushBack(part)
                }
            }
            var parsList = make([]bridge.FilePart, partsList.Len())
            index := 0
            for ele := partsList.Front(); ele != nil; ele = ele.Next() {
                parsList[index] = *ele.Value.(*bridge.FilePart)
                index++
            }
            fi.PartNum = partsList.Len()
            fi.Parts = parsList
        }
        return nil
    }, getFullFileSQL21 + addOn, fid)

    if e2 != nil {
        return nil, e2
    }
    return fi, nil
}


// finishSync=1 只查已在本地同步完成的文件
func GetFileByMd5(md5 string, finishFlag int) (*bridge.File, error) {
    var addOn = ""
    if finishFlag == 1 {
        addOn = " and finish=1"
    } else if finishFlag == 0 {
        addOn = " and finish=0"
    }
    var fi *bridge.File
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    // query file
    e1 := dao.Query(func(rows *sql.Rows) error {
        if rows != nil {
            for rows.Next() {
                var id, partNu int
                var md5, instance string
                e1 := rows.Scan(&id, &md5, &instance, &partNu)
                if e1 != nil {
                    return e1
                } else {
                    fi = &bridge.File{Id: id, Md5: md5, Instance: instance, PartNum: partNu}
                }
            }
        }
        return nil
    }, getFullFileSQL1 + addOn, md5)

    if e1 != nil {
        return nil, e1
    }
    return fi, nil
}


func GetFileByFid(fid int, finishFlag int) (*bridge.File, error) {
    var addOn = ""
    if finishFlag == 1 {
        addOn = " and finish=1"
    } else if finishFlag == 0 {
        addOn = " and finish=0"
    }
    var fi *bridge.File
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    // query file
    e1 := dao.Query(func(rows *sql.Rows) error {
        if rows != nil {
            for rows.Next() {
                var id, partNu int
                var md5, instance string
                e1 := rows.Scan(&id, &md5, &instance, &partNu)
                if e1 != nil {
                    return e1
                } else {
                    fi = &bridge.File{Id: id, Md5: md5, Instance: instance, PartNum: partNu}
                }
            }
        }
        return nil
    }, getFullFileSQL11 + addOn, fid)

    if e1 != nil {
        return nil, e1
    }
    return fi, nil
}



func GetSyncId() (int, error) {
    var id = -1
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    e := dao.Query(func(rows *sql.Rows) error {
        if rows != nil {
            for rows.Next() {
                e := rows.Scan(&id)
                if e != nil {
                    return e
                }
            }
        }
        // consider there is no record in db.
        if id == -1 {
            id = 0
            dao.DoTransaction(func(tx *sql.Tx) error {
                return UpdateSyncId(0, tx)
            })
        }
        return nil
    }, getSyncId)
    if e != nil {
        return 0, e
    }
    return id, nil
}

func UpdateSyncId(newId int, tx *sql.Tx) error {
    state, e2 := tx.Prepare(updateSyncId)
    if e2 != nil {
        return e2
    }
    ret, e3 := state.Exec(newId)
    if e3 != nil {
        return e3
    }
    a, _ := ret.RowsAffected()
    logger.Debug("affect:", a)
    return nil
}

// 文件同步到本地修改状态
func UpdateFileStatus(fid int) error {
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    return dao.DoTransaction(func(tx *sql.Tx) error {

        state, e2 := tx.Prepare(updateFileStatusSQL)
        if e2 != nil {
            return e2
        }
        _, e3 := state.Exec(fid)
        if e3 != nil {
            return e3
        }

        state1, e3 := tx.Prepare(finishSyncTaskSQL)
        if e3 != nil {
            return e3
        }
        _, e4 := state1.Exec(fid)
        if e4 != nil {
            return e4
        }
        return nil
    })
}


// storage查询tracker新文件，基于tracker服务器的Id作为起始
func GetFilesBasedOnId(fid int) (*list.List, error) {
    var files list.List
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    // query file
    e1 := dao.Query(func(rows *sql.Rows) error {
        if rows != nil {
            for rows.Next() {
                var id, partNu int
                var md5, instance string
                e1 := rows.Scan(&id, &md5, &instance, &partNu)
                if e1 != nil {
                    return e1
                } else {
                    fi := &bridge.File{Id: id, Md5: md5, Instance: instance, PartNum: partNu}
                    files.PushBack(fi)
                }
            }
        }
        return nil
    }, getFullFileSQL12, fid)

    if e1 != nil {
        return nil, e1
    }

    var addOn bytes.Buffer
    index := 0
    for ele := files.Front(); ele != nil; ele = ele.Next() {
        addOn.Write([]byte(strconv.Itoa(ele.Value.(*bridge.File).Id)))
        if index != files.Len() - 1 {
            addOn.Write([]byte(","))
        }
        index++
    }
    addOn.Write([]byte(") order by d.id"))
    logger.Debug("exec SQL:\n\t" + getFullFileSQL22 + string(addOn.Bytes()))

    e2 := dao.Query(func(rows *sql.Rows) error {
        if rows != nil {
            var tparsList list.List
            for rows.Next() {
                var fid int
                var size int64
                var md5 string
                e1 := rows.Scan(&fid, &md5, &size)
                if e1 != nil {
                    return e1
                } else {
                    var part = &bridge.FilePart{Fid: fid, Md5: md5, FileSize: size}
                    tparsList.PushBack(part)
                }
            }
            for ele1 := files.Front(); ele1 != nil; ele1 = ele1.Next() {
                fi := ele1.Value.(*bridge.File)
                var parsList = make([]bridge.FilePart, fi.PartNum)
                index := 0
                for ele2 := tparsList.Front(); ele2 != nil; ele2 = ele2.Next() {
                    p := *ele2.Value.(*bridge.FilePart)
                    if fi.Id == p.Fid {
                        parsList[index] = p
                        index++
                    }
                }
                fi.Parts = parsList
            }

        }
        return nil
    }, getFullFileSQL22 + string(addOn.Bytes()))
    if e2 != nil {
        return nil, e2
    }
    return &files, nil
}

// 更新一个tracker的同步ID
func UpdateTrackerSyncId(trackerUUID string, id int) error {
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    return dao.DoTransaction(func(tx *sql.Tx) error {
        state, e2 := tx.Prepare(updateTrackerSyncId)
        if e2 != nil {
            return e2
        }
        _, e3 := state.Exec(trackerUUID, id, trackerUUID)
        if e3 != nil {
            return e3
        }
        return nil
    })
}


// 更新一个tracker的本地ID
func UpdateLocalPushId(trackerUUID string, id int) error {
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    return dao.DoTransaction(func(tx *sql.Tx) error {
        state, e2 := tx.Prepare(updateLocalPushId)
        if e2 != nil {
            return e2
        }
        _, e3 := state.Exec(trackerUUID, trackerUUID, trackerUUID, id)
        if e3 != nil {
            return e3
        }
        return nil
    })
}

// 获取tracker的config
func GetTrackerConfig(trackerUUID string) (*bridge.TrackerConfig, error) {
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)

    var config *bridge.TrackerConfig
    err := dao.Query(func(rows *sql.Rows) error {
        if rows != nil {
            for rows.Next() {
                config = &bridge.TrackerConfig{}
                var trackerSyncId, localPushId int
                e := rows.Scan(&trackerSyncId, &localPushId)
                if e != nil {
                    return e
                }
                config.MasterSyncId = trackerSyncId
                config.LocalPushId = localPushId
            }
        }
        return nil
    }, getTrackerConfig, trackerUUID)
    if err != nil {
        return nil, err
    }
    return config, nil
}

func createBatchPartSQL(parts []bridge.FilePart) string {
    var sql bytes.Buffer
    sql.Write([]byte("insert into parts(md5, size) values"))
    index := 0
    for i := range parts {
        sql.Write([]byte("('"))
        sql.Write([]byte(parts[i].Md5))
        sql.Write([]byte("',"))
        sql.Write([]byte(strconv.FormatInt(parts[i].FileSize, 10)))
        sql.Write([]byte(")"))
        if index != len(parts) - 1 {
            sql.Write([]byte("',"))
        }
        index++
    }
    return string(sql.Bytes())
}