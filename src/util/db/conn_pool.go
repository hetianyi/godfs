package db

import (
	"container/list"
	"sync"
	"util/logger"
	"errors"
	"fmt"
	"time"
)

type IDbConnPool interface {
	InitPool(connSize int)
	GetDB() *DAO
	ReturnDB(dao *DAO)
}

type DbConnPool struct {
	connSize  int
	dbList    *list.List
	fetchLock *sync.Mutex
	writeLock *sync.Mutex
	listLock  *sync.Mutex
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
	pool.writeLock = new(sync.Mutex)
	pool.listLock = new(sync.Mutex)
	for i := 0; i < poolSize; i++ {
		dao := &DAO{}
		e := dao.InitDB(i)
		if e != nil {
			i--
			continue
		}
		pool.dbList.PushBack(dao)
	}
}

//fetch dao
func (pool *DbConnPool) GetDB() (*DAO, error) {
	pool.fetchLock.Lock()
	defer pool.fetchLock.Unlock()
	waits := 0
	for {
		if pool.dbList.Front() == nil {
			if waits > 30 {
				return nil, errors.New("cannot fetch db connection from pool: wait time out")
			}
			fmt.Print("\n等待数据库连接..........\n")
			logger.Debug("no connection available")
			time.Sleep(time.Millisecond * 100)
			waits++
			continue
		}
		pool.listLock.Lock()
		dao := pool.dbList.Remove(pool.dbList.Front())
		pool.listLock.Unlock()
		logger.Debug("using db connection of index:", dao.(*DAO).index, " left connections:", pool.dbList.Len())
		return dao.(*DAO), nil
	}
}

// return dao
func (pool *DbConnPool) ReturnDB(dao *DAO) {
	if dao != nil {
		logger.Trace("return db connection of index:", dao.index)
		pool.listLock.Lock()
		pool.dbList.PushBack(dao)
		pool.listLock.Unlock()
	} else {
		logger.Error("\n\n\n---------=========error return nil dao=========---------")
	}
}
