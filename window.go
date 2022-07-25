package ratelimiter

import (
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
}

func (w *windowLimiter) Allow() bool {
	return w.AllowN(1)
}

func (w *windowLimiter) AllowN(num int64) bool {
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

func NewWindowLimiter(timeInterval int64, limit int64) Limiter {
	return &windowLimiter{
		TimeInterval: timeInterval,
		Limit:        limit,
		startAt:      time.Now().UnixMilli(),
		counter:      0,
	}
}
