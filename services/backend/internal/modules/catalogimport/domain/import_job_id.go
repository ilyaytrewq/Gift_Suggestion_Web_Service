package domain

import "github.com/google/uuid"

type ImportJobID struct {
	value uuid.UUID
}

func NewImportJobID(raw string) (ImportJobID, error) {
	if raw == "" {
		return ImportJobID{}, ErrImportJobIDEmpty
	}

	parsed, err := uuid.Parse(raw)
	if err != nil {
		return ImportJobID{}, ErrInvalidImportJobID
	}

	return ImportJobID{value: parsed}, nil
}

func (id ImportJobID) String() string {
	return id.value.String()
}
