package notification

import (
	"context"
	"testing"
	"time"

	"github.com/godoylucase/rate-limit/internal/configs"
	"github.com/godoylucase/rate-limit/internal/errs"
	"github.com/godoylucase/rate-limit/internal/models"

	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

type CheckLimitFn func(ctx context.Context, key string, limit int64, tWindow time.Duration) (*models.RateLimitStatus, error)
type SendFn func(ctx context.Context, userID string, message string) error

type RateLimitMock struct {
	CheckLimitFn CheckLimitFn
}

type GatewayMock struct {
	SendFn SendFn
}

func (r *RateLimitMock) CheckLimit(ctx context.Context, key string, limit int64, tWindow time.Duration) (*models.RateLimitStatus, error) {
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
	now := time.Now()

	tests := []struct {
		name         string
		notif        *models.Notification
		config       *configs.LimitConfigMap
		checkLimitFn CheckLimitFn
		sendFn       SendFn
		expectedErr  error
	}{
		{
			name: "valid notification",
			notif: &models.Notification{
				Message: "Test message",
				UserID:  ksuid.New(),
				Type:    "Test type",
			},
			config: &config,
			checkLimitFn: func(ctx context.Context, key string, limit int64, tWindow time.Duration) (*models.RateLimitStatus, error) {
				return &models.RateLimitStatus{
					State:     models.Allowed,
					Count:     1,
					ExpiresAt: now.Unix(),
				}, nil
			},
			sendFn:      func(ctx context.Context, userID string, message string) error { return nil },
			expectedErr: nil,
		},
		{
			name: "invalid notification type",
			notif: &models.Notification{
				Message: "Test message",
				UserID:  ksuid.New(),
				Type:    "Invalid Test type",
			},
			config:      &config,
			expectedErr: errs.ErrInvalidArguments,
		},
		{
			name: "invalid notification",
			notif: &models.Notification{
				Message: "",
				UserID:  ksuid.New(),
				Type:    "Test type",
			},
			config:      &config,
			expectedErr: errs.ErrInvalidArguments,
		},
		{
			name: "rate limit check error",
			notif: &models.Notification{
				Message: "Test message",
				UserID:  ksuid.New(),
				Type:    "Test type",
			},
			config: &config,
			checkLimitFn: func(ctx context.Context, key string, limit int64, tWindow time.Duration) (*models.RateLimitStatus, error) {
				return &models.RateLimitStatus{
					State:     models.Denied,
					Count:     1,
					ExpiresAt: now.Unix(),
				}, nil
			},
			expectedErr: &errs.ErrExceededRateLimit{
				State:     string(models.Denied),
				Count:     1,
				ExpiresAt: now.Unix(),
			},
		},
		{
			name: "gateway send error",
			notif: &models.Notification{
				Message: "Test message",
				UserID:  ksuid.New(),
				Type:    "Test type",
			},
			config: &config,
			checkLimitFn: func(ctx context.Context, key string, limit int64, tWindow time.Duration) (*models.RateLimitStatus, error) {
				return &models.RateLimitStatus{
					State:     models.Allowed,
					Count:     1,
					ExpiresAt: now.Unix(),
				}, nil
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

			if err := s.Send(ctx, tt.notif); err != nil {
				require.ErrorContains(t, err, tt.expectedErr.Error())
			}
		})
	}
}
