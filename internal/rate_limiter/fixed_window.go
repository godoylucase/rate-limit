package rate_limiter

import (
	"context"
	"errors"
	"fmt"
	"rate-limit/internal/errs"
	"time"

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

func (fwc *fixedWindowCounter) CheckLimit(ctx context.Context, key string, limit int64, tWindow time.Duration) error {
	pipe := fwc.redis.TxPipeline()

	// Get the current value and TTL in a single call
	getResult := pipe.Get(ctx, key)
	ttlResult := pipe.TTL(ctx, key)

	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("failed to execute pipeline with get and ttl to key %v with error: %w", key, err)
	}

	// Extract TTL value
	ttlDuration, err := ttlResult.Result()
	if err != nil {
		return fmt.Errorf("failed to get TTL for key %v with error: %w", key, err)
	}

	// Set expiration if necessary
	if ttlDuration == keyWithoutExpire || ttlDuration == keyThatDoesNotExist {
		if err := fwc.redis.PExpire(ctx, key, tWindow).Err(); err != nil {
			return fmt.Errorf("failed to set an expiration to key %v with error: %w", key, err)
		}
	}

	// Retrieve total requests or initialize if the key does not exist
	total, err := getResult.Uint64()
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("failed to get total requests for key %v with error: %w", key, err)
	}

	// Check limit and deny if exceeded
	if int64(total) >= limit {
		return errs.ErrExceededRateLimit
	}

	// Increment total requests
	total, err = fwc.redis.Incr(ctx, key).Uint64()
	if err != nil {
		return fmt.Errorf("failed to increment key %v with error: %w", key, err)
	}

	// Check limit again and deny if exceeded after increment
	if int64(total) > limit {
		return errs.ErrExceededRateLimit
	}

	// Allow request
	return nil
}
