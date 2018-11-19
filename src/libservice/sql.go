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



	// web manager
	insert_web_tracker = `insert into web_trackers(host, port, status, secret, remark) values(?, ?, ?, ?, ?)`
	update_web_tracker_status = `update web_trackers set status = ? where id = ?`
	get_all_web_trackers = `select id, host, port, status, remark, secret from web_trackers`
	get_exists_trackers = `select id, host, port, status, remark, secret from web_trackers where status != ?`
	custom_get_web_tracker = `select id, host, port, status, remark, secret from web_trackers a where `
	check_web_trackers = `select count(*) from web_trackers a where a.host = ? and a.port = ?`


	insert_web_storage = `insert into web_storages(host, port, status, tracker, uuid, total_files, grop, instance_id, http_port, 
													http_enable, start_time, downloads, uploads, disk, read_only, ioin, ioout) 
							values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	update_web_storage = `update web_storages set host = ?, port = ?, status = ?, total_files = ?, grop = ?, instance_id = ?, 
                            http_port = ?, http_enable = ?, start_time = ?, downloads = ?, uploads = ?, disk = ?, read_only = ?, ioin = ?, ioout = ?
                                    where uuid = ? and tracker = ?`
	custom_get_web_storages = `select a.id, a.host, a.port, a.status, a.tracker, a.uuid, a.total_files, a.grop, a.instance_id, 
								a.http_port, a.http_enable, a.start_time, a.downloads, a.uploads, a.disk, a.read_only, a.ioin, a.ioout from web_storages a where `
	get_web_storage_status = `select a.id, a.status, a.uuid from web_storages a where a.tracker = ?`
	mark_dead_web_storage = `update web_storages set status = ? where tracker = ? and uuid in(?) `
	check_web_storages = `select count(*) from web_storages a where a.uuid = ? and a.tracker = ?`

	insert_storage_log = `insert into web_storage_logs(storage, log_time, ioin, ioout, disk, mem, download, upload) values(?, ?, ?, ?, ?, ?, ?, ?)`



)

var dbPool *db.DbConnPool

func SetPool(pool *db.DbConnPool) {
	dbPool = pool
}
