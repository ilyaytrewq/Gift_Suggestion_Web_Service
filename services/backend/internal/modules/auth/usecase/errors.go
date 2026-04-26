package usecase

import "errors"

var (
	ErrNilUserRepository               = errors.New("user repository is nil")
	ErrNilRegistrationRepo             = errors.New("registration repository is nil")
	ErrNilEmailVerificationRepo        = errors.New("email verification repository is nil")
	ErrNilSessionRepository            = errors.New("session repository is nil")
	ErrNilPasswordResetRepo            = errors.New("password reset repository is nil")
	ErrNilAccessTokenManager           = errors.New("access token manager is nil")
	ErrNilRefreshTokenGenerator        = errors.New("refresh token generator is nil")
	ErrNilVerificationTokenGenerator   = errors.New("email verification token generator is nil")
	ErrNilResetTokenGenerator          = errors.New("password reset token generator is nil")
	ErrNilEmailNotifier                = errors.New("auth email notifier is nil")
	ErrNilLogger                       = errors.New("logger is nil")
	ErrNilUserIDGenerator              = errors.New("user id generator is nil")
	ErrNilSessionIDGenerator           = errors.New("session id generator is nil")
	ErrNilResetTokenIDGenerator        = errors.New("password reset token id generator is nil")
	ErrNilVerificationTokenIDGenerator = errors.New("email verification token id generator is nil")
	ErrNilClock                        = errors.New("clock is nil")
)
