package pool

import (
    "container/list"
    "errors"
    "sync"
    "util/logger"
    "util/common"
)


func NewPool(ThreadSize int, MaxWaitSize int) (*Pool, error) {
    return initPool(ThreadSize, MaxWaitSize)
}

type ITask interface {
    Run(pool *Pool)
}

type Task struct {
    f func()
}


type IPool interface {
    Exec(task *func()) error     // 添加一个人物到队列
    RunTask(f func())                  // 当一个任务结束后调用此方法从队列中取任务
    taskFinish()          // 线程空闲时由线程调用
    modifyActiveCount() *int         // 线程空闲时由线程调用，返回当前正忙的线程数
}

type Pool struct {
    ThreadSize int              // 线程池大小
    MaxWaitSize int             // 最大等待任务数
    finishChan chan int         // 任务结束时的通知通道
    WaitingTaskList *list.List      // 等待队列
    activeGoroutineMutex sync.Mutex
    reassignTaskMutex sync.Mutex
    activeGoroutine int         // 当前正在运行的Goroutine数量
}




func initPool(ThreadSize int, MaxWaitSize int) (*Pool, error) {
    logger.Info("initial thread pool...")
    if ThreadSize <= 0 {
        return nil, errors.New("ThreadSize must be positive")
    }
    p := &Pool{
        ThreadSize: ThreadSize,
        MaxWaitSize: MaxWaitSize,
        finishChan: make(chan int),
        WaitingTaskList: list.New(),
        activeGoroutineMutex: *new(sync.Mutex),
        reassignTaskMutex: *new(sync.Mutex),
        activeGoroutine: 0,
    }
    return p, nil
}


func (pool *Pool) Exec(t func()) error {
    logger.Info("pool get new task...")
    // if no free thread found then put the task in 'pool.WaitingList'.
    if *pool.modifyActiveCount(0) >= pool.ThreadSize {
        logger.Info("push task into waiting list...")
        if pool.WaitingTaskList.Len() >= pool.MaxWaitSize {
            return errors.New("wait list full, can not take any more")
        }
        pool.WaitingTaskList.PushBack(t)
    } else {
        pool.RunTask(t)
    }
    return nil
}

func (pool *Pool) modifyActiveCount(count int) *int {
    pool.activeGoroutineMutex.Lock()
    defer pool.activeGoroutineMutex.Unlock()
    pool.activeGoroutine += count
    return &pool.activeGoroutine
}

func (pool *Pool) taskFinish() {
    pool.reassignTaskMutex.Lock()
    defer pool.reassignTaskMutex.Unlock()
    runningCount := pool.modifyActiveCount(-1)
    waitList := pool.WaitingTaskList
    logger.Info("thread free and get new task...", waitList.Len(), "|", *runningCount)
    if e := waitList.Front(); e != nil {
        f := waitList.Remove(e).(func())
        pool.RunTask(f)
    }

}

func (pool *Pool) RunTask(f func()) {
    pool.modifyActiveCount(1)
    nt := &Task{f: f}
    go nt.Run(pool)
}


func (t *Task) Run(pool *Pool) {
    logger.Info("get a new task!")
    // set task to nil and notice finally
    defer func() {
        logger.Info("finish task!")
        pool.taskFinish()
    }()
    common.Try(func() {
        t.f()
    }, func(i interface{}) {
        logger.Info("exec task error: ", i)
    })
}






