package domain

import "github.com/google/uuid"

type CategoryID struct {
	value uuid.UUID
}

func NewCategoryID(raw string) (CategoryID, error) {
	if raw == "" {
		return CategoryID{}, ErrCategoryIDEmpty
	}
	if uuid.Validate(raw) != nil {
		return CategoryID{}, ErrInvalidCategoryID
	}

	parsed, err := uuid.Parse(raw)
	if err != nil {
		return CategoryID{}, ErrInvalidCategoryID
	}

	return CategoryID{value: parsed}, nil
}

func (id CategoryID) String() string {
	return id.value.String()
}
