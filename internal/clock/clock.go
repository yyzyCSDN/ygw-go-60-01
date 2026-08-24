// Package clock 提供时间源抽象，供签名时间戳、退避调度与超时策略使用。
// 生产环境使用 SystemClock，测试与复现使用 ManualClock 精确控制时间。
package clock

import (
	"sync"
	"time"
)

// Clock 是时间源的统一接口。
type Clock interface {
	Now() time.Time
}

// SystemClock 返回真实系统时间。
type SystemClock struct{}

// Now 返回当前 UTC 时间。
func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}

// ManualClock 是可在测试中手动推进的时间源。
type ManualClock struct {
	mu  sync.Mutex
	now time.Time
}

// NewManualClock 以给定时间初始化手动时钟。
func NewManualClock(at time.Time) *ManualClock {
	return &ManualClock{now: at}
}

// Now 返回当前手动时间。
func (m *ManualClock) Now() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.now
}

// Set 把手动时间拨到指定时刻。
func (m *ManualClock) Set(at time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.now = at
}

// Advance 按给定步长推进手动时间。
func (m *ManualClock) Advance(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.now = m.now.Add(d)
}
