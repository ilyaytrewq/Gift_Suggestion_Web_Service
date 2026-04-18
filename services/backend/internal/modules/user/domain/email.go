package domain

import "strings"

type Email struct {
	value string
}

func NewEmail(email string) (Email, error) {
	return newEmail(email)
}

func newEmail(email string) (Email, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if isBlank(normalized) {
		return Email{}, ErrEmailEmpty
	}
	if isValidEmail(normalized) != nil {
		return Email{}, ErrInvalidEmail
	}

	return Email{value: normalized}, nil
}

func (e Email) IsValid() error {
	if isBlank(e.value) {
		return ErrEmailEmpty
	}
	if isValidEmail(e.value) != nil {
		return ErrInvalidEmail
	}

	return nil
}

func (e Email) String() string {
	return e.value
}
