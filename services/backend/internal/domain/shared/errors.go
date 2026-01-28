package shared

import (
	"errors"
)

var (
	ErrInvalidAgeLimit = errors.New("invalid age limit")
	ErrInvalidPrice    = errors.New("invalid price")
)
