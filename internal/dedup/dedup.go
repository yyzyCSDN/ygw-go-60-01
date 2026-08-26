// Package dedup 实现投递幂等去重。去重键由回调 ID、事件序号与内容
// 散列共同派生；事件进入投递前标记一次，投递成功后必须清理，
// 否则正常重发会被误判为重复而丢弃。
package dedup

import (
	"sync"
	"time"
)

// Entry 是一条去重键记录。Completed 表示对应事件已投递成功，
// 该键随后会被清理；ExpiresAt 用于兜底清理长期未完成键。
type Entry struct {
	Key        string
	CallbackID string
	Sequence   uint64
	ExpiresAt  time.Time
}

// Store 保存去重键。
type Store struct {
	mu   sync.Mutex
	keys map[string]*Entry
}

// NewStore 创建空去重存储。
func NewStore() *Store {
	return &Store{keys: make(map[string]*Entry)}
}

// Mark 为回调与事件序号登记去重键，返回键名。重复标记同一键会刷新
// 过期时间但不会产生第二条记录。
func (s *Store) Mark(callbackID string, sequence uint64, ttl time.Duration) string {
	key := DedupKey(callbackID, sequence)
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.keys[key]
	if !ok {
		entry = &Entry{
			Key:        key,
			CallbackID: callbackID,
			Sequence:   sequence,
		}
		s.keys[key] = entry
	}
	entry.ExpiresAt = time.Now().UTC().Add(ttl)
	return key
}

// Check 判断去重键是否仍然有效：存在、未完成且未过期才算命中。
func (s *Store) Check(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.keys[key]
	if !ok {
		return false
	}
	return !time.Now().UTC().After(entry.ExpiresAt)
}

// Remove 直接删除一个去重键。
func (s *Store) Remove(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.keys, key)
}

// Len 返回当前去重键数量。
func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.keys)
}
