package ratelimiter_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/geekmanvip/ratelimiter"
)

func TestWindowLimiter(t *testing.T) {
	limiter := ratelimiter.NewWindowLimiter(1000, 2)
	test(limiter)
}

func TestSlideWindowLimiter(t *testing.T) {
	limiter := ratelimiter.NewSlideWindowLimiter(1000, 2, 10)
	test(limiter)
}

func test(limiter ratelimiter.Limiter) {
	for i := 0; i < 10; i++ {
		acq := limiter.Allow()
		t := time.Now().Unix()
		fmt.Printf("%d 次请求 %d 是否被接受 %t \n", i, t, acq)
		time.Sleep(time.Millisecond * 300)
	}
}
