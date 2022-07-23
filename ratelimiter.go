package ratelimiter

import (
	"math"
	"sync"
	"time"
)

// 定义限流器接口
type Limiter interface {
	Allow() bool
	AllowN(num int64) bool
}

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

// 滑动窗口限流器
type slideWindowLimiter struct {
	sync.Mutex
	// 时间间隔，单位毫秒，默认 1000，也就是一秒
	TimeInterval int64
	// 流量限制
	Limit int64
	// 将单位时间窗口切割成多少个格子进行滑动，默认是 10
	SplitNum int64
	// 窗口开始时间
	startAt int64
	// 每个小格子的时间间隔
	eachTime int64
	// 每个小格子累积的请求数
	eachCouters [][2]int64
}

func (sw *slideWindowLimiter) Allow() bool {
	return sw.AllowN(1)
}

func (sw *slideWindowLimiter) AllowN(num int64) bool {
	sw.Lock()
	defer sw.Unlock()

	currentTime := time.Now().UnixMilli()
	// 计算当前属于哪个小格子
	nowIndex := int64(math.Floor(float64(currentTime-sw.startAt)/float64(sw.eachTime))) % sw.SplitNum
	// 如果这个格子已经过了一个完整时间窗口，统计数据无效，直接清零
	if currentTime-sw.eachCouters[nowIndex][0] >= sw.TimeInterval {
		sw.eachCouters[nowIndex][1] = 0
	}
	sw.eachCouters[nowIndex][0] = currentTime

	var sum int64 = 0
	lastTime := currentTime - sw.TimeInterval
	for _, item := range sw.eachCouters {
		// 已经过期的格子不计入总数，因为有些格子，可能因为访问频率过低，一直没有被触发，所以还是要判断
		if item[0] >= lastTime {
			sum += item[1]
		}
	}

	if sum+num <= sw.Limit {
		sw.eachCouters[nowIndex][1] += num
		return true
	}

	return false
}

func NewSlideWindowLimiter(timeInterval int64, limit int64, splitNum int64) Limiter {
	return &slideWindowLimiter{
		TimeInterval: timeInterval,
		Limit:        limit,
		SplitNum:     splitNum,
		startAt:      time.Now().UnixMilli(),
		eachTime:     timeInterval / splitNum,
		eachCouters:  make([][2]int64, splitNum),
	}
}
