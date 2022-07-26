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
	// Allow 允许一次请求
	Allow() bool
	// AllowN 允许多次请求
	AllowN(num int64) bool
	// WithRedis 使用 Redis 进行限流
	WithRedis(redisKey string) Limiter
	// 内部 Redis 实现
	allowWithRedis(num int64) bool
}
