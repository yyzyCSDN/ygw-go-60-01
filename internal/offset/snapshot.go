package offset

import (
	"sort"

	"hookrelay/internal/model"
)

// Snapshot 返回当前已确认位点的快照列表，按回调 ID 排序。
// 服务正常关闭与周期落盘时调用，快照内容必须与 Acknowledge
// 推进后的位点一致。
func (s *Store) Snapshot() []model.OffsetSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]string, 0, len(s.snapshots))
	for id := range s.snapshots {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]model.OffsetSnapshot, 0, len(ids))
	for _, id := range ids {
		result = append(result, s.snapshots[id])
	}
	return result
}
