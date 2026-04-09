// Copyright (c) 2025 Youzill
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package session

import (
	redis_pool "bvtc/tool/pool"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
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

func SetNewCookie(cookieFile string, sid string) error {
	rdb := redis_pool.GetRdb()
	rctx := redis_pool.GetRctx()
	if sid == "" {
		return fmt.Errorf("sid is empty")
	}
	if rdb == nil {
		return fmt.Errorf("redis client is nil")
	}
	key := "session:" + sid
	if err := rdb.HSet(rctx, key, map[string]interface{}{
		"cookieFile": cookieFile,
		"createdAt":  time.Now().Format(time.RFC3339),
		"isValid":    "true",
	}).Err(); err != nil {
		return fmt.Errorf("redis HSet failed: %w", err)
	}
	ok, err := rdb.Expire(rctx, key, 10*time.Minute).Result()
	if err != nil {
		return fmt.Errorf("redis Expire failed: %w", err)
	}
	if !ok {
		return fmt.Errorf("redis Expire returned false")
	}
	return nil
}

func GetCookieBySession(sid string) string {
	rdb := redis_pool.GetRdb()
	rctx := redis_pool.GetRctx()
	key := "session:" + sid
	cookieFile, _ := rdb.HGet(rctx, key, "cookieFile").Result()
	return cookieFile
}

// 存储二维码的 UniKey
func SetNewQrcodeUniKey(sid string, uniKey string) error {
	rdb := redis_pool.GetRdb()
	rctx := redis_pool.GetRctx()
	if sid == "" || uniKey == "" {
		return fmt.Errorf("sid or uniKey is empty")
	}
	if rdb == nil {
		return fmt.Errorf("redis client is nil")
	}

	key := "qrcode:" + sid
	err := rdb.Set(rctx, key, uniKey, 2*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("set qrcode key failed: %w", err)
	}
	return nil
}

// 获取二维码 UniKey
func GetQrcodeUniKeyBySession(sid string) (string, error) {
	rdb := redis_pool.GetRdb()
	rctx := redis_pool.GetRctx()
	if sid == "" {
		return "", fmt.Errorf("sid is empty")
	}
	if rdb == nil {
		return "", fmt.Errorf("redis client is nil")
	}

	key := "qrcode:" + sid
	uniKey, err := rdb.Get(rctx, key).Result()
	if err != nil {
		return "", err
	}
	return uniKey, nil
}

// 删除二维码 UniKey
func DelQrcodeUniKey(sid string) error {
	rdb := redis_pool.GetRdb()
	rctx := redis_pool.GetRctx()
	if sid == "" {
		return fmt.Errorf("sid is empty")
	}
	if rdb == nil {
		return fmt.Errorf("redis client is nil")
	}

	key := "qrcode:" + sid
	_, err := rdb.Del(rctx, key).Result()
	if err != nil {
		return fmt.Errorf("del qrcode key failed: %w", err)
	}
	return nil
}
