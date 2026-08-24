package queue

import "hookrelay/internal/model"

// Batch 返回序号落在 (after, after+size] 区间内的事件切片。
// 窗口是左开右闭的：after 表示已投递到的序号，下一个窗口从 after+1 开始，
// 保证连续两个窗口不重叠、不漏事件。
func (q *Queue) Batch(after uint64, size int) []*model.Event {
	if size <= 0 {
		return nil
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	start := after
	end := after + uint64(size)
	begin := q.lowerBound(start)
	var result []*model.Event
	for i := begin; i < len(q.events); i++ {
		if q.events[i].Seq > end {
			break
		}
		result = append(result, q.events[i].Clone())
	}
	return result
}

// lowerBound 返回第一个序号不小于 target 的事件下标。
func (q *Queue) lowerBound(target uint64) int {
	left, right := 0, len(q.events)
	for left < right {
		mid := (left + right) / 2
		if q.events[mid].Seq < target {
			left = mid + 1
		} else {
			right = mid
		}
	}
	return left
}
