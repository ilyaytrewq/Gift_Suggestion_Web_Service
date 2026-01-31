package authsession

import (
	"encoding/base64"

	"github.com/pkg/errors"
)

const tokenBytesLen = 32

func (s AuthSessionID) IsValid() error {
	if err := validateToken(string(s)); err != nil {
		return errors.Wrap(ErrInvalidAuthSessionID, err.Error())
	}
	return nil
}

func (s CSRFSecret) IsValid() error {
	if err := validateToken(string(s)); err != nil {
		return errors.Wrap(ErrInvalidCSRFSecret, err.Error())
	}
	return nil
}

func validateToken(s string) error {
	decoded, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return ErrInvalidB64string
	}
	if len(decoded) != tokenBytesLen {
		return ErrInvalidLength
	}

	return nil
}
