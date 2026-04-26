package domain

import (
	"time"

	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
)

type PasswordResetToken struct {
	id        PasswordResetTokenID
	userID    userdomain.UserID
	tokenHash string
	expiresAt time.Time
	createdAt time.Time
	usedAt    *time.Time
}

func NewPasswordResetToken(
	id PasswordResetTokenID,
	userID userdomain.UserID,
	tokenHash string,
	now time.Time,
	expiresAt time.Time,
) (PasswordResetToken, error) {
	if tokenHash == "" {
		return PasswordResetToken{}, ErrResetTokenHashEmpty
	}

	timestamp := now.UTC()

	return PasswordResetToken{
		id:        id,
		userID:    userID,
		tokenHash: tokenHash,
		expiresAt: expiresAt.UTC(),
		createdAt: timestamp,
	}, nil
}

func RestorePasswordResetToken(
	id string,
	userID string,
	tokenHash string,
	expiresAt time.Time,
	createdAt time.Time,
	usedAt *time.Time,
) (PasswordResetToken, error) {
	tokenID, err := NewPasswordResetTokenID(id)
	if err != nil {
		return PasswordResetToken{}, err
	}

	uid, err := userdomain.NewUserID(userID)
	if err != nil {
		return PasswordResetToken{}, err
	}

	if tokenHash == "" {
		return PasswordResetToken{}, ErrResetTokenHashEmpty
	}

	return PasswordResetToken{
		id:        tokenID,
		userID:    uid,
		tokenHash: tokenHash,
		expiresAt: expiresAt.UTC(),
		createdAt: createdAt.UTC(),
		usedAt:    cloneTimePtr(usedAt),
	}, nil
}

func (t *PasswordResetToken) MarkUsed(now time.Time) {
	if t == nil {
		return
	}

	timestamp := now.UTC()
	t.usedAt = &timestamp
}

func (t PasswordResetToken) IsExpired(now time.Time) bool {
	return now.UTC().After(t.expiresAt)
}

func (t PasswordResetToken) IsUsed() bool {
	return t.usedAt != nil
}

func (t PasswordResetToken) ID() PasswordResetTokenID {
	return t.id
}

func (t PasswordResetToken) UserID() userdomain.UserID {
	return t.userID
}

func (t PasswordResetToken) TokenHash() string {
	return t.tokenHash
}

func (t PasswordResetToken) ExpiresAt() time.Time {
	return t.expiresAt
}

func (t PasswordResetToken) CreatedAt() time.Time {
	return t.createdAt
}

func (t PasswordResetToken) UsedAt() *time.Time {
	return cloneTimePtr(t.usedAt)
}
