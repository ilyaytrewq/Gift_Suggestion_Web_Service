package domain

type Password struct {
	value string
}

func NewPassword(password string) (Password, error) {
	return newPassword(password)
}

func newPassword(password string) (Password, error) {
	if err := isValidPassword(password); err != nil {
		return Password{}, err
	}

	return Password{value: password}, nil
}

func (p Password) IsValid() bool {
	return isValidPassword(p.value) == nil
}

func (p Password) String() string {
	return p.value
}
