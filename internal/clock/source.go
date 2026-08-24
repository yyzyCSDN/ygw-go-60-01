package clock

import "time"

// Source 是投递中心组件共享的时间读取器，统一从这里取当前时间。
// 组件构造时注入一次，之后每次调用 Now 都读取底层时钟的最新值，
// 避免组件内缓存旧时间戳导致跨组件行为不一致。
type Source struct {
	clock Clock
}

// NewSource 包装一个时钟源。
func NewSource(c Clock) *Source {
	return &Source{clock: c}
}

// Now 读取底层时钟的最新时间。
// 不缓存构造时的读数：物理时钟回拨或长时间运行后，调用方仍拿到当前
// 时间，签名时间戳与下游的时效校验保持一致。
func (s *Source) Now() time.Time {
	return s.clock.Now()
}
