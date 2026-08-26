package queue

import "hookrelay/internal/model"

// Slot 是一次批量投递的窗口描述：Start 是窗口首个事件的序号，
// End 是窗口最后一个事件的序号，Events 是窗口内的克隆事件。
type Slot struct {
	Start  uint64
	End    uint64
	Events []*model.Event
}

// NextSlot 以 after 为已投递位点切出下一个投递窗口。
// 窗口为空时返回 nil，表示队列暂时没有可投递的新事件。
func (q *Queue) NextSlot(after uint64, size int) *Slot {
	events := q.Batch(after, size)
	if len(events) == 0 {
		return nil
	}
	return &Slot{
		Start:  events[0].Seq,
		End:    events[len(events)-1].Seq,
		Events: events,
	}
}
