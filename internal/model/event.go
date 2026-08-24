// Package model 定义 HookRelay 的核心领域模型：回调事件、回调注册、
// 投递任务、死信记录与位点快照。所有状态流转都围绕这些模型展开。
package model

import (
	"fmt"
	"time"

	"github.com/cespare/xxhash/v2"
)

// Event 表示一条进入投递中心的回调事件。Seq 是入队序号，由队列分配，
// 用于批量窗口切分与投递位点推进；Payload 是下游回调的请求体。
type Event struct {
	ID        string
	Type      string
	Payload   []byte
	Seq       uint64
	CreatedAt time.Time
}

// NewEvent 构造一条事件。ID 为空时使用事件内容散列生成稳定 ID，
// 保证相同类型与请求体的事件在重复入队前可被识别。
func NewEvent(id, eventType string, payload []byte) *Event {
	if id == "" {
		id = fmt.Sprintf("evt-%016x", xxhash.Sum64(append([]byte(eventType), payload...)))
	}
	return &Event{
		ID:        id,
		Type:      eventType,
		Payload:   append([]byte(nil), payload...),
		CreatedAt: time.Now().UTC(),
	}
}

// Body 返回可安全发送的请求体。空 payload 返回空切片而不是 nil，
// 避免下游序列化与签名逻辑对 nil 切片产生歧义。
func (e *Event) Body() []byte {
	if len(e.Payload) == 0 {
		return []byte{}
	}
	return append([]byte(nil), e.Payload...)
}

// ContentHash 计算事件类型与请求体的 xxhash 摘要，用于幂等去重键。
func (e *Event) ContentHash() uint64 {
	seed := xxhash.Sum64String(e.Type)
	return xxhash.Sum64(append([]byte(fmt.Sprintf("%016x", seed)), e.Payload...))
}

// Clone 深拷贝事件，避免队列与投递链路共享同一底层切片。
func (e *Event) Clone() *Event {
	return &Event{
		ID:        e.ID,
		Type:      e.Type,
		Payload:   append([]byte(nil), e.Payload...),
		Seq:       e.Seq,
		CreatedAt: e.CreatedAt,
	}
}
