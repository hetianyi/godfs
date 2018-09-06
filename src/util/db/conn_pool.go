package db

import (
    "container/list"
    "sync"
    "time"
    "util/logger"
)

type IDbConnPool interface {
    InitPool(connSize int)
    GetDB() *DAO
    ReturnDB(dao *DAO)
}

type DbConnPool struct {
    connSize int
    dbList *list.List
    fetchLock *sync.Mutex
}

func NewPool(poolSize int) *DbConnPool {
    pool := &DbConnPool{}
    pool.InitPool(poolSize)
    return pool
}

// init db connection pool
func (pool *DbConnPool) InitPool(poolSize int) {
    pool.connSize = poolSize
    pool.dbList = list.New()
    pool.fetchLock = new(sync.Mutex)
    for i := 0; i < poolSize; i++ {
        dao := &DAO{}
        dao.InitDB(i)
        pool.dbList.PushBack(dao)
    }
}

//fetch dao
func (pool *DbConnPool) GetDB() *DAO {
    pool.fetchLock.Lock()
    defer pool.fetchLock.Unlock()
    for {
        dao := pool.dbList.Remove(pool.dbList.Front())
        if dao == nil {
            logger.Debug("no connection available")
            time.Sleep(time.Millisecond * 100)
        } else {
            logger.Trace("using db connection of index:", dao.(*DAO).index)
            return dao.(*DAO)
        }
    }
}

// return dao
func (pool *DbConnPool) ReturnDB(dao *DAO) {
    if dao != nil {
        logger.Trace("return db connection of index:", dao.index)
        pool.dbList.PushBack(dao)
    }
}

