package domain

import "errors"

var (
	ErrSessionIDEmpty            = errors.New("session id is empty")
	ErrInvalidSessionID          = errors.New("session id has invalid format")
	ErrResetTokenIDEmpty         = errors.New("password reset token id is empty")
	ErrInvalidResetTokenID       = errors.New("password reset token id has invalid format")
	ErrRefreshTokenHashEmpty     = errors.New("refresh token hash is empty")
	ErrResetTokenHashEmpty       = errors.New("password reset token hash is empty")
	ErrSessionNotFound           = errors.New("auth session not found")
	ErrPasswordResetTokenExpired = errors.New("password reset token is expired")
	ErrSessionExpired            = errors.New("auth session is expired")
	ErrSessionRevoked            = errors.New("auth session is revoked")
)
