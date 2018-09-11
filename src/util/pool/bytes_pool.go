package pool

import (
    "container/list"
    "sync"
    "util/logger"
)

type BytesPool struct {
    bufferMap map[int]*list.List
    maxSize int
    lock *sync.Mutex
}

//
func NewBytesPool(maxSize int) *BytesPool {
    // by default, cache 50 []byte
    if maxSize <= 0 {
        maxSize = 10
    }
    return &BytesPool{maxSize: maxSize, bufferMap: make(map[int]*list.List), lock: new(sync.Mutex)}
}

func (pool *BytesPool) Apply(length int) []byte {
    pool.lock.Lock()
    defer pool.lock.Unlock()
    ls := pool.bufferMap[length]
    if ls == nil {
        ls = list.New()
        pool.bufferMap[length] = ls
    }
    if ls.Front() != nil {
        logger.Debug("use buffered bytes of length:", length)
        return ls.Remove(ls.Front()).([]byte)
    }
    logger.Debug("create bytes buffer of length:", length)
    return make([]byte, length)
}


func (pool *BytesPool) Recycle(buffer []byte) {
    pool.lock.Lock()
    defer pool.lock.Unlock()
    ls := pool.bufferMap[len(buffer)]
    if ls == nil {
        ls = list.New()
        pool.bufferMap[len(buffer)] = ls
    }
    // pool is full, discard
    if ls.Len() >= pool.maxSize {
        // for small bytes buffer, try to cache more.
        if len(buffer) > 1024 || ls.Len() > pool.maxSize * 100 {
            logger.Debug("discard bytes buffer of length:", len(buffer))
            return
        }
    }
    logger.Debug("cache bytes buffer of length:", len(buffer))
    ls.PushBack(buffer)
}


