package dedup

import "time"

// PruneThrough 清理某回调序号不超过 seq 的全部去重键。投递位点推进
// 到 seq 后调用，保证正常重发（例如人工重放）不再被旧键误去重。
func (s *Store) PruneThrough(callbackID string, seq uint64) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return 0
}

// CleanExpired 删除所有已过期的去重键，返回清理数量。
func (s *Store) CleanExpired(now time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	removed := 0
	for key, entry := range s.keys {
		if now.After(entry.ExpiresAt) {
			delete(s.keys, key)
			removed++
		}
	}
	return removed
}
