package domain

import "github.com/google/uuid"

type ResultID struct {
	value uuid.UUID
}

func NewResultID(raw string) (ResultID, error) {
	if raw == "" {
		return ResultID{}, ErrResultIDEmpty
	}

	parsed, err := uuid.Parse(raw)
	if err != nil {
		return ResultID{}, ErrInvalidResultID
	}

	return ResultID{value: parsed}, nil
}

func (id ResultID) String() string {
	return id.value.String()
}
