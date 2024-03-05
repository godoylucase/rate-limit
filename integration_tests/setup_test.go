package integration_tests

import (
	"fmt"
	"os"
	"rate-limit/internal/configs"
	"rate-limit/internal/infrastructure/gateway"
	"rate-limit/internal/notification"
	"rate-limit/internal/rate_limiter"
	"testing"

	"github.com/go-redis/redis/v8"
)

var (
	notificationSrv *notification.Service
)

func TestMain(m *testing.M) {
	conf, err := configs.LoadLimitConfigs("limit_configs.json")
	if err != nil {
		panic(fmt.Errorf("failed to load limit configuration: %w", err))
	}

	redisStorage := redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "", DB: 0})

	rlimiter := rate_limiter.NewSlidingWindowCounter(redisStorage)

	noOpGateway := &gateway.LogGW{}

	notificationSrv = notification.NewService(rlimiter, noOpGateway, conf)

	code := m.Run()

	os.Exit(code)
}
