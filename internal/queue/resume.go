package queue

import "hookrelay/internal/model"

// ResumeFrom 返回序号大于 seq 的全部剩余事件，用于服务重启后
// 从未确认位点续投。seq 来自 offset 恢复后的最新位点，续投必须
// 从 seq+1 开始，保证不重复投递已确认事件。
func (q *Queue) ResumeFrom(seq uint64) []*model.Event {
	q.mu.Lock()
	defer q.mu.Unlock()
	begin := q.lowerBound(seq)
	result := make([]*model.Event, 0, len(q.events)-begin)
	for i := begin; i < len(q.events); i++ {
		result = append(result, q.events[i].Clone())
	}
	return result
}
