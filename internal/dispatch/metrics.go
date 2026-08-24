package dispatch

import "sync/atomic"

// Metrics 汇总投递中心的运行指标，供监控页面与统计接口展示。
type Metrics struct {
	delivered atomic.Uint64
	failed    atomic.Uint64
	retried   atomic.Uint64
	dead      atomic.Uint64
	skipped   atomic.Uint64
}

// NewMetrics 创建空指标。
func NewMetrics() *Metrics {
	return &Metrics{}
}

// RecordDelivered 记录一次成功投递。
func (m *Metrics) RecordDelivered() {
	m.delivered.Add(1)
}

// RecordFailed 记录一次失败。
func (m *Metrics) RecordFailed() {
	m.failed.Add(1)
}

// RecordRetried 记录一次进入退避重试。
func (m *Metrics) RecordRetried() {
	m.retried.Add(1)
}

// RecordDead 记录一次进入死信。
func (m *Metrics) RecordDead() {
	m.dead.Add(1)
}

// RecordSkipped 记录一次被去重窗口跳过。
func (m *Metrics) RecordSkipped() {
	m.skipped.Add(1)
}

// Snapshot 返回当前指标的可读快照。
func (m *Metrics) Snapshot() map[string]uint64 {
	return map[string]uint64{
		"delivered": m.delivered.Load(),
		"failed":    m.failed.Load(),
		"retried":   m.retried.Load(),
		"dead":      m.dead.Load(),
		"skipped":   m.skipped.Load(),
	}
}
