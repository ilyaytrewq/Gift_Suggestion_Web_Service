package domain

import "errors"

var (
	ErrImportJobIDEmpty        = errors.New("import job id is empty")
	ErrInvalidImportJobID      = errors.New("import job id has invalid format")
	ErrImportErrorIDEmpty      = errors.New("import error id is empty")
	ErrInvalidImportErrorID    = errors.New("import error id has invalid format")
	ErrInvalidSourceFormat     = errors.New("source format is invalid")
	ErrSourceFilenameEmpty     = errors.New("source filename is empty")
	ErrNegativeSourceSize      = errors.New("source size is negative")
	ErrImportErrorCodeEmpty    = errors.New("import error code is empty")
	ErrImportErrorMessageEmpty = errors.New("import error message is empty")
)
