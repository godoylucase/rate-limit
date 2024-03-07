package rate_limiter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/godoylucase/rate-limit/models"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/ksuid"
)

type slidingWindowCounter struct {
	redis *redis.Client
}

func newSlidingWindowCounter(redis *redis.Client) *slidingWindowCounter {
	return &slidingWindowCounter{
		redis: redis,
	}
}

// CheckLimit checks the rate limit for a given key within a sliding window.
// It counts the number of non-expired requests in the sorted set and compares it to the specified limit.
// If the number of requests exceeds the limit, it returns a RateLimitStatus with the state set to Denied.
// Otherwise, it returns a RateLimitStatus with the state set to Allowed.
// The RateLimitStatus also includes the count of requests and the expiration timestamp in milliseconds.
// The sliding window duration is specified by tWindow.
// The key is used to identify the rate limit in the sorted set.
// The function uses a Redis pipeline to efficiently execute multiple Redis commands in a single round trip.
// If any error occurs during the execution, it returns an error.
func (swc *slidingWindowCounter) CheckLimit(ctx context.Context, key string, limit int64, tWindow time.Duration) (*models.RateLimitStatus, error) {
	now := time.Now()
	expiresAtMs := now.Add(tWindow)
	// Calculate the minimum timestamp allowed within the sliding window
	minimum := now.Add(-tWindow)

	// Count how many non-expired requests we have in the sorted set before adding the current request
	result, err := swc.redis.ZCount(ctx, key, strconv.FormatInt(minimum.UnixMilli(), 10), "+inf").Uint64()
	if err == nil && int64(result) >= limit {
		return &models.RateLimitStatus{
			State:       models.Denied,
			Count:       int(result),
			ExpiresAtMs: expiresAtMs.UnixMilli(),
		}, nil
	}

	pipe := swc.redis.TxPipeline()
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
		return nil, fmt.Errorf("failed to execute sorted set pipeline for key: %v with error: %w", key, err)
	}

	// Check for errors in removing expired items
	if err := removeByScore.Err(); err != nil {
		return nil, fmt.Errorf("failed to remove items from key: %v with error: %w", key, err)
	}

	// Check for errors in adding the current item
	if err := add.Err(); err != nil {
		return nil, fmt.Errorf("failed to add item to key: %v with error: %w", key, err)
	}

	// Retrieve the total number of non-expired requests
	total, err := count.Result()
	if err != nil {
		return nil, fmt.Errorf("failed to count items for key: %v with error: %w", key, err)
	}

	// Check if the total requests exceed the specified limit
	if total > limit {
		return &models.RateLimitStatus{
			State:       models.Denied,
			Count:       int(total),
			ExpiresAtMs: expiresAtMs.UnixMilli(),
		}, nil
	}

	// No rate limit exceeded
	return &models.RateLimitStatus{
		State:       models.Allowed,
		Count:       int(total),
		ExpiresAtMs: expiresAtMs.UnixMilli(),
	}, nil
}
