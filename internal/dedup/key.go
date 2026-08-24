package dedup

import (
	"fmt"

	"github.com/cespare/xxhash/v2"
)

// DedupKey 由回调 ID 与事件序号派生去重键。使用 xxhash 压缩为
// 固定长度十六进制串，避免以原始字符串作为键占用过多内存。
func DedupKey(callbackID string, sequence uint64) string {
	return fmt.Sprintf("dk-%016x", xxhash.Sum64(append([]byte(callbackID), byte(sequence>>56), byte(sequence>>48), byte(sequence>>40), byte(sequence>>32), byte(sequence>>24), byte(sequence>>16), byte(sequence>>8), byte(sequence))))
}
