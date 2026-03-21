package domain

import "github.com/google/uuid"

type UserID struct {
	value uuid.UUID
}

func newUserID(id string) (UserID, error) {
	if isBlank(id) {
		return UserID{}, ErrUserIDEmpty
	}
	if !isValidUserID(id) {
		return UserID{}, ErrInvalidUserID
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return UserID{}, ErrInvalidUserID
	}

	return UserID{value: uid}, nil
}

func (id UserID) IsValid() error {
	if isBlank(id.value.String()) {
		return ErrUserIDEmpty
	}

	if uuid.Validate(id.value.String()) != nil {
		return ErrInvalidUserID
	}

	return nil
}
