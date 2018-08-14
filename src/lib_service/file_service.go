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

    getLocalPushFiles  = `select a.local_push_id from trackers a where a.uuid=?`
    getDownloadFiles   = `select id from files a where a.finish=0 limit ?`
    getFullFileSQL1  = `select b.id, b.md5, b.instance, parts_num from files b where b.md5=? `
    getFullFileSQL11  = `select b.id, b.md5, b.instance, parts_num from files b where b.id=? `
    getFullFileSQL12  = `select b.id, b.md5, b.instance, parts_num from files b where b.id > ? limit 10`
    getFullFileSQL13  = `select b.id, b.md5, b.instance, parts_num from files b where b.id in`
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
    getFullFileSQL23  = `select b.id, d.md5, d.size
                        from files b
                        left join parts_relation c on b.id = c.fid
                        left join parts d on c.pid = d.id where b.id in`

    updateTrackerSyncId = `replace into trackers(uuid, tracker_sync_id, last_reg_time, local_push_id)
                            values(?, ?, datetime('now','localtime'), (select local_push_id from trackers where uuid = ?))`
    updateLocalPushId = `replace into trackers(uuid, tracker_sync_id, last_reg_time, local_push_id)
                            values(?, 
                            (select tracker_sync_id from trackers where uuid = ?),
                            (select last_reg_time from trackers where uuid = ?), ?)`

    getTrackerConfig = `select tracker_sync_id, local_push_id from trackers where uuid=?`

    confirmLocalInstanceUUID = `replace into sys(key, value) values(
                'uuid',
                (select case when 
                (select count(*) from sys where key = 'uuid') = 0 then ? 
                else (select value from sys where key = 'uuid') end))`

    getLocalInstanceUUID = `select value from sys where key = 'uuid'`

    regStorageClient = `replace into clients(uuid, last_reg_time) values(?, datetime('now','localtime'))`

    existsStorageClient = `select count(*) from clients a where a.uuid = ?`

)

var dbPool *db.DbConnPool

func SetPool(pool *db.DbConnPool) {
    dbPool = pool
}

// get file id by md5
func GetFileId(md5 string, dao *db.DAO) (int, error) {
    var id = 0
    if dao == nil {
        dao = dbPool.GetDB()
        defer dbPool.ReturnDB(dao)
    }
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
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    fid, ee := GetFileId(md5, dao)
    if ee != nil {
        return ee
    }
    // file exists, will skip
    if fid != 0 {
        return nil
    }
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
        return nil
    })
}


// TODO continue here
// store file info which pulled from tracker server.
// return immediate if error occurs
func StorageAddTrackerPulledFile(fis []bridge.File, trackerUUID string) error {
    if fis == nil || len(fis) == 0 {
        return nil
    }
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    for i := range fis {
        fi := fis[i]
        fid, ee := GetFileId(fi.Md5, dao)
        if ee != nil {
            return ee
        }
        // skip if file exists
        if fid != 0 {
            e1 := dao.DoTransaction(func(tx *sql.Tx) error {
                return UpdateTrackerSyncId(trackerUUID, fi.Id, tx)
            })
            if e1 != nil {
                return e1
            }
            continue
        }

        e8 := dao.DoTransaction(func(tx *sql.Tx) error {
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
            if ee := UpdateTrackerSyncId(trackerUUID, fi.Id, tx); ee != nil {
                return ee
            }
            // no need any more.
            //return AddSyncTask(fid, app.TASK_DOWNLOAD_FILE, tx)
            return nil
        })
        if e8 != nil {
            return e8
        }
    }
    return nil
}


// tracker add file
// parts is map[md5]partsize
func TrackerAddFile(meta *bridge.OperationRegisterFileRequest) error {
    var fid int
    var e error
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)

    for i := range meta.Files {
        fi := meta.Files[i]
        fid, e = GetFileId(fi.Md5, dao)
        if e != nil {
            return e
        }
        if fid > 0 {
            continue
        }
        err := dao.DoTransaction(func(tx *sql.Tx) error {
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
            return err
        }
    }

    return nil
}

