package ratelimiter

import (
	"github.com/go-redis/redis/v8"
	"math"
	"sync"
	"time"
)

// 令牌桶限流器
type tokenBucketLimiter struct {
	sync.Mutex
	// 桶容量
	Capacity int64
	// 每秒消费的速率
	Rate int64
	// 上次更新时间，秒
	lastTime int64
	// 桶内剩余水量
	waterNum int64
	// Redis key
	redisKey string
	// 错误信息
	err error
}

func (t *tokenBucketLimiter) Allow() bool {
	return t.AllowN(1)
}

func (t *tokenBucketLimiter) AllowN(num int64) bool {
	if t.redisKey != "" {
		return t.allowWithRedis(num)
	}

	t.Lock()
	defer t.Unlock()

	currentTime := time.Now().Unix()
	// 剩余令牌为之前的令牌+这段时间内发放的令牌
	leftToken := t.waterNum + (currentTime-t.lastTime)*t.Rate
	// fmt.Println(t.waterNum, leftToken)
	leftToken = int64(math.Min(float64(t.Capacity), float64(leftToken)))

	t.lastTime = currentTime
	if leftToken-num >= 0 {
		t.waterNum = leftToken - num
		return true
	}

	return false
}

func (t *tokenBucketLimiter) allowWithRedis(num int64) bool {
	lua := redis.NewScript(`
		-- 外部参数接收
		local key = KEYS[1]
		local capacity = tonumber(ARGV[1])
		local rate = tonumber(ARGV[2])
		local num = tonumber(ARGV[3])
		
		-- 获取 Redis 的毫秒
		local tmp_time = redis.call("TIME")
		local currentTime = tonumber(tmp_time[1])
		
		-- 获取 Redis 中存储的 lastTime 等信息
		local rateInfo = redis.call("HMGET", key, "lastTime", "waterNum")
		local lastTime = rateInfo[1]
		local waterNum = rateInfo[2]
		
		-- 剩余令牌
		local leftToken = waterNum + (currentTime - lastTime) * rate
		if leftToken > capacity then 
			leftToken = capacity
		end

		redis.call("HSET", key, "lastTime", currentTime)
		if (leftToken - num) > 0 then
			redis.call("HSET", key, "waterNum", leftToken - num)
			return 1
		end
		return 0
	`)
	res, err := lua.Run(ctx,
		rdb,
		[]string{t.redisKey},
		t.Capacity,
		t.Rate,
		num,
	).Int()
	t.err = err
	if res == 1 {
		return true
	}
	return false
}

func (t *tokenBucketLimiter) WithRedis(redisKey string) Limiter {
	if rdb != nil {
		t.redisKey = redisPrefix + tokenBucketPrefix + redisKey

		rdb.HSetNX(ctx, t.redisKey, "lastTime", t.lastTime)
		rdb.HSetNX(ctx, t.redisKey, "waterNum", t.waterNum)
		// todo hash 的过期问题，需要看下是不是需要自动续期
		rdb.Expire(ctx, t.redisKey, time.Hour)
	} else {
		t.err = RedisInitErr
	}

	return t
}

func (t *tokenBucketLimiter) Err() error {
	return t.err
}

func NewTokenBucketLimiter(capacity int64, rate int64) Limiter {
	var err error
	if capacity < 1 {
		err = CapacityErr
	}
	if rate < 1 {
		err = RateErr
	}
	if capacity < rate {
		err = CapacityLessRateErr
	}
	return &tokenBucketLimiter{
		Capacity: capacity,
		Rate:     rate,
		lastTime: time.Now().Unix(),
		waterNum: capacity,
		err:      err,
	}
}
