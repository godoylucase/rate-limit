package gateway

import (
	"context"
	"log"
	"time"
)

type LogGW struct{}

func (gateway *LogGW) Send(ctx context.Context, userID string, message string) error {
	log.Printf("notification request sent to userID [%v] at time: %v \n", userID, time.Now().UnixMilli())
	return nil
}
