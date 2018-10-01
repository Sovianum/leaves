package executor

import (
	"sync"
	"sync/atomic"
)

func NewThreadExecutor(workerNum int) Executor {
	result := &threadExecutor{
		tasks: sync.Map{},

		notifyChan: make(chan struct{}, 1),
		taskChan:   make(chan *poolTask),
		stopChan:   make(chan struct{}),

		workers: make([]*worker, workerNum),
	}
	for i := 0; i != workerNum; i++ {
		result.workers[i] = &worker{result.taskChan, make(chan struct{})}
	}
	return result
}

type Executor interface {
	Start()
	Stop()
	Execute(func()) TaskResult
}

type TaskResult struct {
	sync *sync.Mutex
}

func (tr *TaskResult) Wait() {
	tr.sync.Lock()
	tr.sync.Unlock()
}

type threadExecutor struct {
	tasks sync.Map

	requestedTaskCnt uint64
	startedTaskCnt   uint64

	notifyChan chan struct{}
	stopChan   chan struct{}
	taskChan   chan *poolTask
	workers    []*worker
}

type poolTask struct {
	task func()
	sync sync.Locker
}

func (tp *threadExecutor) Start() {
	go func() {
		for {
			select {
			case <-tp.stopChan:
				return
			case <-tp.notifyChan:
				for tp.startedTaskCnt != tp.requestedTaskCnt {
					id := tp.startedTaskCnt
					tp.startedTaskCnt++
					iTask, _ := tp.tasks.Load(id)
					tp.tasks.Delete(id)
					task := iTask.(poolTask)
					tp.taskChan <- &task
				}
			}
		}
	}()
	for _, w := range tp.workers {
		go w.start()
	}
}

func (tp *threadExecutor) Stop() {
	tp.stopChan <- struct{}{}
	for _, w := range tp.workers {
		w.stop()
	}
}

func (tp *threadExecutor) Execute(f func()) TaskResult {
	id := atomic.AddUint64(&tp.requestedTaskCnt, 1)
	m := &sync.Mutex{}
	m.Lock()
	tp.tasks.Store(id-1, poolTask{f, m})

	select {
	case tp.notifyChan <- struct{}{}:
	default:
	}
	return TaskResult{m}
}

type worker struct {
	taskChan chan *poolTask
	stopChan chan struct{}
}

func (w *worker) start() {
	for {
		select {
		case <-w.stopChan:
			return
		case task := <-w.taskChan:
			if task != nil {
				task.task()
				task.sync.Unlock()
			}
		}
	}
}

func (w *worker) stop() {
	w.stopChan <- struct{}{}
}
