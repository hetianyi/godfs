package pool

import (
	"app"
	"container/list"
	"errors"
	"net"
	"strconv"
	"sync"
	"util/logger"
)

// error marks that connection pool is full
var ErrFullConnectionPool = errors.New("connection pool is full")

// simple tcp connection pool
type ClientConnectionPool struct {
	connMap           map[string]*list.List
	activeConnCounter map[string]int
	getLock           *sync.Mutex
	statusLock        *sync.Mutex
	maxConnPerServer  int // 客户端和每个服务建立的最大连接数，web项目中建议设置为和最大线程相同的数量
	totalActiveConn   int
}

// Init maxConnPerServer: max connection for each server
func (pool *ClientConnectionPool) Init(maxConnPerServer int) {
	pool.getLock = new(sync.Mutex)
	pool.statusLock = new(sync.Mutex)
	pool.connMap = make(map[string]*list.List)
	pool.activeConnCounter = make(map[string]int)
	if maxConnPerServer <= 0 || maxConnPerServer > 100 {
		maxConnPerServer = 10
	}
	pool.maxConnPerServer = maxConnPerServer
}

// GetServerKey key = InstanceId@AdvertiseAddr:AdvertisePort!Group
func GetServerKey(server *app.ServerInfo) string {
	host, port := server.GetHostAndPortByAccessFlag()
	return host + ":" + strconv.Itoa(port)
}

// GetConn connection pool has not been implemented.
// for now, one client only support single connection with each storage.
func (pool *ClientConnectionPool) GetConn(server *app.ServerInfo) (net.Conn, error) {
	// TODO lock range is too wide
	pool.getLock.Lock()
	defer pool.getLock.Unlock()
	list := pool.getConnMap(server)
	if list.Len() > 0 {
		logger.Debug("reuse existing connection")
		return list.Remove(list.Front()).(net.Conn), nil
	}
	if pool.IncreaseActiveConnection(server, 0) < pool.maxConnPerServer {
		bridge, e := pool.newConnection(server)
		if e != nil && !server.IsTracker {
			logger.Debug("switch connection flag to advertise address")
			server.SwitchAccessFlag()
			return pool.newConnection(server)
		}
		return bridge, e
	}
	return nil, ErrFullConnectionPool
}

// newConnection only connect but not validate this connection
func (pool *ClientConnectionPool) newConnection(server *app.ServerInfo) (net.Conn, error) {
	host, port := server.GetHostAndPortByAccessFlag()
	logger.Debug("connecting to server " + host + ":" + strconv.Itoa(port) + "...")
	d := net.Dialer{Timeout: app.TCPDialogTimeout}
	conn, e := d.Dial("tcp", host+":"+strconv.Itoa(port))
	if e != nil {
		logger.Debug("error connect to storage server "+host+":"+strconv.Itoa(port), ">", e.Error())
		return nil, e
	}
	pool.IncreaseActiveConnection(server, 1)
	return conn, nil
}

// ReturnConnBridge finish using tcp connection bridge and return it to connection pool.
func (pool *ClientConnectionPool) ReturnConnBridge(server *app.ServerInfo, conn net.Conn) {
	pool.getLock.Lock()
	defer pool.getLock.Unlock()
	connList := pool.getConnMap(server)
	logger.Debug("return health connection:", connList.Len())
	connList.PushBack(conn)
}

// ReturnBrokenConnBridge finish using tcp connection bridge and return it to connection pool.
func (pool *ClientConnectionPool) ReturnBrokenConnBridge(server *app.ServerInfo, conn net.Conn) {
	pool.getLock.Lock()
	defer pool.getLock.Unlock()
	conn.Close()
	pool.IncreaseActiveConnection(server, -1)
	logger.Trace("return broken connection:", pool.connMap[GetServerKey(server)].Len())
}

// IncreaseActiveConnection increase/decrease the connection size mark if a server
func (pool *ClientConnectionPool) IncreaseActiveConnection(server *app.ServerInfo, value int) int {
	pool.statusLock.Lock()
	defer pool.statusLock.Unlock()
	pool.totalActiveConn += value
	oldVal := pool.activeConnCounter[GetServerKey(server)]
	pool.activeConnCounter[GetServerKey(server)] = oldVal + value
	return oldVal + value
}

// getConnMap get server connection managed mapping
func (pool *ClientConnectionPool) getConnMap(server *app.ServerInfo) *list.List {
	uid := GetServerKey(server)
	connList := pool.connMap[uid]
	if connList == nil {
		connList = new(list.List)
	}
	pool.connMap[uid] = connList
	return connList
}
