package workqueue

import (
	"fmt"
)

func NewWorkQueue(callback func(interface{}) error) *WorkQueue {
	return &WorkQueue{
		queue:    NewQueue(),
		callback: callback,
		workDown: make(chan bool),
	}
}

type WorkQueue struct {
	queue *Queue

	// work func
	callback func(interface{}) error
	workDown chan bool
}

func isClosed(ch <-chan bool) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}

func (wq *WorkQueue) Run() {
	go wq.worker()
}

func (wq *WorkQueue) Enqueue(obj interface{}) {
	if wq.IsShuttingDown() {
		fmt.Printf("queue has been shutdown, fail to enqueue: %v\n", obj)
		return
	}

	wq.queue.Add(obj)
}

// ShutDown 通知 queue shutdown，然后等待 Run 结束
func (wq *WorkQueue) ShutDown() {
	wq.queue.ShutDown()
	<-wq.workDown
}

func (wq *WorkQueue) IsShuttingDown() bool {
	return wq.queue.ShuttingDown()
}

func (wq *WorkQueue) worker() {
	for {
		obj, quit := wq.queue.Get()
		if quit {
			// 如果 wq.workDown 没有数据，则关闭它
			if !isClosed(wq.workDown) {
				close(wq.workDown)
			}
			return
		}

		wq.callback(obj)
		wq.queue.Done(obj)
	}
}
