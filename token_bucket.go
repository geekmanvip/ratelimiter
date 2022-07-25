package ratelimiter

import (
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
}

func (t *tokenBucketLimiter) Allow() bool {
	return t.AllowN(1)
}

func (t *tokenBucketLimiter) AllowN(num int64) bool {
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

func NewTokenBucketLimiter(capacity int64, rate int64) Limiter {
	return &tokenBucketLimiter{
		Capacity: capacity,
		Rate:     rate,
		lastTime: time.Now().Unix(),
		waterNum: capacity,
	}
}
