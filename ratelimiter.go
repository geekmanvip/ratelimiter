package ratelimiter

import (
	"sync"
	"time"
)

// 定义限流器接口
type Limiter interface {
	Allow() bool
	AllowN(num int64) bool
}

type windowLimiter struct {
	sync.Mutex
	// 时间间隔，单位毫秒，默认 1000，也就是一秒
	timeInterval int64
	// 流量限制
	limit int64
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
	if currentTime-w.startAt >= w.timeInterval {
		w.startAt = currentTime
		w.counter = 0
	}

	// 窗口未满，则可以执行
	if w.counter+num <= w.limit {
		w.counter += num
		return true
	}

	return false
}

func NewWindowLimiter(timeInterval int64, limit int64) Limiter {
	return &windowLimiter{
		timeInterval: timeInterval,
		limit:        limit,
		startAt:      time.Now().UnixMilli(),
		counter:      0,
	}
}
