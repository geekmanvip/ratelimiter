package ratelimiter

import (
	"github.com/go-redis/redis/v8"
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
	// 错误信息
	err error
}

func (sw *slideWindowLimiter) Allow() bool {
	return sw.AllowN(1)
}

func (sw *slideWindowLimiter) AllowN(num int64) bool {
	if sw.redisKey != "" {
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
		local currentTime = math.ceil(tonumber(tmp_time[1] * 1000 + tmp_time[2]/1000))

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
		local nowIndex = math.ceil((currentTime - startAt)/eachTime)%splitNum + 1
		
		-- 如果这个格子已经过了一个完整时间窗口，统计数据无效，直接清零
		if (currentTime - unixTimesTable[nowIndex]) >= timeInterval then
			countTable[nowIndex] = 0
		end
		unixTimesTable[nowIndex] = currentTime

		-- 计算总数
		local sum = 0
		local lastTime = currentTime - timeInterval
		for i, item in pairs(countTable) do
			if tonumber(unixTimesTable[i]) >= lastTime then
				sum =  sum + item
			end
		end
		redis.call("HSET", key, "unixTimes", table.concat(unixTimesTable, ","))

		if sum + num <= limit then
			countTable[nowIndex] = countTable[nowIndex] + num
			redis.call("HSET", key, "counters", table.concat(countTable, ","))
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
	).Int()
	sw.err = err
	if res == 1 {
		return true
	}
	return false
}

func (sw *slideWindowLimiter) WithRedis(redisKey string) Limiter {
	if rdb != nil {
		sw.redisKey = redisPrefix + slideWindowPrefix + redisKey

		rdb.HSetNX(ctx, sw.redisKey, "startAt", time.Now().UnixMilli())
		rdb.HSetNX(ctx, sw.redisKey, "counters", joinStr("0", sw.SplitNum))
		rdb.HSetNX(ctx, sw.redisKey, "unixTimes", joinStr("0", sw.SplitNum))
		// todo hash 的过期问题，需要看下是不是需要自动续期
		rdb.Expire(ctx, sw.redisKey, time.Hour)
	} else {
		sw.err = RedisInitErr
	}

	return sw
}

func (sw *slideWindowLimiter) Err() error {
	return sw.err
}

// NewSlideWindowLimiter 创建滑动窗口限流器
func NewSlideWindowLimiter(timeInterval int64, limit int64, args ...int64) Limiter {
	var splitNum int64 = 10
	if len(args) > 0 && args[0] > 1 && args[0] <= 100 {
		splitNum = args[0]
	}

	var err error
	if timeInterval <= 100 {
		err = TimeIntervalErr
	}
	if limit < 1 {
		err = LimitErr
	}

	return &slideWindowLimiter{
		TimeInterval: timeInterval,
		Limit:        limit,
		SplitNum:     splitNum,
		startAt:      time.Now().UnixMilli(),
		eachTime:     timeInterval / splitNum,
		eachCounters: make([][2]int64, splitNum),
		err:          err,
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
