package pool

import (
    "container/list"
    "errors"
    "sync"
    "util/logger"
    "util/common"
)


func NewPool(MaxActiveSize int, MaxWaitSize int) (*Pool, error) {
    return initPool(MaxActiveSize, MaxWaitSize)
}

type ITask interface {
    Run(pool *Pool)
}

type Task struct {
    f func()
}


type IPool interface {
    Exec(task *func()) error     // add a new task to task list
    RunTask(f func())                  // 当一个任务结束后调用此方法从队列中取任务
    taskFinish()          // 线程空闲时由线程调用
    modifyActiveCount() *int         // 线程空闲时由线程调用，返回当前正忙的线程数
}

type Pool struct {
    MaxActiveSize int              // 最大同时激活任务数
    MaxWaitSize int             // 最大等待任务数
    finishChan chan int         // 任务结束时的通知通道
    WaitingTaskList *list.List      // 等待队列
    activeGoroutineMutex sync.Mutex
    reassignTaskMutex sync.Mutex
    activeGoroutine int         // 当前正在运行的Goroutine数量
    chanNotify chan int         // chan通道，通知新的任务到来
}




func initPool(MaxActiveSize int, MaxWaitSize int) (*Pool, error) {
    logger.Info("initial thread pool...")
    if MaxActiveSize <= 0 {
        return nil, errors.New("ThreadSize must be positive")
    }
    p := &Pool{
        MaxActiveSize: MaxActiveSize,
        MaxWaitSize: MaxWaitSize,
        finishChan: make(chan int),
        WaitingTaskList: list.New(),
        activeGoroutineMutex: *new(sync.Mutex),
        reassignTaskMutex: *new(sync.Mutex),
        activeGoroutine: 0,
        chanNotify: make(chan int),
    }
    go p.RunTask()
    return p, nil
}


func (pool *Pool) Exec(t func()) error {
    logger.Info("pool get new task")
    // if no free thread found then put the task in 'pool.WaitingList'.
    if pool.WaitingTaskList.Len() + pool.modifyActiveCount(0) >= pool.MaxActiveSize + pool.MaxWaitSize {
        return errors.New("wait list full, can not take any more")
    }
    logger.Info("push task into waiting list")
    pool.WaitingTaskList.PushBack(t)
    pool.chanNotify <- 1
    return nil
}

func (pool *Pool) modifyActiveCount(count int) int {
    pool.activeGoroutineMutex.Lock()
    defer pool.activeGoroutineMutex.Unlock()
    pool.activeGoroutine += count
    return pool.activeGoroutine
}

func (pool *Pool) taskFinish() {
    pool.reassignTaskMutex.Lock()
    defer pool.reassignTaskMutex.Unlock()
    pool.modifyActiveCount(-1)
    logger.Info("thread free")
    pool.chanNotify <- 0
}

func (pool *Pool) RunTask() {
    for {
        logger.Info("waiting for new task...")
        <- pool.chanNotify
        logger.Info("?????????")
        waitList := pool.WaitingTaskList
        if waitList.Len() > 0 {
            pool.modifyActiveCount(1)
            nt := &Task{f: waitList.Remove(waitList.Front()).(func())}
            go nt.Run(pool)
        } else {
            logger.Info("waiting list has no task.")
        }
    }
}


func (t *Task) Run(pool *Pool) {
    logger.Info("get a new task")
    // set task to nil and notice finally
    defer func() {
        pool.taskFinish()
    }()
    common.Try(func() {
        t.f()
    }, func(i interface{}) {
        logger.Error("exec task error: ", i)
    })
}






