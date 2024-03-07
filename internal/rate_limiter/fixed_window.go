package rate_limiter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/godoylucase/rate-limit/internal/models"

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

// CheckLimit checks the rate limit for a given key. It retrieves the current value and TTL of the key from Redis.
// If the key does not exist or has expired, it sets an expiration time for the key.
// It then checks if the total requests exceed the specified limit and denies the request if it does.
// If the total requests do not exceed the limit, it increments the total requests and allows the request.
// If any error occurs during the process, it returns an error.
func (fwc *fixedWindowCounter) CheckLimit(ctx context.Context, key string, limit int64, tWindow time.Duration) (*models.RateLimitStatus, error) {
	pipe := fwc.redis.TxPipeline()

	// Get the current value and TTL in a single call
	getResult := pipe.Get(ctx, key)
	ttlResult := pipe.TTL(ctx, key)

	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("failed to execute pipeline with get and ttl to key %v with error: %w", key, err)
	}

	// Extract TTL value
	ttlDuration, err := ttlResult.Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get TTL for key %v with error: %w", key, err)
	}

	// Set expiration if necessary
	if ttlDuration == keyWithoutExpire || ttlDuration == keyThatDoesNotExist {
		if err := fwc.redis.PExpire(ctx, key, tWindow).Err(); err != nil {
			return nil, fmt.Errorf("failed to set an expiration to key %v with error: %w", key, err)
		}
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(ttlDuration)

	// Retrieve total requests or initialize if the key does not exist
	total, err := getResult.Uint64()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("failed to get total requests for key %v with error: %w", key, err)
	}

	// Check limit and deny if exceeded
	if int64(total) >= limit {
		return &models.RateLimitStatus{
			State:     models.Denied,
			Count:     int(total),
			ExpiresAt: expiresAt.Unix(),
		}, nil
	}

	// Increment total requests
	total, err = fwc.redis.Incr(ctx, key).Uint64()
	if err != nil {
		return nil, fmt.Errorf("failed to increment key %v with error: %w", key, err)
	}

	// Check limit again and deny if exceeded after increment
	if int64(total) > limit {
		return &models.RateLimitStatus{
			State:     models.Denied,
			Count:     int(total),
			ExpiresAt: expiresAt.Unix(),
		}, nil
	}

	// Allow request
	return &models.RateLimitStatus{
		State:     models.Allowed,
		Count:     int(total),
		ExpiresAt: expiresAt.Unix(),
	}, nil
}
