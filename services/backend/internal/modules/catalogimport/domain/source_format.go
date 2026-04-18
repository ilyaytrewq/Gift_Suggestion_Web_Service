package domain

import (
	"path/filepath"
	"strings"
)

type SourceFormat string

const (
	SourceFormatCSV  SourceFormat = "csv"
	SourceFormatJSON SourceFormat = "json"
	SourceFormatXLSX SourceFormat = "xlsx"
)

func DetectSourceFormat(filename string) (SourceFormat, error) {
	trimmed := strings.TrimSpace(filename)
	if trimmed == "" {
		return "", ErrSourceFilenameEmpty
	}

	switch strings.ToLower(strings.TrimPrefix(filepath.Ext(trimmed), ".")) {
	case string(SourceFormatCSV):
		return SourceFormatCSV, nil
	case string(SourceFormatJSON):
		return SourceFormatJSON, nil
	case string(SourceFormatXLSX):
		return SourceFormatXLSX, nil
	default:
		return "", ErrInvalidSourceFormat
	}
}

func RestoreSourceFormat(raw string) (SourceFormat, error) {
	switch SourceFormat(strings.TrimSpace(raw)) {
	case SourceFormatCSV:
		return SourceFormatCSV, nil
	case SourceFormatJSON:
		return SourceFormatJSON, nil
	case SourceFormatXLSX:
		return SourceFormatXLSX, nil
	default:
		return "", ErrInvalidSourceFormat
	}
}
