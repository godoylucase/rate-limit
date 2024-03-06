package notification

import (
	"testing"

	"github.com/segmentio/ksuid"
)

func TestIsValid(t *testing.T) {
	tests := []struct {
		name     string
		notif    *Notification
		expected bool
	}{
		{
			name: "valid notification",
			notif: &Notification{
				Message: "Test message",
				UserID:  ksuid.New(),
				Type:    "Test type",
			},
			expected: true,
		},
		{
			name: "empty message",
			notif: &Notification{
				Message: "",
				UserID:  ksuid.New(),
				Type:    "Test type",
			},
			expected: false,
		},
		{
			name: "nil user ID",
			notif: &Notification{
				Message: "Test message",
				UserID:  ksuid.Nil,
				Type:    "Test type",
			},
			expected: false,
		},
		{
			name: "empty type",
			notif: &Notification{
				Message: "Test message",
				UserID:  ksuid.New(),
				Type:    "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValid(tt.notif); got != tt.expected {
				t.Errorf("isValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}
