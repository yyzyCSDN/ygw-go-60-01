package deadletter

import "hookrelay/internal/model"

// PurgeResolved 删除状态为 resolved 的死信记录，最多清理 max 条，
// 防止已处理记录无限累积。返回实际清理数量。
func (s *Store) PurgeResolved(max int) int {
	if max <= 0 {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	removed := 0
	kept := s.letters[:0]
	for _, letter := range s.letters {
		if letter.Status == model.DeadLetterResolved && removed < max {
			delete(s.byID, letter.ID)
			removed++
			continue
		}
		kept = append(kept, letter)
	}
	s.letters = kept
	return removed
}
