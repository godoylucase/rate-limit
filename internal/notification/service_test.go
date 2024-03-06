package notification

import (
	"context"
	"rate-limit/internal/configs"
	"rate-limit/internal/errs"
	"testing"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
)

type CheckLimitFn func(ctx context.Context, key string, limit int64, tWindow time.Duration) error
type SendFn func(ctx context.Context, userID string, message string) error

type RateLimitMock struct {
	CheckLimitFn CheckLimitFn
}

type GatewayMock struct {
	SendFn SendFn
}

func (r *RateLimitMock) CheckLimit(ctx context.Context, key string, limit int64, tWindow time.Duration) error {
	return r.CheckLimitFn(ctx, key, limit, tWindow)
}

func (g *GatewayMock) Send(ctx context.Context, userID string, message string) error {
	return g.SendFn(ctx, userID, message)
}

func TestService_Send(t *testing.T) {
	ctx := context.Background()

	conf := map[string]*configs.LimitConfig{
		"Test type": {
			Type:    "Test type",
			Limit:   1,
			WSizeMs: 1000,
		},
	}

	config := configs.LimitConfigMap(conf)

	tests := []struct {
		name         string
		notif        *Notification
		config       *configs.LimitConfigMap
		checkLimitFn CheckLimitFn
		sendFn       SendFn
		expectedErr  error
	}{
		{
			name: "valid notification",
			notif: &Notification{
				Message: "Test message",
				UserID:  ksuid.New(),
				Type:    "Test type",
			},
			config: &config,
			checkLimitFn: func(ctx context.Context, key string, limit int64, tWindow time.Duration) error {
				return nil
			},
			sendFn: func(ctx context.Context, userID string, message string) error {
				return nil
			},
			expectedErr: nil,
		},
		{
			name: "invalid notification type",
			notif: &Notification{
				Message: "Test message",
				UserID:  ksuid.New(),
				Type:    "Invalid Test type",
			},
			config: &config,
			checkLimitFn: func(ctx context.Context, key string, limit int64, tWindow time.Duration) error {
				return nil
			},
			sendFn: func(ctx context.Context, userID string, message string) error {
				return nil
			},
			expectedErr: errs.ErrInvalidArguments,
		},
		{
			name: "invalid notification",
			notif: &Notification{
				Message: "",
				UserID:  ksuid.New(),
				Type:    "Test type",
			},
			config: &config,
			checkLimitFn: func(ctx context.Context, key string, limit int64, tWindow time.Duration) error {
				return nil
			},
			sendFn: func(ctx context.Context, userID string, message string) error {
				return nil
			},
			expectedErr: errs.ErrInvalidArguments,
		},
		{
			name: "rate limit check error",
			notif: &Notification{
				Message: "Test message",
				UserID:  ksuid.New(),
				Type:    "Test type",
			},
			config: &config,
			checkLimitFn: func(ctx context.Context, key string, limit int64, tWindow time.Duration) error {
				return errs.ErrExceededRateLimit
			},
			sendFn: func(ctx context.Context, userID string, message string) error {
				return nil
			},
			expectedErr: errs.ErrExceededRateLimit,
		},
		{
			name: "gateway send error",
			notif: &Notification{
				Message: "Test message",
				UserID:  ksuid.New(),
				Type:    "Test type",
			},
			config: &config,
			checkLimitFn: func(ctx context.Context, key string, limit int64, tWindow time.Duration) error {
				return nil
			},
			sendFn: func(ctx context.Context, userID string, message string) error {
				return errs.ErrInternalError
			},
			expectedErr: errs.ErrInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService(&RateLimitMock{CheckLimitFn: tt.checkLimitFn}, &GatewayMock{SendFn: tt.sendFn}, *tt.config)

			err := s.Send(ctx, tt.notif)
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}
