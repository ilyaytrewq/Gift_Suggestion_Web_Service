package domain

import "github.com/google/uuid"

type ConnectionID struct {
	value uuid.UUID
}

func NewConnectionID(raw string) (ConnectionID, error) {
	if raw == "" {
		return ConnectionID{}, ErrConnectionIDEmpty
	}

	if uuid.Validate(raw) != nil {
		return ConnectionID{}, ErrInvalidConnectionID
	}

	parsed, err := uuid.Parse(raw)
	if err != nil {
		return ConnectionID{}, ErrInvalidConnectionID
	}

	return ConnectionID{value: parsed}, nil
}

func (id ConnectionID) String() string {
	return id.value.String()
}
