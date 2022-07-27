package ratelimiter

import (
	"github.com/go-redis/redis/v8"
	"sync"
	"time"
)

// 漏桶限流器
type leakBucketLimiter struct {
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

func (l *leakBucketLimiter) Allow() bool {
	return l.AllowN(1)
}

func (l *leakBucketLimiter) AllowN(num int64) bool {
	if l.redisKey != "" {
		return l.allowWithRedis(num)
	}

	l.Lock()
	defer l.Unlock()

	currentTime := time.Now().Unix()
	// 已流出的水，过了多少时间，每秒流出的速率，计算出已经流出的水
	outWater := (currentTime - l.lastTime) * l.Rate
	// 剩余的流量，使用有效流入的流量减去已经流出的流量，看剩余流量是否超过了桶的流量，如果未超出，容量就+1，否则就丢弃
	currentWater := l.waterNum - outWater
	if currentWater < 0 { // 剩余的数量不能是负数，最多就是变成 0
		currentWater = 0
	}
	l.lastTime = currentTime

	if currentWater+num <= l.Capacity {
		l.waterNum = currentWater + num
		return true
	}

	return false
}

func (l *leakBucketLimiter) allowWithRedis(num int64) bool {
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
		
		-- 已流出的水，过了多少时间，每秒流出的速率，计算出已经流水的谁
		local outWater = (currentTime - lastTime) * rate
		
		-- 剩余的流量，使用有效流入的流量减去已经流出的流量，看剩余流量是否超过了桶的流量，如果未超出，容量就+1，否则就丢弃
		local currentWater = waterNum - outWater
		if currentWater < 0 then 
			currentWater = 0
		end

		redis.call("HSET", key, "lastTime", currentTime)
		if (currentWater + num) <= capacity then
			redis.call("HSET", key, "waterNum", currentWater + num)
			return 1
		end
		return 0
	`)
	res, err := lua.Run(ctx,
		rdb,
		[]string{l.redisKey},
		l.Capacity,
		l.Rate,
		num,
	).Int()
	l.err = err
	if res == 1 {
		return true
	}
	return false
}

func (l *leakBucketLimiter) WithRedis(redisKey string) Limiter {
	if rdb != nil {
		l.redisKey = redisPrefix + leakBucketPrefix + redisKey

		rdb.HSetNX(ctx, l.redisKey, "lastTime", l.lastTime)
		rdb.HSetNX(ctx, l.redisKey, "waterNum", l.waterNum)
		// todo hash 的过期问题，需要看下是不是需要自动续期
		rdb.Expire(ctx, l.redisKey, time.Hour)
	} else {
		l.err = RedisInitErr
	}
	return l
}

func (l *leakBucketLimiter) Err() error {
	return l.err
}

// NewLeakBucketLimiter 容量和 rate 的单位都是秒，这边暂不支持其他的时间单位
func NewLeakBucketLimiter(capacity int64, rate int64) Limiter {
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
	return &leakBucketLimiter{
		Capacity: capacity,
		Rate:     rate,
		lastTime: time.Now().Unix(),
		waterNum: 0,
		err:      err,
	}
}
