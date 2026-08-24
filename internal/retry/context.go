package retry

import (
	"context"
	"time"
)

// 别名避免与标准库 context 直接耦合，便于统一更换实现。
type contextContext = context.Context
type contextCancel = context.CancelFunc

func contextWithTimeout(parent contextContext, timeout time.Duration) (contextContext, contextCancel) {
	if timeout <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, timeout)
}
