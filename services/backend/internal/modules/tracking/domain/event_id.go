package domain

import "github.com/google/uuid"

type EventID struct {
	value uuid.UUID
}

func NewEventID(raw string) (EventID, error) {
	if raw == "" {
		return EventID{}, ErrEventIDEmpty
	}
	if uuid.Validate(raw) != nil {
		return EventID{}, ErrInvalidEventID
	}

	parsed, err := uuid.Parse(raw)
	if err != nil {
		return EventID{}, ErrInvalidEventID
	}

	return EventID{value: parsed}, nil
}

func (id EventID) String() string {
	return id.value.String()
}
