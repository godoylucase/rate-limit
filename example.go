/*
Package main is the main package of the rate-limit example program.
It demonstrates how to use a rate limiter to send notifications to users.
*/
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/godoylucase/rate-limit/configs"
	"github.com/godoylucase/rate-limit/models"
	"github.com/godoylucase/rate-limit/notification"
	"github.com/godoylucase/rate-limit/rate_limiter"
	"github.com/segmentio/ksuid"

	"github.com/go-redis/redis/v8"
)

const (
	notificationCount = 9999
)

type gateway struct{}

func (g *gateway) Send(ctx context.Context, userID string, message string) error {
	log.Printf("Sending message to user %v: %v", userID, message)
	return nil
}


func main() {
	ctx := context.Background()

	conf, err := configs.Load("example_config.json")
	if err != nil {
		log.Panicf("failed to load configurations: %v", err)
	}

	redisCli := redis.NewClient(&redis.Options{Addr: conf.RedisAddr, Password: "", DB: 0})
	rateLimiter := rate_limiter.Get(conf.RateLimiterType, redisCli)

	srv := notification.NewService(rateLimiter, &gateway{}, conf.Limits)

	userID := ksuid.New()
	for i := 0; i < notificationCount; i++ {
		if err := srv.Send(ctx, &models.Notification{
			UserID:  userID,
			Type:    "status",
			Message: fmt.Sprintf("Hello, world! [%v]", i),
		}); err != nil {
			log.Printf("failed to send notification: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
