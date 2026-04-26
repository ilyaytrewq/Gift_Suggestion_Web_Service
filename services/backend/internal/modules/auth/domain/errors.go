package domain

import "errors"

var (
	ErrEmailVerificationTokenIDEmpty   = errors.New("email verification token id is empty")
	ErrInvalidEmailVerificationTokenID = errors.New("email verification token id has invalid format")
	ErrSessionIDEmpty                  = errors.New("session id is empty")
	ErrInvalidSessionID                = errors.New("session id has invalid format")
	ErrResetTokenIDEmpty               = errors.New("password reset token id is empty")
	ErrInvalidResetTokenID             = errors.New("password reset token id has invalid format")
	ErrRefreshTokenHashEmpty           = errors.New("refresh token hash is empty")
	ErrEmailVerificationTokenHashEmpty = errors.New("email verification token hash is empty")
	ErrResetTokenHashEmpty             = errors.New("password reset token hash is empty")
	ErrEmailVerificationTokenNotFound  = errors.New("email verification token not found")
	ErrEmailVerificationTokenExpired   = errors.New("email verification token is expired")
	ErrEmailVerificationTokenUsed      = errors.New("email verification token is already used")
	ErrSessionNotFound                 = errors.New("auth session not found")
	ErrPasswordResetTokenNotFound      = errors.New("password reset token not found")
	ErrPasswordResetTokenExpired       = errors.New("password reset token is expired")
	ErrPasswordResetTokenUsed          = errors.New("password reset token is already used")
	ErrSessionExpired                  = errors.New("auth session is expired")
	ErrSessionRevoked                  = errors.New("auth session is revoked")
)
