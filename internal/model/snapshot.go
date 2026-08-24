package model

import "time"

// OffsetSnapshot 是一条回调的投递位点快照。Sequence 表示该回调
// 已确认投递到的最大事件序号，重启恢复时以快照为准。
type OffsetSnapshot struct {
	CallbackID string
	Sequence   uint64
	UpdatedAt  time.Time
}

// NewOffsetSnapshot 构造位点快照。
func NewOffsetSnapshot(callbackID string, sequence uint64) OffsetSnapshot {
	return OffsetSnapshot{
		CallbackID: callbackID,
		Sequence:   sequence,
		UpdatedAt:  time.Now().UTC(),
	}
}
