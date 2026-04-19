package domain

import "github.com/google/uuid"

type RequestID struct {
	value uuid.UUID
}

func NewRequestID(raw string) (RequestID, error) {
	if raw == "" {
		return RequestID{}, ErrRequestIDEmpty
	}

	parsed, err := uuid.Parse(raw)
	if err != nil {
		return RequestID{}, ErrInvalidRequestID
	}

	return RequestID{value: parsed}, nil
}

func (id RequestID) String() string {
	return id.value.String()
}
