package usecase

import "errors"

var (
	ErrNilUserRepository = errors.New("user repository is nil")
	ErrNilClock          = errors.New("clock is nil")
)
