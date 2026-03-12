package auth

import (
	"context"
	"time"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/authsession"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/user"
	"github.com/pkg/errors"
)

type AuthUseСase struct {
	sessionRepository SessionRepository
	userRepository    UserRepository
	ttl               time.Duration
}

func NewAuthUseСase(sessionRepository SessionRepository, userRepository UserRepository, ttl time.Duration) (*AuthUseСase, error) {
	if sessionRepository == nil {
		return nil, ErrNilSessionRepository
	}
	if userRepository == nil {
		return nil, ErrNilUserRepository
	}
	if ttl <= 0 {
		return nil, ErrInvalidTTL
	}
	return &AuthUseСase{
		sessionRepository: sessionRepository,
		userRepository:    userRepository,
		ttl:               ttl,
	}, nil
}

func (au *AuthUseСase) Login(ctx context.Context, input LoginInput) (*authsession.AuthSession, error) {
	email, err := user.NewEmail(input.Email)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new email")
	}

	acc, err := au.userRepository.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, errors.Wrap(err, "failed to get user by email")
	}

	password, err := user.NewPassword(input.Password)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new password")
	}

	if !acc.ComparePassword(password) {
		return nil, ErrUserWrongPassword
	}

	session, err := authsession.NewAuthSession(acc.ID(), au.ttl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create auth session")
	}

	err = au.sessionRepository.Save(ctx, session)
	if err != nil {
		return nil, errors.Wrap(err, "failed to save session")
	}

	return session, nil
}

func (au *AuthUseСase) IsAuthorized(ctx context.Context, input IsAuthorizedInput) (bool, error) {
	userID, err := user.NewUserID(input.UserID)
	if err != nil {
		return false, errors.Wrap(err, "failed to create new user id")
	}

	authSessionID, err := authsession.NewAuthSessionIDFromString(input.AuthSessionID)
	if err != nil {
		return false, errors.Wrap(err, "failed to create auth session id")
	}

	csrfSecret, err := authsession.NewCSRFSecretFromString(input.CSRFSecret)
	if err != nil {
		return false, errors.Wrap(err, "failed to create CSRF secret")
	}

	session, err := au.sessionRepository.Get(ctx, authSessionID, csrfSecret)
	if err != nil {
		return false, errors.Wrap(err, "failed to get session")
	}

	if session.UserID == nil || session.UserID != userID {
		return false, nil
	}
	if time.Now().After(session.ExpiresAt) {
		return false, nil
	}

	return true, nil
}

func (au *AuthUseСase) Logout(ctx context.Context, input LogoutInput) error {
	authSessionID, err := authsession.NewAuthSessionIDFromString(input.AuthSessionID)
	if err != nil {
		return errors.Wrap(err, "failed to create auth session ID")
	}

	csrfSecret, err := authsession.NewCSRFSecretFromString(input.CSRFSecret)
	if err != nil {
		return errors.Wrap(err, "failed to create CSRF secret")
	}

	session, err := au.sessionRepository.Get(ctx, authSessionID, csrfSecret)
	if err != nil {
		if errors.As(err, ErrSessionNotFound) {
			return ErrSessionNotFound
		}
		return errors.Wrap(err, "failed to get session")
	}

	err = au.sessionRepository.Delete(ctx, session.AuthSessionID)
	if err != nil {
		return errors.Wrap(err, "failed to delete session")
	}

	return nil
}
