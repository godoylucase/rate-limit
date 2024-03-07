// Package rate_limiter provides a rate limiter implementation for controlling the rate of requests.
// It includes two types of rate limiters: FixedWindowCounter and SlidingWindowCounter.
// The Get function returns the appropriate rate limiter based on the provided type.
package rate_limiter

import (
	"context"
	"time"

	"github.com/godoylucase/rate-limit/internal/models"

	"github.com/go-redis/redis/v8"
)

const (
	FixedWindowCounter   = "sliding_window"
	SlidingWindowCounter = "fixed_window"
)

// RateLimiter is an interface that defines the methods for checking the rate limit.
type RateLimiter interface {
	CheckLimit(ctx context.Context, key string, limit int64, tWindow time.Duration) (*models.RateLimitStatus, error)
}

// Get returns the appropriate rate limiter based on the provided type.
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
