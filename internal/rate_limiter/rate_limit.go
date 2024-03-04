package rate_limiter

import (
	"context"
	"fmt"
	"rate-limit/errs"
	"rate-limit/internal/notification"
	"time"

	"github.com/go-redis/redis/v8"
)

type slidingWindowCounter struct {
	redis *redis.Client
}

func NewSlidingWindowCounter() *slidingWindowCounter {
	return &slidingWindowCounter{
		redis: redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "", DB: 0}),
	}
}

func (swc *slidingWindowCounter) CheckLimit(ctx context.Context, lconfig *notification.LimitConfig) error {
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
	result, err := script.Run(ctx, swc.redis, []string{lconfig.Key}, currentTime, lconfig.WindowSize.Seconds()).Result()
	if err != nil {
		return fmt.Errorf("error executing redis Lua script: %w", errs.ErrInternal)
	}

	// Check against the limit
	count := result.(int64)
	if count > lconfig.Limit {
		return errs.ErrExceededRateLimit
	}

	return nil
}
