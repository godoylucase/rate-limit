package errs

import (
	"errors"
)

var (
	ErrExceededRateLimit = errors.New("exceeded rate limit")
	ErrInternal          = errors.New("internal error")
	ErrInvalidArguments  = errors.New("invalid arguments")
)
