package ratelimiter

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"sync"
	"time"
)

// 固定窗口限流器
type windowLimiter struct {
	sync.Mutex
	// 时间间隔，单位毫秒，默认 1000，也就是一秒
	TimeInterval int64
	// 流量限制
	Limit int64
	// 窗口开始时间
	startAt int64
	// 窗口内累积的请求数
	counter int64
	// 计数存储在 Redis 中的时候的 key
	redisKey string
}

func (w *windowLimiter) Allow() bool {
	return w.AllowN(1)
}

func (w *windowLimiter) AllowN(num int64) bool {
	// 设置了 Redis 存储，则使用 Redis 实现分布式限流
	if rdb != nil {
		return w.allowWithRedis(num)
	}

	w.Lock()
	defer w.Unlock()

	currentTime := time.Now().UnixMilli()
	// 窗口更新，数据重置
	if currentTime-w.startAt >= w.TimeInterval {
		w.startAt = currentTime
		w.counter = 0
	}

	// 窗口未满，则可以执行
	if w.counter+num <= w.Limit {
		w.counter += num
		return true
	}

	return false
}

// 使用 redis 实现限流
func (w *windowLimiter) allowWithRedis(num int64) bool {
	if w.redisKey == "" {
		w.redisKey = "rl:wd:" + uuid.NewString()
	}

	lua := redis.NewScript(`
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local num = tonumber(ARGV[2])
		local expire_time = ARGV[3]
		-- 首先判断键是否存在
		local is_exists = redis.call("EXISTS", key)
		-- 键存在就加num
		if is_exists == 1 then
			if redis.call("INCRBY", key, num) > limit then
				redis.call("DECRBY", key, num)
				return 0
			else
				return 1
			end
		-- 不存在的话，就设置键，同时设置有效期
		else
			redis.call("SET", key, num)
			redis.call("PEXPIRE", key, expire_time)
    		return 1  
		end
	`)
	res, _ := lua.Run(context.Background(), rdb, []string{w.redisKey}, w.Limit, num, w.TimeInterval).Int()
	if res == 1 {
		return true
	}
	return false
}

func NewWindowLimiter(timeInterval int64, limit int64) Limiter {
	return &windowLimiter{
		TimeInterval: timeInterval,
		Limit:        limit,
		startAt:      time.Now().UnixMilli(),
		counter:      0,
	}
}