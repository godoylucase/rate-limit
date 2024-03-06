package integration_tests

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type NotificationServiceSuite struct {
	suite.Suite
}

func TestNotificationServiceSuite(t *testing.T) {
	suite.Run(t, new(NotificationServiceSuite))
}

func (ns *NotificationServiceSuite) TestSendNotificationsNoRateLimited_SlidingWindowRateLimiter() {
	given, when, then := NotificationServiceTestStages(ns.T())

	given.
		a_rate_limit_configuration().and().
		a_no_op_gateway().and().
		a_redis_sliding_window_rate_limiter().and().
		a_notification_service().and().
		status_notifications_group_with_limit_size()

	when.
		the_service_sends_notifications()

	then.
		all_the_notifications_have_been_sent()
}

func (ns *NotificationServiceSuite) TestSendNotificationsRateLimited_SlidingWindowRateLimiter() {
	given, when, then := NotificationServiceTestStages(ns.T())

	given.
		a_rate_limit_configuration().and().
		a_no_op_gateway().and().
		a_redis_sliding_window_rate_limiter().and().
		a_notification_service().and().
		status_notifications_group_with_twice_limit_size()

	when.
		the_service_sends_notifications()

	then.
		half_of_the_notifications_have_been_sent()
}

func (ns *NotificationServiceSuite) TestSendNotificationsNoRateLimited_FixedWindowRateLimiter() {
	given, when, then := NotificationServiceTestStages(ns.T())

	given.
		a_rate_limit_configuration().and().
		a_no_op_gateway().and().
		a_redis_fixed_window_rate_limiter().and().
		a_notification_service().and().
		status_notifications_group_with_limit_size()

	when.
		the_service_sends_notifications()

	then.
		all_the_notifications_have_been_sent()
}

func (ns *NotificationServiceSuite) TestSendNotificationsRateLimited_FixedWindowRateLimiter() {
	given, when, then := NotificationServiceTestStages(ns.T())

	given.
		a_rate_limit_configuration().and().
		a_no_op_gateway().and().
		a_redis_fixed_window_rate_limiter().and().
		a_notification_service().and().
		status_notifications_group_with_twice_limit_size()

	when.
		the_service_sends_notifications()

	then.
		half_of_the_notifications_have_been_sent()
}
