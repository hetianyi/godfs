package lib_client

import (
    "lib_common/bridge"
    "strconv"
    "util/logger"
    "net"
    "app"
    "container/list"
    "sync"
)

// server uid : server connections
var connMap map[string]list.List
var activeConnCounter map[string]int
var getLock *sync.Mutex

func init() {
    getLock = new(sync.Mutex)
}

func GetStorageServerUID(server *bridge.Member) string {
    return server.BindAddr + ":" + strconv.Itoa(server.Port) + ":" + server.Group + ":" + server.InstanceId
}


func GetConnBridge(server *bridge.Member) (*bridge.Bridge, error) {
    getLock.Lock()
    defer getLock.Unlock()
    list := connMap[GetStorageServerUID(server)]
    if list.Len() > 0 {
        return list.Remove(list.Front()).(*bridge.Bridge), nil
    }

    return newConnection(server)
}

func newConnection(server *bridge.Member)(*bridge.Bridge, error) {
    logger.Debug("connecting to storage server...")
    con, e := net.Dial("tcp", server.BindAddr + strconv.Itoa(server.Port))
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
    IncreaseActiveConnection(server, 1)
    return connBridge, nil
}

func ReturnConnBridge(*bridge.Bridge) {

}

func IncreaseActiveConnection(server *bridge.Member, value int) {

}





