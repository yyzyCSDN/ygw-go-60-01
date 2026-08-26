package route

import (
	"sort"

	"hookrelay/internal/model"
)

// Match 返回与事件类型匹配且处于启用状态的全部回调，按 ID 排序。
// 投递中心对每个命中回调各生成一条投递任务。
func (r *Registry) Match(eventType string) []*model.Callback {
	r.mu.RLock()
	ids := append([]string(nil), r.byType[eventType]...)
	r.mu.RUnlock()
	sort.Strings(ids)
	result := make([]*model.Callback, 0, len(ids))
	for _, id := range ids {
		cb, ok := r.Get(id)
		if ok && cb.Matches(eventType) {
			result = append(result, cb)
		}
	}
	return result
}
