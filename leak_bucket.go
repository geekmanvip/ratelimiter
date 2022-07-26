package ratelimiter

import (
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
}

func (l *leakBucketLimiter) Allow() bool {
	return l.AllowN(1)
}

func (l *leakBucketLimiter) AllowN(num int64) bool {
	l.Lock()
	defer l.Unlock()

	currentTime := time.Now().Unix()
	// 已流出的水，过了多少时间，每秒流出的速率，计算出已经流水的谁
	outWater := (currentTime - l.lastTime) * l.Rate
	// 剩余的流量，使用有效流入的流量减去已经流出的流量，看剩余流量是否超过了桶的流量，如果未超出，容量就+1，否则就丢弃
	currentWater := l.waterNum - outWater
	l.lastTime = currentTime

	if currentWater+num <= l.Capacity {
		l.waterNum = currentWater + num
		return true
	}

	return false
}

func (l *leakBucketLimiter) WithRedis(key string) Limiter {
	l.redisKey = redisPrefix + leakBucketPrefix + key
	return l
}

// NewLeakBucketLimiter 容量和 rate 的单位都是秒，这边暂不支持其他的时间单位
func NewLeakBucketLimiter(capacity int64, rate int64) Limiter {
	return &leakBucketLimiter{
		Capacity: capacity,
		Rate:     rate,
		lastTime: time.Now().Unix(),
		waterNum: 0,
	}
}
