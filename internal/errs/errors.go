package errs

import (
	"errors"
)

var (
	ErrExceededRateLimit = errors.New("exceeded rate limit")
	ErrInvalidArguments  = errors.New("invalid arguments")
)
