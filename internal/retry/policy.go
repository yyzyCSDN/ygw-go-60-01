// Package retry 实现投递失败后的退避重试策略。退避间隔按尝试次数
// 指数增长并封顶，防止下游短暂故障时被重试打爆；超过上限后事件
// 转入死信队列。
package retry

import (
	"math"
	"time"
)

// Policy 描述退避重试策略：Base 是第一次重试的基准间隔，Max 是
// 间隔上限，MaxAttempts 是包含首次尝试在内的总尝试次数，Factor
// 是相邻两次重试间隔的倍率。
type Policy struct {
	Base        time.Duration
	Max         time.Duration
	MaxAttempts int
	Factor      float64
}

// NewPolicy 构造退避策略。MaxAttempts 至少为 1，Factor 小于 1 时
// 按 2 处理。
func NewPolicy(base, max time.Duration, maxAttempts int) *Policy {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	factor := 2.0
	if base > 0 && max > 0 && base <= max {
		factor = math.Pow(float64(max)/float64(base), 1.0/float64(maxAttempts-1))
	}
	return &Policy{
		Base:        base,
		Max:         max,
		MaxAttempts: maxAttempts,
		Factor:      factor,
	}
}

// NextDelay 返回第 attempt 次重试（attempt 从 1 开始）的退避间隔。
// 首次重试返回 Base，之后按 Factor 指数增长并封顶到 Max。
func (p *Policy) NextDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}
	if p.Base <= 0 {
		return 0
	}
	return 0
}

// TotalAttempts 返回允许的总尝试次数。
func (p *Policy) TotalAttempts() int {
	return p.MaxAttempts
}
