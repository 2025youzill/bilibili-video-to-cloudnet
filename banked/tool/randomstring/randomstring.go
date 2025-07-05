package randomstring

import (
	"math/rand"
	"sync"
	"time"
)

var (
	globalRand *rand.Rand
	initOnce   sync.Once
)

func init() {
	// 利用 init 函数确保初始化（线程安全）
	initGlobalRand()
}

func initGlobalRand() {
	globalRand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[globalRand.Intn(len(charset))]
	}
	return string(b)
}