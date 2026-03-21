package domain

import (
	"net/mail"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

func isBlank(value string) bool {
	return strings.TrimSpace(value) == ""
}

func isValidUserID(id string) bool {
	return uuid.Validate(id) == nil
}

func isValidEmail(email string) error {
	_, err := mail.ParseAddress(email)
	return err
}

const (
	minPasswordLen = 8
	maxPasswordLen = 72
)

func isValidPassword(password string) error {
	if len(password) < minPasswordLen {
		return ErrPasswordTooShort
	}
	if len(password) > maxPasswordLen {
		return ErrPasswordTooLong
	}

	var hasLower, hasUpper, hasDigit, hasSpecial bool

	for _, r := range password {
		switch {
		case unicode.IsSpace(r):
			return ErrPasswordContainsSpace
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	if !hasLower {
		return ErrPasswordNoLower
	}
	if !hasUpper {
		return ErrPasswordNoUpper
	}
	if !hasDigit {
		return ErrPasswordNoDigit
	}
	if !hasSpecial {
		return ErrPasswordNoSpecial
	}

	return nil
}
