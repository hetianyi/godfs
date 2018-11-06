package libservice

const (
    insert_web_tracker = `insert into web_trackers(host, port, status, remark, uuid)
                            values(?, ?, ?, ?, ?)`
)

func AddWebTracker() {
    
}


