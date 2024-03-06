package rate_limiter

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	FixedWindowCounter   = "sliding_window"
	SlidingWindowCounter = "fixed_window"
)

type RateLimiter interface {
	CheckLimit(ctx context.Context, key string, limit int64, tWindow time.Duration) error
}

func Get(typ string, redis *redis.Client) RateLimiter {
	switch typ {
	case FixedWindowCounter:
		return newFixedWindowCounter(redis)
	case SlidingWindowCounter:
		return newSlidingWindowCounter(redis)
	default:
		return newSlidingWindowCounter(redis)
	}
}
