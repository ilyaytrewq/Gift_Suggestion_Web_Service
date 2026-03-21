package domain

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type PasswordHash struct {
	value []byte
}

func newPasswordHash(password Password) (PasswordHash, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password.value), bcrypt.DefaultCost)
	if err != nil {
		return PasswordHash{}, errors.Wrap(err, "failed to generate password hash")
	}

	return PasswordHash{value: hash}, nil
}
