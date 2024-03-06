package gateway

import (
	"context"
	"log"
	"time"
)

// LogGW is a struct representing a logging gateway.
type LogGW struct{}

// Send sends a notification request to the specified user ID with the given message.
// It logs the notification request and returns nil.
func (gateway *LogGW) Send(ctx context.Context, userID string, message string) error {
	log.Printf("notification request sent to userID [%v] at time: %v \n", userID, time.Now().UnixMilli())
	return nil
}
