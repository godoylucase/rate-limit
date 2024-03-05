package rate_limiter

import (
	"context"
	"fmt"
	"rate-limit/errs"
	"rate-limit/internal/notification"
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

func (swc *slidingWindowCounter) CheckLimit(ctx context.Context, req *notification.CheckLimitRequest) error {
	now := time.Now()

	minimum := now.Add(-req.TWindow)

	p := swc.redis.Pipeline()

	// remove all requests that have already expired on this set
	removeByScore := p.ZRemRangeByScore(ctx, req.Key, "0", strconv.FormatInt(minimum.UnixMilli(), 10))

	// add the current request
	add := p.ZAdd(ctx, req.Key, &redis.Z{
		Score:  float64(now.UnixMilli()),
		Member: ksuid.New(),
	})

	// count how many non-expired requests we have on the sorted set
	count := p.ZCount(ctx, req.Key, "-inf", "+inf")

	if _, err := p.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute sorted set pipeline for key: %v with error: %w", req.Key, err)
	}

	if err := removeByScore.Err(); err != nil {
		return fmt.Errorf("failed to remove items from key: %v with error: %w", req.Key, err)
	}

	if err := add.Err(); err != nil {
		return fmt.Errorf("failed to add item to key: %v with error: %w", req.Key, err)
	}

	totalRequests, err := count.Result()
	if err != nil {
		return fmt.Errorf("failed to count items for key: %v with error: %w", req.Key, err)
	}

	if totalRequests > req.Limit {
		return errs.ErrExceededRateLimit
	}

	return nil
}
