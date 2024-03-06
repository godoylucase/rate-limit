package notification

import (
	"context"
	"fmt"
	"rate-limit/errs"
	"rate-limit/internal/configs"
	"time"
)

type Gateway interface {
	Send(ctx context.Context, userID string, message string) error
}

type RateLimiter interface {
	CheckLimit(ctx context.Context, key string, limit int64, tWindow time.Duration) error
}

type Service struct {
	gateway  Gateway
	rlimiter RateLimiter
	lconfigs configs.LimitConfigMap
}

func NewService(rlimiter RateLimiter, gateway Gateway, lconfigs configs.LimitConfigMap) *Service {
	return &Service{
		gateway:  gateway,
		rlimiter: rlimiter,
		lconfigs: lconfigs,
	}
}

func (s *Service) Send(ctx context.Context, notif *Notification) error {
	if !isValid(notif) {
		return fmt.Errorf("invalid notification values: %w", errs.ErrInvalidArguments)
	}

	conf := s.lconfigs.Get(notif.Type)
	key := fmt.Sprintf("%v-%v", notif.UserID.String(), notif.Type)

	if err := s.rlimiter.CheckLimit(ctx, key, conf.Limit, conf.WindowsSizeDuration()); err != nil {
		return fmt.Errorf("error checking rate limit for notification type %v: %w", notif.Type, err)
	}

	if err := s.gateway.Send(ctx, notif.UserID.String(), notif.Message); err != nil {
		return fmt.Errorf("gateway error when sending notification: %w", err)
	}

	return nil
}
