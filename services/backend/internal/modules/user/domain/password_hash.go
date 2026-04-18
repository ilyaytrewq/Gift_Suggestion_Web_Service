package domain

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type PasswordHash struct {
	value []byte
}

func NewPasswordHash(password Password) (PasswordHash, error) {
	return newPasswordHash(password)
}

func newPasswordHash(password Password) (PasswordHash, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password.value), bcrypt.DefaultCost)
	if err != nil {
		return PasswordHash{}, errors.Wrap(err, "failed to generate password hash")
	}

	return PasswordHash{value: hash}, nil
}

func RestorePasswordHash(hash string) (PasswordHash, error) {
	if isBlank(hash) {
		return PasswordHash{}, ErrPasswordHashEmpty
	}

	return PasswordHash{value: []byte(hash)}, nil
}

func (h PasswordHash) String() string {
	return string(h.value)
}

func (h PasswordHash) Compare(password Password) bool {
	return bcrypt.CompareHashAndPassword(h.value, []byte(password.value)) == nil
}
