package notification

import (
	"time"

	"github.com/segmentio/ksuid"
)

type Notification struct {
	Type    string
	UserID  ksuid.KSUID
	Message string
}

func isValid(notif *Notification) bool {
	if len(notif.Message) == 0 {
		return false
	}

	if notif.UserID.IsNil() {
		return false
	}

	if len(notif.Type) == 0 {
		return false
	}

	return true
}

type CheckLimitRequest struct {
	Key     string
	Limit   int64
	TWindow time.Duration
}
