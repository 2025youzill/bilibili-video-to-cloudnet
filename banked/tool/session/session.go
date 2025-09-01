package session

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateSessionID(length int) string {
	if length <= 0 {
		length = 16
	}
	buf := make([]byte, length)
	n, err := rand.Read(buf)
	if err != nil || n != length {
		// 保持签名不变但保证安全：读失败时直接中止执行，交由上层恢复中间件处理
		panic("secure random generator unavailable")
	}
	return hex.EncodeToString(buf)
}
