// Package offset 管理每条回调的投递位点。位点是"已确认投递成功"
// 的最大事件序号，投递确认后立即推进，重启时按快照恢复，保证
// 已投递事件不会被重复投递。
package offset

import (
	"errors"
	"sync"

	"hookrelay/internal/model"
)

// Store 保存回调级投递位点。current 是已确认位点，snapshots 是
// 落盘快照；Acknowledge 同时推进 current 与 snapshots，确认即位点
// 前进，重启后按 snapshots 恢复，已投递事件不会被重复投递。
type Store struct {
	mu        sync.Mutex
	current   map[string]uint64
	snapshots map[string]model.OffsetSnapshot
}

// NewStore 创建空位点存储。
func NewStore() *Store {
	return &Store{
		current:   make(map[string]uint64),
		snapshots: make(map[string]model.OffsetSnapshot),
	}
}

// Acknowledge 确认某回调已投递到 seq，立即推进位点：current 与
// snapshots 同步前进，使下一次 Sweep 从 seq 之后切窗、重启落盘
// 快照也反映已确认序号。位点只前进不回退：迟到的旧确认（seq 不
// 大于当前位点）是空操作，不会把位点拉回去。
func (s *Store) Acknowledge(callbackID string, seq uint64) error {
	if callbackID == "" {
		return errors.New("callback id is required")
	}
	if seq == 0 {
		return errors.New("acknowledged sequence must be positive")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if seq <= s.current[callbackID] {
		return nil
	}
	s.current[callbackID] = seq
	s.snapshots[callbackID] = model.NewOffsetSnapshot(callbackID, seq)
	return nil
}

// Current 返回某回调当前已确认的位点。
func (s *Store) Current(callbackID string) uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.current[callbackID]
}

// All 返回全部回调当前位点的副本。
func (s *Store) All() map[string]uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make(map[string]uint64, len(s.current))
	for id, seq := range s.current {
		result[id] = seq
	}
	return result
}

// CallbackCount 返回有投递位点的回调数量。
func (s *Store) CallbackCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.current)
}
