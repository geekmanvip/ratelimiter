package ratelimiter_test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/geekmanvip/ratelimiter"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	tearDown()
	os.Exit(code)
}

// 初始化
func setup() {
	fmt.Println("初始化 redis")
	err := ratelimiter.SetRedisStorage(ratelimiter.RedisConfig{
		Host:     "127.0.0.1",
		Port:     6379,
		Password: "",
		Db:       0,
	})
	if err != nil {
		fmt.Println(err)
	}
}

// 运行结束后回收
func tearDown() {
	fmt.Println("运行结束后回收")
}

// 固定时间窗口测试
func TestWindowLimiter(t *testing.T) {
	limiter := ratelimiter.NewWindowLimiter(1000, 5)
	test(limiter)
}
func TestWindowLimiter_WithRedis(t *testing.T) {
	limiter := ratelimiter.NewWindowLimiter(1000, 5).WithRedis("test")
	test(limiter)
}

// 滑动时间窗口测试
func TestSlideWindowLimiter(t *testing.T) {
	limiter := ratelimiter.NewSlideWindowLimiter(1000, 1)
	test(limiter)
}
func TestSlideWindowLimiter_WithRedis(t *testing.T) {
	limiter := ratelimiter.NewSlideWindowLimiter(1000, 1).WithRedis("test")
	test(limiter)
}

// 漏桶测试
func TestLeakBucketLimiter(t *testing.T) {
	limiter := ratelimiter.NewLeakBucketLimiter(4, 2)
	test(limiter)
}
func TestLeakBucketLimiter_WithRedis(t *testing.T) {
	limiter := ratelimiter.NewLeakBucketLimiter(4, 2).WithRedis("test")
	test(limiter)
}

// 令牌桶测试
func TestTokenBucketLimiter(t *testing.T) {
	limiter := ratelimiter.NewTokenBucketLimiter(4, 2)
	test(limiter)
}
func TestTokenBucketLimiter_WithRedis(t *testing.T) {
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
