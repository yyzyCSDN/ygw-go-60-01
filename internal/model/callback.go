package model

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

// Callback 描述一条回调注册：当事件类型与 EventType 匹配时，
// 投递中心会把事件请求体 POST 到 URL，并用 Secret 做 HMAC 签名。
// Timeout 是单次回调的超时上限，MaxAttempts 是重试总次数。
type Callback struct {
	ID          string
	EventType   string
	URL         string
	Secret      string
	Enabled     bool
	Timeout     time.Duration
	MaxAttempts int
	CreatedAt   time.Time
}

// NewCallback 构造一条回调注册，并填充缺省超时与重试次数。
func NewCallback(id, eventType, rawURL, secret string) *Callback {
	return &Callback{
		ID:          id,
		EventType:   eventType,
		URL:         strings.TrimSpace(rawURL),
		Secret:      secret,
		Enabled:     true,
		Timeout:     5 * time.Second,
		MaxAttempts: 3,
		CreatedAt:   time.Now().UTC(),
	}
}

// Validate 校验回调注册的必填字段与 URL 协议。
func (c *Callback) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return errors.New("callback id is required")
	}
	if strings.TrimSpace(c.EventType) == "" {
		return errors.New("event type is required")
	}
	parsed, err := url.Parse(c.URL)
	if err != nil {
		return err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("callback url must use http or https scheme")
	}
	if parsed.Host == "" {
		return errors.New("callback url must include a host")
	}
	if c.Timeout <= 0 {
		return errors.New("callback timeout must be positive")
	}
	if c.MaxAttempts <= 0 {
		return errors.New("callback max attempts must be positive")
	}
	return nil
}

// Matches 判断事件类型是否命中该回调注册。
func (c *Callback) Matches(eventType string) bool {
	return c.Enabled && c.EventType == eventType
}
