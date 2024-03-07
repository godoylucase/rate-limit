package errs

import (
	"errors"
	"fmt"
)

// ErrInvalidArguments is an error indicating that the provided arguments are invalid.
var ErrInvalidArguments = errors.New("invalid arguments")

// ErrInternalError is an error indicating an internal error occurred.
var ErrInternalError = errors.New("internal error")

// ErrExceededRateLimit is an error indicating that the rate limit has been exceeded.
type ErrExceededRateLimit struct {
	State     string // The state associated with the rate limit.
	Count     int    // The number of requests made within the rate limit.
	ExpiresAt int64  // The timestamp when the rate limit expires.
}

// Error returns the string representation of the ErrExceededRateLimit error.
func (e *ErrExceededRateLimit) Error() string {
	return fmt.Sprintf("rate limit exceeded: state=%v, count=%v, expiresAt=%v", e.State, e.Count, e.ExpiresAt)
}
