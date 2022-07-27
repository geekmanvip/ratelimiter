package ratelimiter

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

var (
	rdb *redis.Client
	ctx = context.Background()
	// 限流器创建的 hash key 集合
	redisHashKeys   []string
	redisHashExpire = time.Minute * 10
	redisRefreshTTL = redisHashExpire / 2
)

type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	Db       int    `json:"db"`
}

func SetRedisStorage(redisConfig RedisConfig) error {
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password: redisConfig.Password, // no password set
		DB:       redisConfig.Db,       // use default DB
	})

	// ping 超时时间
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	go autoReExpire()

	_, err := rdb.Ping(ctx).Result()
	return err
}

// Redis hash key 设置有效期，并且加入数组中，由后台协程确保在进程退出前不会过期
func setHashKeyExpire(key string) {
	redisHashKeys = append(redisHashKeys, key)
	if rdb != nil {
		rdb.Expire(ctx, key, redisHashExpire)
	}
}

// 自动给 key 续期
func autoReExpire() {
	for range time.Tick(redisRefreshTTL) {
		if rdb == nil {
			continue
		}
		for _, key := range redisHashKeys {
			rdb.Expire(ctx, key, redisHashExpire)
		}
	}
}
