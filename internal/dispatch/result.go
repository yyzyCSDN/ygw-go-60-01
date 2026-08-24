package dispatch

import (
	"fmt"
	"time"
)

// Result 是一次回调投递的结果，记录状态码、响应体与耗时。
// 2xx 状态码视为成功，其余状态由调用方决定是否重试。
type Result struct {
	StatusCode int
	Body       []byte
	Err        error
	Duration   time.Duration
}

// Success 判断结果是否表示投递成功。
func (r *Result) Success() bool {
	return r.Err == nil && r.StatusCode >= 200 && r.StatusCode < 300
}

// Error 返回失败描述，便于写入任务与死信记录。
func (r *Result) Error() string {
	if r.Err != nil {
		return r.Err.Error()
	}
	if r.StatusCode == 0 {
		return "empty response"
	}
	return fmt.Sprintf("callback returned status %d", r.StatusCode)
}

// BodyText 返回响应体的可读摘要，最多保留前 256 字节。
func (r *Result) BodyText() string {
	if len(r.Body) == 0 {
		return ""
	}
	limit := len(r.Body)
	if limit > 256 {
		limit = 256
	}
	return string(r.Body[:limit])
}
