package session

import (
	"strings"

	"github.com/google/uuid"
)

func (id SessionID) IsValid() bool {
	return isValidSessionID(string(id))
}

func isValidSessionID(id string) bool {
	return uuid.Validate(id) == nil
}

func isBlank(value string) bool {
	return strings.TrimSpace(value) == ""
}
