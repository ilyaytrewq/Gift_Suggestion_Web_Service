package authsession

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/user"
	"github.com/pkg/errors"
)

type (
	AuthSessionID string
	CSRFSecret    string
	AuthSession   struct {
		AuthSessionID *AuthSessionID
		UserID        *user.UserID
		CSRFSecret    *CSRFSecret

		CreatedAt  time.Time
		ExpiresAt  time.Time
		LastUsedAt time.Time
	}
)

func NewAuthSession(userID *user.UserID, ttl time.Duration) (*AuthSession, error) {
	if !userID.IsValid() {
		return nil, ErrInvalidUserID
	}
	if ttl <= 0 {
		return nil, ErrInvalidTTL
	}

	id, err := NewAuthSessionID()
	if err != nil {
		return nil, err
	}
	secret, err := NewCSRFSecret()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &AuthSession{
		AuthSessionID: id,
		UserID:        userID,
		CSRFSecret:    secret,
		CreatedAt:     now,
		ExpiresAt:     now.Add(ttl),
		LastUsedAt:    now,
	}, nil
}

func NewAuthSessionID() (*AuthSessionID, error) {
	s, err := randomB64URL(32)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate auth session id")
	}
	id := AuthSessionID(s)
	return &id, nil
}

func NewAuthSessionIDFromString(s string) (*AuthSessionID, error) {
	id := AuthSessionID(s)
	if id.IsValid() != nil {
		return nil, ErrInvalidAuthSessionID
	}
	return &id, nil
}

func NewCSRFSecret() (*CSRFSecret, error) {
	s, err := randomB64URL(32)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate csrf secret")
	}
	id := CSRFSecret(s)
	return &id, nil
}

func NewCSRFSecretFromString(s string) (*CSRFSecret, error) {
	id := CSRFSecret(s)
	if id.IsValid() != nil {
		return nil, ErrInvalidCSRFSecret
	}
	return &id, nil
}

func randomB64URL(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
