package rate_limiter

import (
	"context"
	"fmt"
	"rate-limit/errs"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/ksuid"
)

type slidingWindowCounter struct {
	redis *redis.Client
}

func NewSlidingWindowCounter(redis *redis.Client) *slidingWindowCounter {
	return &slidingWindowCounter{
		redis: redis,
	}
}

func (swc *slidingWindowCounter) CheckLimit(ctx context.Context, key string, limit int64, tWindow time.Duration) error {
	now := time.Now()

	pipe := swc.redis.TxPipeline()

	// Calculate the minimum timestamp allowed within the sliding window
	minimum := now.Add(-tWindow)

	// Remove all requests that have already expired within the sliding window
	removeByScore := pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(minimum.UnixMilli(), 10))

	// Add the current request to the sorted set
	add := pipe.ZAdd(ctx, key, &redis.Z{
		Score:  float64(now.UnixMilli()),
		Member: ksuid.New(),
	})

	// Count how many non-expired requests we have in the sorted set
	count := pipe.ZCount(ctx, key, "-inf", "+inf")

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute sorted set pipeline for key: %v with error: %w", key, err)
	}

	// Check for errors in removing expired items
	if err := removeByScore.Err(); err != nil {
		return fmt.Errorf("failed to remove items from key: %v with error: %w", key, err)
	}

	// Check for errors in adding the current item
	if err := add.Err(); err != nil {
		return fmt.Errorf("failed to add item to key: %v with error: %w", key, err)
	}

	// Retrieve the total number of non-expired requests
	totalRequests, err := count.Result()
	if err != nil {
		return fmt.Errorf("failed to count items for key: %v with error: %w", key, err)
	}

	// Check if the total requests exceed the specified limit
	if totalRequests > limit {
		return errs.ErrExceededRateLimit
	}

	// No rate limit exceeded
	return nil
}
