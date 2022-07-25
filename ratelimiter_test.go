package ratelimiter_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/geekmanvip/ratelimiter"
)

func TestWindowLimiter(t *testing.T) {
	ratelimiter.SetRedisStorage(ratelimiter.RedisConfig{
		Host:     "192.168.0.190",
		Port:     8003,
		Password: "",
		Db:       0,
	})

	limiter := ratelimiter.NewWindowLimiter(1000, 2)
	test(limiter)
}

func TestSlideWindowLimiter(t *testing.T) {
	//limiter := ratelimiter.NewSlideWindowLimiter(1000, 2, 10)
	//test(limiter)
}

func TestLeakBucketLimiter(t *testing.T) {
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
