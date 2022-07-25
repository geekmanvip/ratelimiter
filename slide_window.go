package ratelimiter

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"math"
	"sync"
	"time"
)

// 滑动窗口限流器
type slideWindowLimiter struct {
	sync.Mutex
	// 时间间隔，单位毫秒，默认 1000，也就是一秒
	TimeInterval int64
	// 流量限制
	Limit int64
	// 将单位时间窗口切割成多少个格子进行滑动，默认是 10
	SplitNum int64
	// 窗口开始时间
	startAt int64
	// 每个小格子的时间间隔
	eachTime int64
	// 每个小格子累积的请求数
	eachCounters [][2]int64
	// redis 存储的 key(hash 表)
	redisKey string
}

func (sw *slideWindowLimiter) Allow() bool {
	return sw.AllowN(1)
}

func (sw *slideWindowLimiter) AllowN(num int64) bool {
	if rdb != nil {
		return sw.allowWithRedis(num)
	}

	sw.Lock()
	defer sw.Unlock()

	currentTime := time.Now().UnixMilli()
	// 计算当前属于哪个小格子
	nowIndex := int64(math.Floor(float64(currentTime-sw.startAt)/float64(sw.eachTime))) % sw.SplitNum
	// 如果这个格子已经过了一个完整时间窗口，统计数据无效，直接清零
	if currentTime-sw.eachCounters[nowIndex][0] >= sw.TimeInterval {
		sw.eachCounters[nowIndex][1] = 0
	}
	sw.eachCounters[nowIndex][0] = currentTime

	var sum int64 = 0
	lastTime := currentTime - sw.TimeInterval
	for _, item := range sw.eachCounters {
		// 已经过期的格子不计入总数，因为有些格子，可能因为访问频率过低，一直没有被触发，所以还是要判断
		if item[0] >= lastTime {
			sum += item[1]
		}
	}

	if sum+num <= sw.Limit {
		sw.eachCounters[nowIndex][1] += num
		return true
	}

	return false
}

func (sw *slideWindowLimiter) allowWithRedis(num int64) bool {
	// lua 操作序列化等太复杂了，不想写，直接使用 zset 实现滑动日志限流算法，就是耗费的内存有点大
	lua := redis.NewScript(`
		-- 外部参数接收
		local key = KEYS[1]
		local splitNum = tonumber(ARGV[1])
		local timeInterval = tonumber(ARGV[2])
		local eachTime = tonumber(ARGV[3])
		local limit = tonumber(ARGV[4])
		local num = tonumber(ARGV[5])
		
		-- 获取 Redis 的毫秒
		local tmp_time = redis.call("TIME")
		local currentTime = tonumber(tmp_time[1] * 1000 + tmp_time[2]/1000)

		-- 获取 hash 的数据
		local rateInfo = redis.call("HMGET", key, "startAt", "counters", "unixTimes")
		local startAt = rateInfo[1]
		local counters = rateInfo[2]
		local unixTimes = rateInfo[3]

		-- 转成数组
		local countTable = {}	
		string.gsub(counters,'[^,]+',function (w)
			table.insert(countTable, w)
		end)
		local unixTimesTable = {}
		string.gsub(unixTimes,'[^,]+',function (w)
			table.insert(unixTimesTable, w)
		end)
		
		-- 计算当前归属的格子
		local nowIndex = ((currentTime - startAt)/eachTime)%splitNum
		nowIndex = 1
		
		-- 如果这个格子已经过了一个完整时间窗口，统计数据无效，直接清零
		if (currentTime - countTable[nowIndex]) >= timeInterval then
			unixTimesTable[nowIndex] = 0
		end

		-- 计算总数
		local sum = 0
		local lastTime = currentTime - timeInterval
		local newUnixTimesStr = unixTimesTable[1]
		for i, item in pairs(countTable) do
			if unixTimesTable[i] >= lastTime then
				sum =  sum + item
			end

			if i > 0 then
				newUnixTimesStr = newUnixTimesStr..","..unixTimesTable[i]
			end
		end
		redis.call("HSET", key, "unixTimes", newUnixTimesStr)

		if sum + num <= limit then
			countTable[nowIndex] = countTable[nowIndex] + num
			local newCountStr = countTable[1]
			for i, item in pairs(countTable) do
				if i > 0 then
					newCountStr = newCountStr..","..item
				end
			end
			redis.call("HSET", key, "counters", newCountStr)
			return 1
		end
		return 0
	`)
	res, err := lua.Run(ctx,
		rdb,
		[]string{sw.redisKey},
		sw.SplitNum,
		sw.TimeInterval,
		sw.eachTime,
		sw.Limit,
		num,
	).Result()
	fmt.Println(res, err, sw.redisKey)
	if res == 1 {
		return true
	}
	return false
}

// NewSlideWindowLimiter 创建滑动窗口限流器
// @param timeInterval 窗口间隔，单位毫秒，默认 1000ms，也就是 1 秒，正常不要低于这个值
// @param limit 在这个窗口内，不能超过的请求数
// @param splitNum 将这个窗口拆分成多少个子窗口，避免临界窗口问题
func NewSlideWindowLimiter(timeInterval int64, limit int64, splitNum int64) Limiter {
	// todo redis 的 key 需要使用方传递进来，否则的话，分布式就没有一样，只能限制单机，windowLimiter 也有同样的问题
	redisKey := "rl:sw:" + uuid.NewString()
	redisKey = "test" // 测试用
	if rdb != nil {
		// todo 后续可以改为使用 lua 设置为 Redis 服务器时间
		rdb.HSetNX(ctx, redisKey, "startAt", time.Now().UnixMilli())
		rdb.HSetNX(ctx, redisKey, "counters", joinStr("0", splitNum))
		rdb.HSetNX(ctx, redisKey, "unixTimes", joinStr("0", splitNum))
		// hash 的过期问题，需要看下是不是需要自动续期
		rdb.Expire(ctx, redisKey, time.Hour)
	}
	return &slideWindowLimiter{
		TimeInterval: timeInterval,
		Limit:        limit,
		SplitNum:     splitNum,
		startAt:      time.Now().UnixMilli(),
		eachTime:     timeInterval / splitNum,
		eachCounters: make([][2]int64, splitNum),
		redisKey:     redisKey,
	}
}

// 拼接生成 N 个字符
func joinStr(v string, num int64) string {
	result := v
	var i int64
	for i = 0; i < (num - 1); i++ {
		result += "," + v
	}
	return result
}
