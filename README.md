# ratelimiter



#### 使用方法

- 安装

    ```shell
    go get github.com/geekmanvip/ratelimiter
    ```
- 使用

    ```go
    // 导入包
    import (
    	"github.com/geekmanvip/ratelimiter"
    )
    
    // 可选，配置 Redis 存储，如果需要使用分布式限流，那么需要提前设置 Redis 存储
    ratelimiter.SetRedisStorage(ratelimiter.RedisConfig{
        Host:     "127.0.0.1",
        Port:     6379,
        Password: "",
        Db:       0,
    })
    
    // step1 定义一个限流器，可以根据自己的需求，任意选择一个即可
    limiter := ratelimiter.NewWindowLimiter(1000, 2) // 固定窗口限流器
    // limiter := ratelimiter.NewSlideWindowLimiter(1000, 2, 10) // 滑动窗口限流器
    // limiter := ratelimiter.NewLeakBucketLimiter(4, 2) // 漏桶限流器
    // limiter := ratelimiter.NewTokenBucketLimiter(3, 2) // 令牌桶限流器
    
    // 可选，使用 Redis 存储，作为分布式限流，需要定义一个 Redis key，包会对这个 key 加上统一前缀
    limiter := ratelimiter.NewWindowLimiter(1000, 2).WithRedis("test")
    // 也可以这么使用
    // limiter := ratelimiter.NewWindowLimiter(1000, 2)
    // limiter.WithRedis("test")
    
    // 可选，限流器配置错误检测，如果有错误，则后续的 Allow 全部会返回 false 
    if err := limiter.Err(); err != nil {
        log.Println(err)
        return
    }
    
    // step2 限流器通过，则执行某些操作，未通过不执行
    if limiter.Allow() {
        // do somthing
    }
    
    // 可选 同时获取 N 个令牌
    if limiter.AllowN(5) {
        // do somthing
    }
    ```
#### 功能列表
- 支持固定窗口、滑动窗口、漏桶、令牌桶 4 种限流器
- 支持分布式限流器，依赖于 Redis，直接在限流上使用 WithRedis("test") 即可

#### 使用建议

-   使用限流器来保护自己，避免被其他服务请求打挂
    -   建议使用固定窗口限流器(NewWindowLimiter)或者滑动窗口限流器(NewSlideWindowLimiter)
    -   一般作为 middleware 在路由层级配置即可，可以针对具体接口或者某个 group、整个服务，进行配置，根据自己的业务需求来
-   使用限流器来保护下游，避免把其他服务打挂
    -   建议使用漏桶(NewLeakBucketLimiter)或者令牌桶(NewTokenBucketLimiter)限流器
    -   在调用其他服务的 SDK 那边进行配置，可以根据服务名，或者接口名来进行配置
