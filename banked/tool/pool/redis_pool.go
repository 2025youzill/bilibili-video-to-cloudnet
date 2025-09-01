package redis_pool

import (
	"bvtc/config"
	"bvtc/log"
	"context"

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
