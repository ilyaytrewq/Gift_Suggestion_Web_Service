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
