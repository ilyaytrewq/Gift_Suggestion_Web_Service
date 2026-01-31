package session

import "errors"

var (
	ErrSessionIDEmpty     = errors.New("session id is empty")
	ErrInvalidSessionID   = errors.New("session id has invalid format")
	ErrInvalidOwnerUserID = errors.New("owner user id has invalid format")
	ErrInvalidSnapshot    = errors.New("invalid snapshot")
)
