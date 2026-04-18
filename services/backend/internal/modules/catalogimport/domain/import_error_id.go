package domain

import "github.com/google/uuid"

type ImportErrorID struct {
	value uuid.UUID
}

func NewImportErrorID(raw string) (ImportErrorID, error) {
	if raw == "" {
		return ImportErrorID{}, ErrImportErrorIDEmpty
	}

	parsed, err := uuid.Parse(raw)
	if err != nil {
		return ImportErrorID{}, ErrInvalidImportErrorID
	}

	return ImportErrorID{value: parsed}, nil
}

func (id ImportErrorID) String() string {
	return id.value.String()
}
