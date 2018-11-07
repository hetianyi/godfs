package libservice

import "util/db"

const (
	insertFileSQL       = "insert into files(md5, parts_num, grop, instance, finish) values(?,?,?,?,?)"
	updateFileStatusSQL = "update files set finish=1 where id=?"
	insertPartSQL       = "insert into parts(md5, size) select ?,? where not exists (select 1 from parts where md5=?)"
	insertRelationSQL   = "insert into parts_relation(fid, pid) select ?,? where not exists (select 1 from parts_relation where fid=? and pid=?)"
	fileExistsSQL       = "select id from files a where a.md5 = ? "
	partExistsSQL       = "select id from parts a where a.md5 = ?"

	getLocalPushFiles = `select a.local_push_id from trackers a where a.uuid=?`
	getDownloadFiles  = `select id from files a where a.finish=0 limit ?`
	getFullFileSQL1   = `select b.id, b.md5, grop, b.instance, parts_num from files b where b.md5=? `
	getFullFileSQL11  = `select b.id, b.md5, grop, b.instance, parts_num from files b where b.id=? `
	getFullFileSQL12  = `select b.id, b.md5, grop, b.instance, parts_num from files b where b.id > ? limit 50`
	getFullFileSQL13  = `select b.id, b.md5, grop, b.instance, parts_num from files b where b.id in`
	getFullFileSQL2   = `select d.md5, d.size
                        from files b
                        left join parts_relation c on b.id = c.fid
                        left join parts d on c.pid = d.id where b.md5=?`
	getFullFileSQL21 = `select d.md5, d.size
                        from files b
                        left join parts_relation c on b.id = c.fid
                        left join parts d on c.pid = d.id where b.id=?`
	getFullFileSQL22 = `select b.id, d.md5, d.size
                        from files b
                        left join parts_relation c on b.id = c.fid
                        left join parts d on c.pid = d.id where b.id in(`
	getFullFileSQL23 = `select b.id, d.md5, d.size
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

	statisticQuery = `select * from (
                            (select count(*) files from files a),
                            (select count(*) finish from files a where a.finish = 1),
                            (select case when sum(b.size) is null then 0 else sum(b.size) end disk from parts b)  )`

	insert_web_tracker = `insert into web_trackers(host, port, status, secret, remark, uuid) values(?, ?, ?, ?, ?, ?)`
	delete_web_tracker = `delete from web_trackers where id = ?`

	get_all_web_trackers = `select host, port, status, secret from web_trackers`
	check_web_trackers = `select count(*) from web_trackers a where a.host = ? and a.port = ?`






)

var dbPool *db.DbConnPool

func SetPool(pool *db.DbConnPool) {
	dbPool = pool
}
