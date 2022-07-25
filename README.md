# ratelimiter



#### 使用方法

- 安装

    ```shell
    go get github.com/geekmanvip/ratelimiter
    ```
- 使用

    ```go
    // step1 定义一个限流器，可以根据自己的需求，任意选择一个即可
    limiter := ratelimiter.NewWindowLimiter(1000, 2) // 固定窗口限流器
    // limiter := ratelimiter.NewSlideWindowLimiter(1000, 2, 10) // 滑动窗口限流器
    // limiter := ratelimiter.NewLeakBucketLimiter(4, 2) // 漏桶限流器
    // limiter := ratelimiter.NewTokenBucketLimiter(3, 2) // 令牌桶限流器
    
    // step2 限流器通过，则执行某些操作，未通过不执行
    if limiter.Allow() {
        // do somthing
    }
    ```
#### 功能列表
- 支持固定窗口、滑动窗口、漏桶、令牌桶 4 种限流器
- 支持分布式限流器，依赖于 Redis

