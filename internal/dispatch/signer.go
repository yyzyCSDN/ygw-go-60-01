// Package dispatch 负责回调投递执行：构造签名请求、调用下游端点、
// 处理成功确认与失败重试/死信流转。
package dispatch

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"hookrelay/internal/clock"
)

// Signer 为回调请求生成 HMAC 签名。签名覆盖时间戳与请求体，
// 下游验签时用同一算法比对，防止请求被篡改。
type Signer struct {
	clock  clock.Clock
	secret []byte
}

// NewSigner 创建签名器。
func NewSigner(source clock.Clock, secret string) *Signer {
	return &Signer{clock: source, secret: []byte(secret)}
}

// Sign 基于最新时间戳与请求体生成签名，返回签名与签名使用的时间戳。
// 每次调用都重新读取时钟，物理时钟回拨或长时间运行后仍使用当前时间。
func (s *Signer) Sign(body []byte) (string, time.Time) {
	stamp := s.clock.Now()
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(strconv.FormatInt(stamp.Unix(), 10)))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil)), stamp
}

// VerifySignature 校验签名是否与给定时间戳和请求体匹配。
func VerifySignature(secret string, body []byte, timestamp, signature string) error {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	if signature != expected {
		return errors.New("signature mismatch")
	}
	return nil
}

// HeaderNames 返回签名相关请求头名称，供客户端与服务端共用。
func HeaderNames() (timestampHeader, signatureHeader string) {
	return "X-Hook-Timestamp", "X-Hook-Signature"
}

// FormatTimestamp 把时间转换为签名头使用的秒级字符串。
func FormatTimestamp(t time.Time) string {
	return fmt.Sprintf("%d", t.Unix())
}