// mark that a file successfully pushed to tracker.
func FinishLocalFilePushTask(fid int, trackerUUID string) error {
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    return dao.DoTransaction(func(tx *sql.Tx) error {
        state, e2 := tx.Prepare(updateLocalPushId)
        if e2 != nil {
            return e2
        }
        _, e3 := state.Exec(trackerUUID, trackerUUID, trackerUUID, fid)
        if e3 != nil {
            return e3
        }
        return nil
    })
}

// 获取推送到tracker的文件
func GetLocalPushFileTask(tasType int, trackerUUID string) (*bridge.Task, error) {
    var ret = &bridge.Task{FileId: 0, TaskType: tasType}
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    e := dao.Query(func(rows *sql.Rows) error {
        if rows != nil {
            for rows.Next() {
                var fid int
                e1 := rows.Scan(&fid)
                if e1 != nil {
                    return e1
                }
                ret.FileId = fid
            }
        }
        return nil
    }, getLocalPushFiles, trackerUUID)
    if e != nil {
        return nil, e
    }
    return ret, nil
}

// 获取下载任务文件
func GetDownloadFileTask(tasType int) (*list.List, error) {
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
                ret := &bridge.Task{FileId: fid, TaskType: tasType}
                ls.PushBack(ret)
            }
        }
        return nil
    }, getDownloadFiles, 10)
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



func GetFullFileByFids(fids ...int) (*bridge.File, error) {
    if fids == nil || len(fids) == 0 {
        return nil, nil
    }
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    var fi *bridge.File

    // create params
    var buffer bytes.Buffer
    buffer.WriteString("(")
    for i := range fids {
        if i == len(fids) - 1 {
            buffer.WriteString(strconv.Itoa(i))
        } else {
            buffer.WriteString(strconv.Itoa(i))
            buffer.WriteString(",")
        }
    }
    buffer.WriteString(")")

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
    }, getFullFileSQL13 + buffer.String())

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
    }, getFullFileSQL23 + buffer.String())

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


// if file download finish update finish status.
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
func UpdateTrackerSyncId(trackerUUID string, id int, tx *sql.Tx) error {
    if tx == nil {
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
    } else {
        state, e2 := tx.Prepare(updateTrackerSyncId)
        if e2 != nil {
            return e2
        }
        _, e3 := state.Exec(trackerUUID, id, trackerUUID)
        if e3 != nil {
            return e3
        }
    }
    return nil
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



// 更新一个tracker的本地ID
func ConfirmLocalInstanceUUID(uuid string) error {
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    return dao.DoTransaction(func(tx *sql.Tx) error {
        state, e2 := tx.Prepare(confirmLocalInstanceUUID)
        if e2 != nil {
            return e2
        }
        _, e3 := state.Exec(uuid)
        if e3 != nil {
            return e3
        }
        return nil
    })
}

//
func GetLocalInstanceUUID() (string, error) {
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    var uuid string
    err := dao.Query(func(rows *sql.Rows) error {
        if rows != nil {
            for rows.Next() {
                e := rows.Scan(&uuid)
                if e != nil {
                    return e
                }
            }
        }
        return nil
    }, getLocalInstanceUUID)
    if err != nil {
        return "", err
    }
    return uuid, nil
}



func QueryExistsStorageClient(uuid string) bool {
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    var count int
    dao.Query(func(rows *sql.Rows) error {
        if rows != nil {
            for rows.Next() {
                e := rows.Scan(&count)
                if e != nil {
                    return e
                }
            }
        }
        return nil
    }, existsStorageClient)
    if count > 0 {
        return true
    }
    return false
}

// 更新一个tracker的本地ID
func RegisterStorageClient(uuid string) error {
    dao := dbPool.GetDB()
    defer dbPool.ReturnDB(dao)
    return dao.DoTransaction(func(tx *sql.Tx) error {
        state, e2 := tx.Prepare(regStorageClient)
        if e2 != nil {
            return e2
        }
        _, e3 := state.Exec(uuid)
        if e3 != nil {
            return e3
        }
        return nil
    })
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