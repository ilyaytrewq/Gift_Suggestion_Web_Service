package domain

import (
	"strings"
	"time"
)

const (
	defaultInterestSource        = "vk_profile"
	maxImportedInterestLength    = 128
	maxInterestSourceLabelLength = 32
)

type ImportedInterest struct {
	rawValue        string
	normalizedValue string
	sourceLabel     string
	position        int
	importedAt      time.Time
}

func NewImportedInterest(rawValue, sourceLabel string, position int, importedAt time.Time) (ImportedInterest, error) {
	normalizedRaw := normalizeInterestValue(rawValue)
	if normalizedRaw == "" {
		return ImportedInterest{}, ErrImportedInterestEmpty
	}
	if len(normalizedRaw) > maxImportedInterestLength {
		return ImportedInterest{}, ErrImportedInterestTooLong
	}

	normalizedValue := strings.ToLower(normalizedRaw)
	if normalizedValue == "" {
		return ImportedInterest{}, ErrNormalizedInterestEmpty
	}

	normalizedSource := normalizeInterestSource(sourceLabel)
	if normalizedSource == "" {
		return ImportedInterest{}, ErrInvalidInterestSource
	}

	if position < 1 {
		return ImportedInterest{}, ErrInvalidInterestPosition
	}
	if importedAt.IsZero() {
		return ImportedInterest{}, ErrImportedInterestAtZero
	}

	return ImportedInterest{
		rawValue:        normalizedRaw,
		normalizedValue: normalizedValue,
		sourceLabel:     normalizedSource,
		position:        position,
		importedAt:      importedAt.UTC(),
	}, nil
}

func RestoreImportedInterest(rawValue, normalizedValue, sourceLabel string, position int, importedAt time.Time) (ImportedInterest, error) {
	interest, err := NewImportedInterest(rawValue, sourceLabel, position, importedAt)
	if err != nil {
		return ImportedInterest{}, err
	}

	normalized := normalizeInterestValue(normalizedValue)
	if normalized == "" {
		return ImportedInterest{}, ErrNormalizedInterestEmpty
	}
	if len(normalized) > maxImportedInterestLength {
		return ImportedInterest{}, ErrImportedInterestTooLong
	}

	interest.normalizedValue = strings.ToLower(normalized)
	return interest, nil
}

func (i ImportedInterest) RawValue() string {
	return i.rawValue
}

func (i ImportedInterest) NormalizedValue() string {
	return i.normalizedValue
}

func (i ImportedInterest) SourceLabel() string {
	return i.sourceLabel
}

func (i ImportedInterest) Position() int {
	return i.position
}

func (i ImportedInterest) ImportedAt() time.Time {
	return i.importedAt
}

func normalizeInterestValue(raw string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
}

func normalizeInterestSource(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultInterestSource
	}
	if len(value) > maxInterestSourceLabelLength {
		return ""
	}

	return value
}
