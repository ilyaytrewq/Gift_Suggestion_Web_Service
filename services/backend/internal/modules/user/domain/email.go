package domain

type Email struct {
	value string
}

func newEmail(email string) (Email, error) {
	if isBlank(email) {
		return Email{}, ErrEmailEmpty
	}
	if isValidEmail(email) != nil {
		return Email{}, ErrInvalidEmail
	}

	return Email{value: email}, nil
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
