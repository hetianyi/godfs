package pool

import (
	"container/list"
	"errors"
	"sync"
	"util/common"
	"util/logger"
)

// simple goroutine pool implementation for arrange task and limiting max active task

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
	Exec(task *func()) error // add a new task to task list
	runTask(f func())        // keep trying to fetch task
	taskFinish()             // call when a task is finish
	modifyActiveCount(m int) *int // modify current active task size
}

type Pool struct {
	MaxActiveSize        int        // max active size
	MaxWaitSize          int        // max wait size
	WaitingTaskList      *list.List // wait list
	activeGoroutineMutex sync.Mutex
	reassignTaskMutex    sync.Mutex
	activeGoroutine      int      // current active Goroutine number
	chanNotify           chan int // channel to notify to run next task
}

// initial pool
func initPool(MaxActiveSize int, MaxWaitSize int) (*Pool, error) {
	logger.Trace("initial thread pool...")
	if MaxActiveSize <= 0 {
		return nil, errors.New("ThreadSize must be positive")
	}
	p := &Pool{
		MaxActiveSize:        MaxActiveSize,
		MaxWaitSize:          MaxWaitSize,
		WaitingTaskList:      list.New(),
		activeGoroutineMutex: *new(sync.Mutex),
		reassignTaskMutex:    *new(sync.Mutex),
		activeGoroutine:      0,
		chanNotify:           make(chan int),
	}
	go p.runTask()
	return p, nil
}

// add a new task to pool
func (pool *Pool) Exec(t func()) error {
	logger.Trace("pool get new task")
	// if no free thread found then put the task in 'pool.WaitingList'.
	if pool.WaitingTaskList.Len() >= pool.MaxWaitSize {
		return errors.New("pool is full, can not take any more")
	}
	logger.Trace("push task into waiting list")
	pool.WaitingTaskList.PushBack(t)
	// if active task size is not full, start task immediately
	if pool.modifyActiveCount(0) < pool.MaxActiveSize {
		pool.chanNotify <- 1
	}
	return nil
}

// get current active task size
func (pool *Pool) modifyActiveCount(count int) int {
	pool.activeGoroutineMutex.Lock()
	defer pool.activeGoroutineMutex.Unlock()
	pool.activeGoroutine += count
	return pool.activeGoroutine
}

// call by a finished task for notifying a new task can be run
func (pool *Pool) taskFinish() {
	pool.reassignTaskMutex.Lock()
	defer pool.reassignTaskMutex.Unlock()
	pool.modifyActiveCount(-1)
	logger.Trace("thread free")
	pool.chanNotify <- 0
}

// wait signal and fetch task for running
func (pool *Pool) runTask() {
	for {
		logger.Trace("waiting for new task...")
		<-pool.chanNotify
		waitList := pool.WaitingTaskList
		if waitList.Len() > 0 {
			pool.modifyActiveCount(1)
			nt := &Task{f: waitList.Remove(waitList.Front()).(func())}
			go nt.Run(pool)
		} else {
			logger.Trace("waiting list has no task.")
		}
	}
}

// run a task here
func (t *Task) Run(pool *Pool) {
	logger.Trace("get a new task")
	// set task to nil and notice finally
	defer pool.taskFinish()
	common.Try(func() {
		t.f()
	}, func(i interface{}) {
		logger.Error("exec task error: ", i)
	})
}
