package rate_limiter

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type SWCData struct {
	Key   string
	Limit int64
	WSize time.Duration
}

type slidingWindowCounter struct {
	redis *redis.Client
}

func NewSlidingWindowCounter() *slidingWindowCounter {
	return &slidingWindowCounter{
		redis: redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "", DB: 0}),
	}
}

func (swc *slidingWindowCounter) IncrementAndCheckLimit(ctx context.Context, data SWCData) bool {
	currentTime := time.Now().Unix()

	// Lua script to atomically remove old timestamps, add the current timestamp, and get the count
	luaScript := `
        local currentTime = tonumber(ARGV[1])
        local windowSize = tonumber(ARGV[2])

        -- Remove old timestamps
        redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', '(' .. (currentTime - windowSize))

        -- Add current timestamp
        redis.call('ZADD', KEYS[1], currentTime, currentTime)

        -- Get the count
        return redis.call('ZCARD', KEYS[1])
    `

	script := redis.NewScript(luaScript)

	// Execute the Lua script atomically
	result, err := script.Run(ctx, swc.redis, []string{data.Key}, currentTime, data.WSize.Seconds()).Result()
	if err != nil {
		log.Printf("Error executing Lua script: %v", err)
		return false
	}

	// Check against the limit
	count := result.(int64)
	return count <= data.Limit
}
