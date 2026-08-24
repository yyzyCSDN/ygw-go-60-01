package route

import (
	"sort"
	"sync"
)

// HealthTracker 记录每条回调的投递健康度：成功与失败次数、最近一次
// 结果时间。监控页面与统计接口据此展示下游端点状态。
type HealthTracker struct {
	mu    sync.Mutex
	stats map[string]*callbackHealth
}

type callbackHealth struct {
	Success    int64
	Failure    int64
	LastResult string
	LastSeen   int64
}

// NewHealthTracker 创建健康跟踪器。
func NewHealthTracker() *HealthTracker {
	return &HealthTracker{stats: make(map[string]*callbackHealth)}
}

// Record 记录一次投递结果。ok 为 true 表示成功，否则记录失败。
func (h *HealthTracker) Record(callbackID string, ok bool, note string, unixSeconds int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	entry, exists := h.stats[callbackID]
	if !exists {
		entry = &callbackHealth{}
		h.stats[callbackID] = entry
	}
	if ok {
		entry.Success++
	} else {
		entry.Failure++
	}
	entry.LastResult = note
	entry.LastSeen = unixSeconds
}

// List 返回全部回调健康摘要，按回调 ID 排序。
func (h *HealthTracker) List() []map[string]any {
	h.mu.Lock()
	defer h.mu.Unlock()
	ids := make([]string, 0, len(h.stats))
	for id := range h.stats {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		entry := h.stats[id]
		result = append(result, map[string]any{
			"callback_id": id,
			"success":     entry.Success,
			"failure":     entry.Failure,
			"last_result": entry.LastResult,
			"last_seen":   entry.LastSeen,
		})
	}
	return result
}
