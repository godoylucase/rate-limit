package notification

import (
	"github.com/segmentio/ksuid"
)

// Notification represents a notification with a type, user ID, and message.
type Notification struct {
	Type    string         // Type of the notification
	UserID  ksuid.KSUID    // User ID associated with the notification
	Message string         // Message content of the notification
}

// isValid checks if a notification is valid.
// A notification is considered valid if it has a non-empty message,
// a non-nil user ID, and a non-empty type.
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
