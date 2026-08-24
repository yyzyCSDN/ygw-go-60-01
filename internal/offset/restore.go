package offset

import "hookrelay/internal/model"

// Restore 用落盘快照恢复投递位点。恢复采用单调合并：仅当快照序号
// 大于内存中的当前位点时才覆盖，避免重启加载旧快照导致位点回退。
// 快照按回调 ID 去重，同一回调只保留序号最大的一份。
func (s *Store) Restore(snapshots []model.OffsetSnapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	latest := make(map[string]uint64)
	for _, snap := range snapshots {
		if snap.CallbackID == "" {
			continue
		}
		if snap.Sequence > latest[snap.CallbackID] {
			latest[snap.CallbackID] = snap.Sequence
		}
	}
	for id, seq := range latest {
		if seq > s.current[id] {
			s.current[id] = seq
			s.snapshots[id] = model.NewOffsetSnapshot(id, seq)
		}
	}
	return nil
}
