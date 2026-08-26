// Package queue 提供事件的持久化投递队列（进程内实现）。
// 事件按入队顺序获得连续序号，投递以序号窗口为单位批量拉取。
package queue

import (
	"errors"
	"sync"

	"hookrelay/internal/model"
)

// Queue 是投递队列。events 按入队顺序保存，seqIndex 维护序号到
// 下标的映射，支持按窗口批量读取。
type Queue struct {
	mu     sync.Mutex
	events []*model.Event
	byID   map[string]*model.Event
}

// NewQueue 创建空队列。
func NewQueue() *Queue {
	return &Queue{
		byID: make(map[string]*model.Event),
	}
}

// Enqueue 把事件追加到队尾并分配连续序号，返回该事件的序号。
func (q *Queue) Enqueue(ev *model.Event) (uint64, error) {
	if ev == nil {
		return 0, errors.New("event is required")
	}
	if ev.ID == "" {
		return 0, errors.New("event id is required")
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if _, exists := q.byID[ev.ID]; exists {
		return 0, errors.New("event already enqueued")
	}
	seq := uint64(len(q.events)) + 1
	ev.Seq = seq
	q.events = append(q.events, ev)
	q.byID[ev.ID] = ev
	return seq, nil
}

// Len 返回队列中的事件总数。
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.events)
}

// GetByID 按事件 ID 查询事件。
func (q *Queue) GetByID(id string) (*model.Event, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	ev, ok := q.byID[id]
	return ev, ok
}
