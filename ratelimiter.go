package ratelimiter

// redis key 的统一前缀
const (
	redisPrefix       = "rl:"
	windowPrefix      = "wd:"
	slideWindowPrefix = "sw:"
	leakBucketPrefix  = "lb:"
	tokenBucketPrefix = "tb:"
)

// Limiter 定义限流器接口
type Limiter interface {
	Allow() bool
	AllowN(num int64) bool
	WithRedis(key string) Limiter
}
