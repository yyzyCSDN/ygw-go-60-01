package queue

import (
	"errors"
	"time"

	"hookrelay/internal/model"
)

// EventSnapshot 是事件的可序列化快照，用于服务重启时恢复队列。
// 序号与创建时间都会被保留，保证恢复后的批量窗口与位点语义不变。
type EventSnapshot struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Payload   []byte    `json:"payload"`
	Seq       uint64    `json:"seq"`
	CreatedAt time.Time `json:"created_at"`
}

// Snapshot 导出全部事件快照，按入队顺序排列。
func (q *Queue) Snapshot() []EventSnapshot {
	q.mu.Lock()
	defer q.mu.Unlock()
	result := make([]EventSnapshot, 0, len(q.events))
	for _, event := range q.events {
		result = append(result, EventSnapshot{
			ID:        event.ID,
			Type:      event.Type,
			Payload:   append([]byte(nil), event.Payload...),
			Seq:       event.Seq,
			CreatedAt: event.CreatedAt,
		})
	}
	return result
}

// Restore 用快照重建队列。快照必须按序号升序排列，恢复后队列序号
// 与快照一致，已投递位点之外的窗口仍可正常切分。
func (q *Queue) Restore(snapshots []EventSnapshot) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.events) != 0 {
		return errors.New("queue can only be restored when empty")
	}
	events := make([]*model.Event, 0, len(snapshots))
	byID := make(map[string]*model.Event, len(snapshots))
	var previous uint64
	for index, snapshot := range snapshots {
		if snapshot.ID == "" || snapshot.Seq == 0 {
			return errors.New("restore snapshot must carry id and sequence")
		}
		if index > 0 && snapshot.Seq <= previous {
			return errors.New("restore snapshots must be in ascending sequence order")
		}
		previous = snapshot.Seq
		if _, exists := byID[snapshot.ID]; exists {
			return errors.New("restore snapshot contains a duplicate event id")
		}
		event := &model.Event{
			ID:        snapshot.ID,
			Type:      snapshot.Type,
			Payload:   append([]byte(nil), snapshot.Payload...),
			Seq:       snapshot.Seq,
			CreatedAt: snapshot.CreatedAt,
		}
		events = append(events, event)
		byID[event.ID] = event
	}
	q.events = events
	q.byID = byID
	return nil
}
