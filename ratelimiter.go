package ratelimiter

// Limiter 定义限流器接口
type Limiter interface {
	Allow() bool
	AllowN(num int64) bool
}
