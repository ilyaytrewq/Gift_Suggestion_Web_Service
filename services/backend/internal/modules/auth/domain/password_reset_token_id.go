package domain

import "github.com/google/uuid"

type PasswordResetTokenID struct {
	value uuid.UUID
}

func NewPasswordResetTokenID(id string) (PasswordResetTokenID, error) {
	if id == "" {
		return PasswordResetTokenID{}, ErrResetTokenIDEmpty
	}
	if uuid.Validate(id) != nil {
		return PasswordResetTokenID{}, ErrInvalidResetTokenID
	}

	parsed, err := uuid.Parse(id)
	if err != nil {
		return PasswordResetTokenID{}, ErrInvalidResetTokenID
	}

	return PasswordResetTokenID{value: parsed}, nil
}

func (id PasswordResetTokenID) String() string {
	return id.value.String()
}
