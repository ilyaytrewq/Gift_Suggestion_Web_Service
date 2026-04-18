package domain

import "github.com/google/uuid"

type GiftID struct {
	value uuid.UUID
}

func NewGiftID(raw string) (GiftID, error) {
	if raw == "" {
		return GiftID{}, ErrGiftIDEmpty
	}
	if uuid.Validate(raw) != nil {
		return GiftID{}, ErrInvalidGiftID
	}

	parsed, err := uuid.Parse(raw)
	if err != nil {
		return GiftID{}, ErrInvalidGiftID
	}

	return GiftID{value: parsed}, nil
}

func (id GiftID) String() string {
	return id.value.String()
}
