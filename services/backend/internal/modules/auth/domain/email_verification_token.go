package domain

import (
	"time"

	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
)

type EmailVerificationToken struct {
	id        EmailVerificationTokenID
	userID    userdomain.UserID
	tokenHash string
	expiresAt time.Time
	createdAt time.Time
	usedAt    *time.Time
}

func NewEmailVerificationToken(
	id EmailVerificationTokenID,
	userID userdomain.UserID,
	tokenHash string,
	now time.Time,
	expiresAt time.Time,
) (EmailVerificationToken, error) {
	if tokenHash == "" {
		return EmailVerificationToken{}, ErrEmailVerificationTokenHashEmpty
	}

	timestamp := now.UTC()

	return EmailVerificationToken{
		id:        id,
		userID:    userID,
		tokenHash: tokenHash,
		expiresAt: expiresAt.UTC(),
		createdAt: timestamp,
	}, nil
}

func RestoreEmailVerificationToken(
	id string,
	userID string,
	tokenHash string,
	expiresAt time.Time,
	createdAt time.Time,
	usedAt *time.Time,
) (EmailVerificationToken, error) {
	tokenID, err := NewEmailVerificationTokenID(id)
	if err != nil {
		return EmailVerificationToken{}, err
	}

	uid, err := userdomain.NewUserID(userID)
	if err != nil {
		return EmailVerificationToken{}, err
	}

	if tokenHash == "" {
		return EmailVerificationToken{}, ErrEmailVerificationTokenHashEmpty
	}

	return EmailVerificationToken{
		id:        tokenID,
		userID:    uid,
		tokenHash: tokenHash,
		expiresAt: expiresAt.UTC(),
		createdAt: createdAt.UTC(),
		usedAt:    cloneTimePtr(usedAt),
	}, nil
}

func (t *EmailVerificationToken) MarkUsed(now time.Time) {
	if t == nil {
		return
	}

	timestamp := now.UTC()
	t.usedAt = &timestamp
}

func (t EmailVerificationToken) IsExpired(now time.Time) bool {
	return now.UTC().After(t.expiresAt)
}

func (t EmailVerificationToken) IsUsed() bool {
	return t.usedAt != nil
}

func (t EmailVerificationToken) ID() EmailVerificationTokenID {
	return t.id
}

func (t EmailVerificationToken) UserID() userdomain.UserID {
	return t.userID
}

func (t EmailVerificationToken) TokenHash() string {
	return t.tokenHash
}

func (t EmailVerificationToken) ExpiresAt() time.Time {
	return t.expiresAt
}

func (t EmailVerificationToken) CreatedAt() time.Time {
	return t.createdAt
}

func (t EmailVerificationToken) UsedAt() *time.Time {
	return cloneTimePtr(t.usedAt)
}
