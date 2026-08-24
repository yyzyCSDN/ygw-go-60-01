package dispatch

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func hmac256(secret string, body []byte, timestamp string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
