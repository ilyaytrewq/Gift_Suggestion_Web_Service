package domain

import (
	"github.com/google/uuid"
)

type SessionID struct {
	value uuid.UUID
}

func NewSessionID(id string) (SessionID, error) {
	if id == "" {
		return SessionID{}, ErrSessionIDEmpty
	}
	if uuid.Validate(id) != nil {
		return SessionID{}, ErrInvalidSessionID
	}

	parsed, err := uuid.Parse(id)
	if err != nil {
		return SessionID{}, ErrInvalidSessionID
	}

	return SessionID{value: parsed}, nil
}

func (id SessionID) String() string {
	return id.value.String()
}
