package retry

import (
	"time"

	"hookrelay/internal/clock"
)

// Scheduler 把退避策略与时间源绑定，产出具体的下次投递时间。
type Scheduler struct {
	policy *Policy
	clock  clock.Clock
}

// NewScheduler 创建调度器。
func NewScheduler(policy *Policy, source clock.Clock) *Scheduler {
	return &Scheduler{policy: policy, clock: source}
}

// NextAttemptAt 计算第 attempt 次重试应在何时发生。
func (s *Scheduler) NextAttemptAt(attempt int, now time.Time) time.Time {
	return now.Add(s.policy.NextDelay(attempt))
}

// Deadline 为单次投递尝试生成带超时的上下文。超时是投递策略的一部分，
// 由调度器统一给出，慢下游的请求不会无限占用投递 worker。
func (s *Scheduler) Deadline(parent contextContext, timeout time.Duration) (contextContext, contextCancel) {
	return contextWithTimeout(parent, timeout)
}
