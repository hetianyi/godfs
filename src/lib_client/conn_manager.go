package lib_client

import (
    "lib_common/bridge"
    "strconv"
    "util/logger"
    "net"
    "app"
    "container/list"
    "sync"
    "errors"
)

var MAX_CONN_EXCEED_ERROR = errors.New("max client connection reached")

type ConnPool interface {
    Init()
    GetConnBridge(server *bridge.Member) (*bridge.Bridge, error)
    newConnection(server *bridge.Member)(*bridge.Bridge, error)
    ReturnConnBridge(server *bridge.Member, connBridge *bridge.Bridge)
    IncreaseActiveConnection(server *bridge.Member, value int)
}

type ClientConnectionPool struct {
    connMap map[string]list.List
    activeConnCounter map[string]int
    getLock *sync.Mutex
    statusLock *sync.Mutex
    maxConnPerServer int    // 客户端和每个服务建立的最大连接数，web项目中建议设置为和最大线程相同的数量
}

// maxConnPerServer: 每个服务的最大连接数
func (pool *ClientConnectionPool) Init(maxConnPerServer int) {
    pool.getLock = new(sync.Mutex)
    pool.statusLock = new(sync.Mutex)
    pool.connMap = make(map[string]list.List)
    pool.activeConnCounter = make(map[string]int)
    if maxConnPerServer <= 0 || maxConnPerServer > 100 {
        maxConnPerServer = 10
    }
    pool.maxConnPerServer = maxConnPerServer
}

func GetStorageServerUID(server *bridge.Member) string {
    return server.BindAddr + ":" + strconv.Itoa(server.Port) + ":" + server.Group + ":" + server.InstanceId
}

// connection pool has not been implemented.
// for now, one client only support single connection with each storage.
func (pool *ClientConnectionPool) GetConnBridge(server *bridge.Member) (*bridge.Bridge, error) {
    pool.getLock.Lock()
    defer pool.getLock.Unlock()
    list := pool.connMap[GetStorageServerUID(server)]
    if list.Len() > 0 {
        return list.Remove(list.Front()).(*bridge.Bridge), nil
    }
    if pool.IncreaseActiveConnection(server, 0) < pool.maxConnPerServer {
        return pool.newConnection(server)
    }
    return nil, MAX_CONN_EXCEED_ERROR
}

func (pool *ClientConnectionPool) newConnection(server *bridge.Member)(*bridge.Bridge, error) {
    logger.Debug("connecting to storage server...")
    con, e := net.Dial("tcp", server.BindAddr + ":" + strconv.Itoa(server.Port))
    if e != nil {
        logger.Error(e)
        return nil, e
    }
    connBridge := bridge.NewBridge(con)
    e1 := connBridge.ValidateConnection(app.SECRET)
    if e1 != nil {
        connBridge.Close()
        return nil, e1
    }
    logger.Debug("successful validate connection:", e1)
    pool.IncreaseActiveConnection(server, 1)
    return connBridge, nil
}

// finish using tcp connection bridge and return it to connection pool.
func (pool *ClientConnectionPool) ReturnConnBridge(server *bridge.Member, connBridge *bridge.Bridge) {
    pool.getLock.Lock()
    defer pool.getLock.Unlock()
    list := pool.connMap[GetStorageServerUID(server)]
    list.PushBack(connBridge)
}
// finish using tcp connection bridge and return it to connection pool.
func (pool *ClientConnectionPool) ReturnBrokenConnBridge(server *bridge.Member, connBridge *bridge.Bridge) {
    pool.getLock.Lock()
    defer pool.getLock.Unlock()
    connBridge.Close()
    pool.IncreaseActiveConnection(server, -1)
}

func (pool *ClientConnectionPool) IncreaseActiveConnection(server *bridge.Member, value int) int {
    pool.statusLock.Lock()
    defer pool.statusLock.Unlock()
    oldVal := pool.activeConnCounter[GetStorageServerUID(server)]
    pool.activeConnCounter[GetStorageServerUID(server)] = oldVal + value
    return oldVal + value
}





