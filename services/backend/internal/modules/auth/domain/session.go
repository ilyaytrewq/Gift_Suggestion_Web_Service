package domain

import (
	"time"

	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
)

type Session struct {
	id               SessionID
	userID           userdomain.UserID
	refreshTokenHash string
	expiresAt        time.Time
	createdAt        time.Time
	updatedAt        time.Time
	lastUsedAt       *time.Time
	revokedAt        *time.Time
}

func NewSession(
	id SessionID,
	userID userdomain.UserID,
	refreshTokenHash string,
	now time.Time,
	expiresAt time.Time,
) (Session, error) {
	if refreshTokenHash == "" {
		return Session{}, ErrRefreshTokenHashEmpty
	}

	timestamp := now.UTC()

	return Session{
		id:               id,
		userID:           userID,
		refreshTokenHash: refreshTokenHash,
		expiresAt:        expiresAt.UTC(),
		createdAt:        timestamp,
		updatedAt:        timestamp,
	}, nil
}

func RestoreSession(
	id string,
	userID string,
	refreshTokenHash string,
	expiresAt time.Time,
	createdAt time.Time,
	updatedAt time.Time,
	lastUsedAt *time.Time,
	revokedAt *time.Time,
) (Session, error) {
	sessionID, err := NewSessionID(id)
	if err != nil {
		return Session{}, err
	}

	uid, err := userdomain.NewUserID(userID)
	if err != nil {
		return Session{}, err
	}

	if refreshTokenHash == "" {
		return Session{}, ErrRefreshTokenHashEmpty
	}

	return Session{
		id:               sessionID,
		userID:           uid,
		refreshTokenHash: refreshTokenHash,
		expiresAt:        expiresAt.UTC(),
		createdAt:        createdAt.UTC(),
		updatedAt:        updatedAt.UTC(),
		lastUsedAt:       cloneTimePtr(lastUsedAt),
		revokedAt:        cloneTimePtr(revokedAt),
	}, nil
}

func (s *Session) Rotate(refreshTokenHash string, now, expiresAt time.Time) error {
	if refreshTokenHash == "" {
		return ErrRefreshTokenHashEmpty
	}

	timestamp := now.UTC()
	s.refreshTokenHash = refreshTokenHash
	s.expiresAt = expiresAt.UTC()
	s.updatedAt = timestamp
	s.lastUsedAt = &timestamp

	return nil
}

func (s *Session) Revoke(now time.Time) {
	timestamp := now.UTC()
	s.revokedAt = &timestamp
	s.updatedAt = timestamp
}

func (s Session) IsExpired(now time.Time) bool {
	return now.UTC().After(s.expiresAt)
}

func (s Session) IsRevoked() bool {
	return s.revokedAt != nil
}

func (s Session) ID() SessionID {
	return s.id
}

func (s Session) UserID() userdomain.UserID {
	return s.userID
}

func (s Session) RefreshTokenHash() string {
	return s.refreshTokenHash
}

func (s Session) ExpiresAt() time.Time {
	return s.expiresAt
}

func (s Session) CreatedAt() time.Time {
	return s.createdAt
}

func (s Session) UpdatedAt() time.Time {
	return s.updatedAt
}

func (s Session) LastUsedAt() *time.Time {
	return cloneTimePtr(s.lastUsedAt)
}

func (s Session) RevokedAt() *time.Time {
	return cloneTimePtr(s.revokedAt)
}
