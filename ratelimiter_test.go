package ratelimiter_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/geekmanvip/ratelimiter"
)

func TestSetRedisStorage(t *testing.T) {
	ratelimiter.SetRedisStorage(ratelimiter.RedisConfig{
		Host:     "127.0.0.1",
		Port:     6379,
		Password: "",
		Db:       0,
	})
}

func TestWindowLimiter(t *testing.T) {
	//test(ratelimiter.NewWindowLimiter(1000, 2))
}

func TestSlideWindowLimiter(t *testing.T) {
	//ratelimiter.NewSlideWindowLimiter(1000, 2, 10).Allow()
	//test(ratelimiter.NewSlideWindowLimiter(1000, 2, 10))
}

func TestLeakBucketLimiter(t *testing.T) {
	ratelimiter.NewLeakBucketLimiter(4, 2).WithRedis("test").Allow()
	//test(ratelimiter.NewLeakBucketLimiter(4, 2))
}

func TestTokenBucketLimiter(t *testing.T) {
	//test(ratelimiter.NewTokenBucketLimiter(3, 2))
}

func test(limiter ratelimiter.Limiter) {
	fmt.Printf("\n%T\n", limiter)
	for i := 0; i < 20; i++ {
		acq := limiter.Allow()
		t := time.Now().Unix()
		fmt.Printf("%d 次请求 %d 是否被接受 %t \n", i, t, acq)
		time.Sleep(time.Millisecond * 100)
	}
}
