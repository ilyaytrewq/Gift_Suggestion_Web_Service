package auth

import (
	"errors"
)

var (
	ErrNilSessionRepository = errors.New("nil session repository")
	ErrNilUserRepository    = errors.New("nil user repository")
	ErrInvalidTTL           = errors.New("invalid TTL")
)
