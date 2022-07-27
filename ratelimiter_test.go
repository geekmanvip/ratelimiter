package ratelimiter_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/geekmanvip/ratelimiter"
)

func TestSetRedisStorage(t *testing.T) {
	err := setRedisStorage()
	if err != nil {
		fmt.Println(err)
	}
}

// 固定时间窗口测试
func TestWindowLimiter(t *testing.T) {
	limiter := ratelimiter.NewWindowLimiter(1000, 5)
	test(limiter)
}
func TestWindowLimiter_WithRedis(t *testing.T) {
	setRedisStorage()
	limiter := ratelimiter.NewWindowLimiter(1000, 5).WithRedis("test")
	test(limiter)
}

// 滑动时间窗口测试
func TestSlideWindowLimiter(t *testing.T) {
	limiter := ratelimiter.NewSlideWindowLimiter(1000, 1)
	test(limiter)
}
func TestSlideWindowLimiter_WithRedis(t *testing.T) {
	setRedisStorage()
	limiter := ratelimiter.NewSlideWindowLimiter(1000, 1).WithRedis("test")
	test(limiter)
}

// 漏桶测试
func TestLeakBucketLimiter(t *testing.T) {
	limiter := ratelimiter.NewLeakBucketLimiter(4, 2)
	test(limiter)
}
func TestLeakBucketLimiter_WithRedis(t *testing.T) {
	setRedisStorage()
	limiter := ratelimiter.NewLeakBucketLimiter(4, 2).WithRedis("test")
	test(limiter)
}

// 令牌桶测试
func TestTokenBucketLimiter(t *testing.T) {
	limiter := ratelimiter.NewTokenBucketLimiter(4, 2)
	test(limiter)
}
func TestTokenBucketLimiter_WithRedis(t *testing.T) {
	setRedisStorage()
	limiter := ratelimiter.NewTokenBucketLimiter(4, 2).WithRedis("test")
	test(limiter)
}

func test(limiter ratelimiter.Limiter) {
	fmt.Printf("\n%T\n", limiter)
	if err := limiter.Err(); err != nil {
		log.Println(err)
		return
	}

	for i := 0; i < 20; i++ {
		acq := limiter.Allow()
		t := time.Now().Unix()
		fmt.Printf("%d 次请求 %d 是否被接受 %t \n", i, t, acq)
		time.Sleep(time.Millisecond * 100)
	}
}

func setRedisStorage() error {
	return ratelimiter.SetRedisStorage(ratelimiter.RedisConfig{
		Host:     "127.0.0.1",
		Port:     6379,
		Password: "",
		Db:       0,
	})
}
