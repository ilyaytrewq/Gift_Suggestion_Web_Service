package user

import (
	"net/mail"
	"strings"

	"github.com/google/uuid"
)

func isValidUserID(id string) bool {
	return uuid.Validate(id) == nil
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func isBlank(value string) bool {
	return strings.TrimSpace(value) == ""
}

const (
	minPasswordLen = 8
	maxPasswordLen = 72
)
