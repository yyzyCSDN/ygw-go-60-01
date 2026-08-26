package model

import (
	"fmt"
	"time"
)

// DeadLetterStatus 描述死信记录的状态机：pending -> replaying -> resolved。
type DeadLetterStatus string

const (
	DeadLetterPending   DeadLetterStatus = "pending"
	DeadLetterReplaying DeadLetterStatus = "replaying"
	DeadLetterResolved  DeadLetterStatus = "resolved"
)

// DeadLetter 保存重试耗尽后无法投递的事件。Reason 记录最后一次失败原因，
// Payload 保留原始请求体以便人工重放。
type DeadLetter struct {
	ID         string
	EventID    string
	CallbackID string
	Reason     string
	Payload    []byte
	Status     DeadLetterStatus
	CreatedAt  time.Time
	ResolvedAt *time.Time
}

// NewDeadLetter 构造一条死信记录，初始状态为 pending。
func NewDeadLetter(eventID, callbackID, reason string, payload []byte) *DeadLetter {
	return &DeadLetter{
		ID:         fmt.Sprintf("dl-%s-%s", eventID, callbackID),
		EventID:    eventID,
		CallbackID: callbackID,
		Reason:     reason,
		Payload:    append([]byte(nil), payload...),
		Status:     DeadLetterPending,
		CreatedAt:  time.Now().UTC(),
	}
}
