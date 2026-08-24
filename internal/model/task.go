package model

import (
	"fmt"
	"time"
)

// TaskStatus 描述投递任务的状态机取值：
// queued -> delivering -> retrying -> delivered / dead。
type TaskStatus string

const (
	TaskQueued     TaskStatus = "queued"
	TaskDelivering TaskStatus = "delivering"
	TaskRetrying   TaskStatus = "retrying"
	TaskDelivered  TaskStatus = "delivered"
	TaskDead       TaskStatus = "dead"
)

// DeliveryTask 表示一个事件对某条回调的一次投递尝试记录。
// Attempt 记录已经尝试的次数，NextAttemptAt 是退避后的下次投递时间。
type DeliveryTask struct {
	ID            string
	EventID       string
	CallbackID    string
	Status        TaskStatus
	Attempt       int
	NextAttemptAt time.Time
	LastError     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewTask 构造初始的投递任务，状态为 queued。
func NewTask(eventID, callbackID string) *DeliveryTask {
	now := time.Now().UTC()
	return &DeliveryTask{
		ID:         fmt.Sprintf("task-%s-%s", eventID, callbackID),
		EventID:    eventID,
		CallbackID: callbackID,
		Status:     TaskQueued,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// MarkDelivering 把任务状态切换到投递中。
func (t *DeliveryTask) MarkDelivering(now time.Time) {
	t.Status = TaskDelivering
	t.UpdatedAt = now
}

// MarkDelivered 记录投递成功。
func (t *DeliveryTask) MarkDelivered(now time.Time) {
	t.Status = TaskDelivered
	t.UpdatedAt = now
}

// MarkRetrying 记录一次失败并进入退避重试。
func (t *DeliveryTask) MarkRetrying(next time.Time, reason string, now time.Time) {
	t.Attempt++
	t.Status = TaskRetrying
	t.NextAttemptAt = next
	t.LastError = reason
	t.UpdatedAt = now
}

// MarkDead 记录重试耗尽进入死信。
func (t *DeliveryTask) MarkDead(reason string, now time.Time) {
	t.Attempt++
	t.Status = TaskDead
	t.LastError = reason
	t.UpdatedAt = now
}

// DueAt 判断任务在给定时刻是否已经可以再次投递。
func (t *DeliveryTask) DueAt(now time.Time) bool {
	return t.Status == TaskRetrying && !t.NextAttemptAt.After(now)
}
