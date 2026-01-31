package authsession

import (
	"errors"
)

var (
	ErrInvalidAuthSessionID = errors.New("invalid auth session id")
	ErrInvalidCSRFSecret    = errors.New("invalid CSRF secret")
	ErrInvalidUserID        = errors.New("invalid user id")
	ErrInvalidTTL           = errors.New("invalid ttl")

	ErrInvalidLength    = errors.New("invalid length")
	ErrInvalidB64string = errors.New("invalid b64 string")
)
