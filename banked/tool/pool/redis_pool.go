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

package redis_pool

import (
	"bvtc/config"
	"bvtc/log"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client
var ctx = context.Background()

func InitRedis() {
	redisCfg := config.GetConfig().Redis
	rdb = redis.NewClient(&redis.Options{
		Addr:     redisCfg.Host + ":" + redisCfg.Port,
		Password: redisCfg.Password,
		DB:       redisCfg.DB,
	})
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic("fail to connect to redis,error=" + err.Error())
	} else {
		log.Logger.Info("connect to redis success", log.Any("pong = ", pong))
	}

}

func GetRctx() context.Context {
	return ctx
}

func GetRdb() *redis.Client {
	return rdb
}

// 将 Redis 中的会话有效期延长至 7 天，并刷新标记
func ExtendTimeForCookie(sid string) error {
	if sid == "" {
		return fmt.Errorf("sid is empty")
	}
	if rdb == nil {
		return fmt.Errorf("redis client is nil")
	}
	key := "session:" + sid

	exists, err := rdb.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("redis check exists failed: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("session not exists or expired")
	}

	ok, err := rdb.Expire(ctx, key, 7*24*time.Hour).Result()
	if err != nil {
		return fmt.Errorf("redis expire failed: %w", err)
	}
	if !ok {
		return fmt.Errorf("session expire refresh failed")
	}

	return nil
}
