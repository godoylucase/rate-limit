package rate_limiter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/godoylucase/rate-limit/models"

	"github.com/go-redis/redis/v8"
)

const (
	keyThatDoesNotExist = -2
	keyWithoutExpire    = -1
)

type fixedWindowCounter struct {
	redis *redis.Client
}

func newFixedWindowCounter(redis *redis.Client) *fixedWindowCounter {
	return &fixedWindowCounter{
		redis: redis,
	}
}

// CheckLimit checks the rate limit for a given key within a fixed window.
// It increments the counter for the current window, retrieves the current counter value,
// sets the expiration for the window key, and checks against the limit.
// If the total count is less than or equal to the limit, it returns a RateLimitStatus with State Allowed.
// If the total count exceeds the limit, it returns a RateLimitStatus with State Denied.
// The RateLimitStatus also includes the count, which is the current counter value,
// and the expiresAtMs, which is the timestamp when the window expires in milliseconds.
// It returns the RateLimitStatus and any error encountered during the process.
func (fwc *fixedWindowCounter) CheckLimit(ctx context.Context, key string, limit int64, tWindow time.Duration) (*models.RateLimitStatus, error) {
	pipe := fwc.redis.TxPipeline()

	// Get the current timestamp
	now := time.Now()
	expiresAt := now.Add(tWindow).UnixMilli()

	// Check the current counter value
	pipe.Get(ctx, key)

	// Execute the pipeline
	results, err := pipe.Exec(ctx)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("failed to execute pipeline: %v", err)
		}

		// If the key does not exist, we create it and set the expiration
		pipe.Set(ctx, key, 1, tWindow)
		_, err := pipe.Exec(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to execute pipeline: %v", err)
		}

		if limit >= 1 {
			return &models.RateLimitStatus{
				State:       models.Allowed,
				Count:       1,
				ExpiresAtMs: expiresAt,
			}, nil
		}
	}

	// Increment the counter for the current window
	pipe.Incr(ctx, key)

	// Get the current counter value
	pipe.Get(ctx, key)

	// Set expiration for the window key
	//pipe.Expire(ctx, key, tWindow)

	// Execute the pipeline
	results, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute pipeline: %v", err)
	}

	// Retrieve the count from the result
	countResult := results[1].(*redis.StringCmd)
	total, err := countResult.Int64()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve counter value: %v", err)
	}

	// Check against the limit
	if total <= limit {
		return &models.RateLimitStatus{
			State:       models.Allowed,
			Count:       int(total),
			ExpiresAtMs: expiresAt,
		}, nil
	}

	return &models.RateLimitStatus{
		State:       models.Denied,
		Count:       int(total),
		ExpiresAtMs: expiresAt,
	}, nil
}
