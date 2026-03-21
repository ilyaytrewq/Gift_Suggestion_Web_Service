package domain

type Password struct {
	value string
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
