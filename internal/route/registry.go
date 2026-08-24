// Package route 维护回调注册表，负责把事件类型路由到目标回调端点。
// 注册表同时提供启用/停用与按类型匹配能力，是投递链路的第一个环节。
package route

import (
	"errors"
	"sort"
	"sync"

	"hookrelay/internal/model"
)

// Registry 保存全部回调注册，按 ID 与事件类型建立索引。
type Registry struct {
	mu        sync.RWMutex
	callbacks map[string]*model.Callback
	byType    map[string][]string
}

// NewRegistry 创建空注册表。
func NewRegistry() *Registry {
	return &Registry{
		callbacks: make(map[string]*model.Callback),
		byType:    make(map[string][]string),
	}
}

// Register 添加或更新一条回调注册。
func (r *Registry) Register(cb *model.Callback) error {
	if cb == nil {
		return errors.New("callback is required")
	}
	if err := cb.Validate(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	previous, existed := r.callbacks[cb.ID]
	if existed {
		r.removeFromType(previous)
	}
	r.callbacks[cb.ID] = cb
	r.byType[cb.EventType] = append(r.byType[cb.EventType], cb.ID)
	return nil
}

// Unregister 移除一条回调注册。
func (r *Registry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cb, ok := r.callbacks[id]
	if !ok {
		return errors.New("callback not found")
	}
	r.removeFromType(cb)
	delete(r.callbacks, id)
	return nil
}

// Get 按 ID 查询回调注册。
func (r *Registry) Get(id string) (*model.Callback, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cb, ok := r.callbacks[id]
	return cb, ok
}

// List 返回全部回调注册，按 ID 排序便于稳定遍历。
func (r *Registry) List() []*model.Callback {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.callbacks))
	for id := range r.callbacks {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]*model.Callback, 0, len(ids))
	for _, id := range ids {
		result = append(result, r.callbacks[id])
	}
	return result
}

// Len 返回注册表内回调数量。
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.callbacks)
}

func (r *Registry) removeFromType(cb *model.Callback) {
	ids := r.byType[cb.EventType]
	for i, id := range ids {
		if id == cb.ID {
			r.byType[cb.EventType] = append(ids[:i], ids[i+1:]...)
			break
		}
	}
	if len(r.byType[cb.EventType]) == 0 {
		delete(r.byType, cb.EventType)
	}
}
