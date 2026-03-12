package auth

import (
	"context"
	"errors"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/authsession"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/user"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserWrongPassword = errors.New("wrong password")

	ErrSessionNotFound   = errors.New("session not found")
	ErrInvalidCSRFSecret = errors.New("invalid CSRF secret")
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email user.Email) (*user.User, error)
}

type SessionRepository interface {
	Save(ctx context.Context, authSession *authsession.AuthSession) error
	Get(ctx context.Context, authSessionID *authsession.AuthSessionID, secret *authsession.CSRFSecret) (*authsession.AuthSession, error)
	Delete(ctx context.Context, authSessionID *authsession.AuthSessionID) error
}
