package usecase

import "errors"

var (
	ErrNilUserRepository        = errors.New("user repository is nil")
	ErrNilSessionRepository     = errors.New("session repository is nil")
	ErrNilPasswordResetRepo     = errors.New("password reset repository is nil")
	ErrNilAccessTokenManager    = errors.New("access token manager is nil")
	ErrNilRefreshTokenGenerator = errors.New("refresh token generator is nil")
	ErrNilResetTokenGenerator   = errors.New("password reset token generator is nil")
	ErrNilUserIDGenerator       = errors.New("user id generator is nil")
	ErrNilSessionIDGenerator    = errors.New("session id generator is nil")
	ErrNilResetTokenIDGenerator = errors.New("password reset token id generator is nil")
	ErrNilClock                 = errors.New("clock is nil")
)
