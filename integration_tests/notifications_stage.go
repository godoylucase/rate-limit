package integration_tests

import (
	"context"
	"fmt"

	"github.com/godoylucase/rate-limit/configs"
	"github.com/godoylucase/rate-limit/errs"
	"github.com/godoylucase/rate-limit/integration_tests/support/gateway"
	"github.com/godoylucase/rate-limit/models"
	"github.com/godoylucase/rate-limit/notification"
	"github.com/godoylucase/rate-limit/rate_limiter"

	"sync/atomic"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sender func(ctx context.Context, userID string, message string) error

type noOpGW struct {
	sender
}

func (g *noOpGW) Send(ctx context.Context, userID string, message string) error {
	return g.sender(ctx, userID, message)
}

type notif struct {
	itself *models.Notification
	isSent bool
}

type NotificationStage struct {
	t       *testing.T
	assert  *assert.Assertions
	require *require.Assertions

	userID ksuid.KSUID

	conf *configs.NotificationService

	gateway  notification.Gateway
	rlimiter notification.RateLimiter
	service  *notification.Service

	notifications []*notif

	sentCount atomic.Int32
}

func NotificationServiceTestStages(t *testing.T) (*NotificationStage, *NotificationStage, *NotificationStage) {
	stage := &NotificationStage{
		t:       t,
		assert:  assert.New(t),
		require: require.New(t),
		userID:  ksuid.New(),
	}

	return stage, stage, stage
}

func (ns *NotificationStage) and() *NotificationStage {
	return ns
}

// given
func (ns *NotificationStage) a_rate_limit_configuration_from(filepath string) *NotificationStage {
	conf, err := configs.Load(filepath)
	if err != nil {
		panic(fmt.Errorf("failed to load limit configuration: %w", err))
	}
	ns.conf = conf

	return ns
}

func (ns *NotificationStage) a_no_op_gateway() *NotificationStage {
	noOp := &gateway.LogGW{}

	senderFn := func(ctx context.Context, userID string, message string) error {
		ns.sentCount.Add(1)
		return noOp.Send(ctx, userID, message)
	}

	ns.gateway = &noOpGW{senderFn}

	return ns
}

func (ns *NotificationStage) a_redis_rate_limiter() *NotificationStage {
	client := redis.NewClient(&redis.Options{Addr: ns.conf.RedisAddr, Password: "", DB: 0})
	ns.rlimiter = rate_limiter.Get(ns.conf.RateLimiterType, client)

	return ns
}

func (ns *NotificationStage) a_notification_service() *NotificationStage {
	ns.service = notification.NewService(ns.rlimiter, ns.gateway, ns.conf.Limits)
	return ns
}

func (ns *NotificationStage) status_notifications_group_with_limit_size() *NotificationStage {
	conf := ns.conf.Limits.Get("status")
	ns.assert.NotNil(conf)

	ns.a_group_of_notifications_of_type_and_size(conf.Type, int(conf.Limit))

	return ns
}

func (ns *NotificationStage) status_notifications_group_with_twice_limit_size() *NotificationStage {
	conf := ns.conf.Limits.Get("status")
	ns.assert.NotNil(conf)

	ns.a_group_of_notifications_of_type_and_size(conf.Type, int(conf.Limit)*2)

	return ns
}

func (ns *NotificationStage) news_notifications_group_with_twice_limit_size() *NotificationStage {
	conf := ns.conf.Limits.Get("news")
	ns.assert.NotNil(conf)

	ns.a_group_of_notifications_of_type_and_size(conf.Type, int(conf.Limit)*2)

	return ns
}

func (ns *NotificationStage) a_group_of_notifications_of_type_and_size(typ string, size int) *NotificationStage {
	for i := 0; i < size; i++ {
		n := &notif{
			itself: &models.Notification{
				Type:    typ,
				UserID:  ns.userID,
				Message: fmt.Sprintf("message from type %v, and value %v", typ, i+1),
			},
			isSent: false,
		}
		ns.notifications = append(ns.notifications, n)
	}

	return ns
}

// when
func (ns *NotificationStage) the_service_sends_notifications_within_the_time_window() *NotificationStage {
	for _, notif := range ns.notifications {
		conf := ns.conf.Limits.Get(notif.itself.Type)
		ns.assert.NotNil(conf)

		if err := ns.service.Send(context.Background(), notif.itself); err == nil {
			notif.isSent = true
		} else {
			fmt.Printf("error sending notification: %v \n", err)

			var errLimit *errs.ErrExceededRateLimit
			ns.require.ErrorAs(err, &errLimit)
			ns.require.NotNil(errLimit)
			ns.require.Equal(string(models.Denied), errLimit.State)
		}

		// sleep time should be lesser than the window size, so all are sent within the time window
		sleepTime := time.Duration(conf.WSizeMs/int64(len(ns.notifications)+1)) * time.Millisecond
		time.Sleep(sleepTime)
	}

	return ns
}

func (ns *NotificationStage) the_service_sends_notifications_exceeding_the_time_window() *NotificationStage {
	for _, notif := range ns.notifications {
		conf := ns.conf.Limits.Get(notif.itself.Type)
		ns.assert.NotNil(conf)

		if err := ns.service.Send(context.Background(), notif.itself); err == nil {
			notif.isSent = true
		} else {
			fmt.Printf("error sending notification: %v \n", err)

			var errLimit *errs.ErrExceededRateLimit
			ns.require.ErrorAs(err, &errLimit)
			ns.require.NotNil(errLimit)
			ns.require.Equal(string(models.Denied), errLimit.State)
		}

		// news time window is 100ms and limit of 5, so it's going to be exceeded after ~5 notifications
		sleepTime := time.Duration(15) * time.Millisecond
		time.Sleep(sleepTime)
	}

	return ns
}

// then
func (ns *NotificationStage) all_the_notifications_have_been_sent() *NotificationStage {
	for _, n := range ns.notifications {
		ns.require.Truef(n.isSent, "all messages must have been sent")
	}

	ns.require.Equal(len(ns.notifications), int(ns.sentCount.Load()))

	return ns
}

func (ns *NotificationStage) some_notifications_have_been_sent() *NotificationStage {
	notSentIdxs := make([]int, 0)
	for i, n := range ns.notifications {
		if !n.isSent {
			notSentIdxs = append(notSentIdxs, i)
		}
	}

	ns.require.NotEmpty(notSentIdxs)
	for _, idx := range notSentIdxs {
		ns.require.Greater(idx, 0)
		ns.require.Less(idx, len(ns.notifications))
	}

	return ns
}

func (ns *NotificationStage) first_half_of_the_notifications_have_been_sent() *NotificationStage {
	ns.require.Equal(len(ns.notifications)/2, int(ns.sentCount.Load()))

	return ns
}
