// Package notification provides a service for sending notifications with rate limiting.
package notification

import (
	"context"
	"fmt"

	"github.com/godoylucase/rate-limit/configs"
	"github.com/godoylucase/rate-limit/errs"
	"github.com/godoylucase/rate-limit/models"

	"time"
)

// Gateway defines the interface for sending notifications.
type Gateway interface {
	Send(ctx context.Context, userID string, message string) error
}

// RateLimiter is an interface that defines the methods for checking the rate limit.
type RateLimiter interface {
	CheckLimit(ctx context.Context, key string, limit int64, tWindow time.Duration) (*models.RateLimitStatus, error)
}

// Service is a notification service that sends notifications with rate limiting.
type Service struct {
	gateway  Gateway
	rlimiter RateLimiter
	lconfigs configs.LimitConfigMap
}

// NewService creates a new instance of the Service.
func NewService(rlimiter RateLimiter, gateway Gateway, lconfigs configs.LimitConfigMap) *Service {
	return &Service{
		gateway:  gateway,
		rlimiter: rlimiter,
		lconfigs: lconfigs,
	}
}

// Send sends a notification using the specified context and notification data.
// It performs validation, checks the rate limit, and sends the notification using the gateway.
func (s *Service) Send(ctx context.Context, notif *models.Notification) error {
	if !models.IsValid(notif) {
		return fmt.Errorf("invalid notification values: %w", errs.ErrInvalidArguments)
	}

	conf := s.lconfigs.Get(notif.Type)
	if conf == nil {
		return fmt.Errorf("notification type %v not found in config: %w", notif.Type, errs.ErrInvalidArguments)
	}

	key := fmt.Sprintf("%v-%v", notif.UserID.String(), notif.Type)

	status, err := s.rlimiter.CheckLimit(ctx, key, conf.Limit, conf.WindowsSizeDuration())
	if err != nil {
		return fmt.Errorf("error checking rate limit for notification type %v: %w", notif.Type, err)
	} else if status.State == models.Denied {
		return &errs.ErrExceededRateLimit{
			State:     string(status.State),
			Count:     status.Count,
			ExpiresAt: status.ExpiresAt,
		}

	}

	if err := s.gateway.Send(ctx, notif.UserID.String(), notif.Message); err != nil {
		return fmt.Errorf("gateway error when sending notification: %w", err)
	}

	return nil
}
