package workqueue

import (
	"fmt"
	"sync"
)

func NewQueue() *Queue {
	return &Queue{
		dirty:        NewSet(),
		processing:   NewSet(),
		cond:         sync.NewCond(&sync.Mutex{}),
		shuttingDown: false,
	}
}

// 如果元素正在处理的时候又被添加进来了，不直接添加到 queue 中
// 而是先添加到 dirty 中，等元素被处理完了之后再添加到 queue 中
// 这样保证了：
// 1. 正在处理的元素不重复
// 2. 队列中的元素不重复
// 3. 正在处理元素时，多次添加重复的元素只能保存一份
// 但是这么做的意义在哪儿呢？
type Queue struct {
	// 顺序的元素队列，在 queue 中的元素一定在 dirty 而且不在 processing 中
	queue []t

	// 所有要被处理的元素集合
	dirty *Set

	// 正在处理的元素集合，一定不在 queue 中，但是有可能在 dirty 中
	processing *Set

	cond *sync.Cond

	shuttingDown bool
}

func (q *Queue) Add(item interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if q.shuttingDown {
		return
	}
	if q.dirty.Has(item) {
		return
	}

	q.dirty.Insert(item)

	// 如果当前元素正在被处理，则不重复向 queue 中添加
	if q.processing.Has(item) {
		fmt.Printf("element has processing %v\n", item)
		return
	}

	q.queue = append(q.queue, item)
	q.cond.Signal()
}

func (q *Queue) Len() int {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return len(q.queue)
}

func (q *Queue) Get() (item interface{}, shutdown bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	// 队列为空，并且没通知结束 -> 死等
	for len(q.queue) == 0 && !q.shuttingDown {
		q.cond.Wait()
	}

	// Wait 被唤醒有两种情况：
	// 1. Add 元素之后，Signal
	// 2. 外界 ShutDown Queue
	// 既然现在队列依旧为空，那么说明一定是第二种情况
	if len(q.queue) == 0 {
		return nil, true
	}

	item, q.queue = q.queue[0], q.queue[1:]

	q.processing.Insert(item)
	q.dirty.Delete(item)

	return item, false
}

// Done 用来表示一个元素被处理完了
// 如果 dirty 中有则表示在处理期间又被添加了，则需要再次添加到 queue 中
func (q *Queue) Done(item interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	q.processing.Delete(item)

	if q.dirty.Has(item) {
		q.queue = append(q.queue, item)
		q.cond.Signal()
	}
}

func (q *Queue) ShutDown() {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	q.shuttingDown = true
	q.cond.Broadcast() // 唤醒所有等待的协程
}

func (q *Queue) ShuttingDown() bool {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	return q.shuttingDown
}
