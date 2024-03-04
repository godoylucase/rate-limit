package notification

import (
	"context"
	"fmt"
	"rate-limit/errs"
	"rate-limit/internal/configs"
)

type Gateway interface {
	Send(ctx context.Context, userID string, message string) error
}

type RateLimiter interface {
	CheckLimit(ctx context.Context, conf *LimitConfig) error
}

type service struct {
	gateway  Gateway
	rlimiter RateLimiter
	lconfigs configs.LimitConfigs
}

func NewService(rlimiter RateLimiter, gateway Gateway, lconfigs configs.LimitConfigs) *service {
	return &service{
		gateway:  gateway,
		rlimiter: rlimiter,
		lconfigs: lconfigs,
	}
}

func (s *service) Send(ctx context.Context, notif *Notification) error {
	if !isValid(notif) {
		return fmt.Errorf("invalid notification values: %w", errs.ErrInvalidArguments)
	}

	conf := s.lconfigs.Get(notif.Type)
	key := fmt.Sprintf("%v-%v", notif.UserID.String(), notif.Type)

	if err := s.rlimiter.CheckLimit(ctx, &LimitConfig{
		Key:        key,
		Limit:      conf.Limit,
		WindowSize: conf.UnitToDuration(),
	}); err != nil {
		return fmt.Errorf("error checking rate limit for notification type %v: %w", notif.Type, err)
	}

	if err := s.gateway.Send(ctx, notif.UserID.String(), notif.Message); err != nil {
		return fmt.Errorf("gateway error when sending notification: %w", err)
	}

	return nil
}
