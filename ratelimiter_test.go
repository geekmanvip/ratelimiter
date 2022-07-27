package ratelimiter_test

import (
	"fmt"
	"log"
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
	//limiter := ratelimiter.NewWindowLimiter(1000, 5).WithRedis("test")
	//test(limiter)
}

func TestSlideWindowLimiter(t *testing.T) {
	//limiter := ratelimiter.NewSlideWindowLimiter(1000, 1)
	//test(limiter)
}

func TestLeakBucketLimiter(t *testing.T) {
	//ratelimiter.NewLeakBucketLimiter(4, 2).WithRedis("test").Allow()
	limiter := ratelimiter.NewLeakBucketLimiter(4, 2)
	test(limiter)
	//test(ratelimiter.NewLeakBucketLimiter(4, 2))
}

func TestTokenBucketLimiter(t *testing.T) {
	//test(ratelimiter.NewTokenBucketLimiter(3, 2).WithRedis("test"))
}

func test(limiter ratelimiter.Limiter) {
	if err := limiter.Err(); err != nil {
		log.Println(err)
		return
	}
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
