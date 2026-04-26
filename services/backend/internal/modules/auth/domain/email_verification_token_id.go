package domain

import "github.com/google/uuid"

type EmailVerificationTokenID struct {
	value uuid.UUID
}

func NewEmailVerificationTokenID(id string) (EmailVerificationTokenID, error) {
	if id == "" {
		return EmailVerificationTokenID{}, ErrEmailVerificationTokenIDEmpty
	}

	parsed, err := uuid.Parse(id)
	if err != nil {
		return EmailVerificationTokenID{}, ErrInvalidEmailVerificationTokenID
	}
	if parsed == uuid.Nil {
		return EmailVerificationTokenID{}, ErrInvalidEmailVerificationTokenID
	}

	return EmailVerificationTokenID{value: parsed}, nil
}

func (id EmailVerificationTokenID) String() string {
	return id.value.String()
}
