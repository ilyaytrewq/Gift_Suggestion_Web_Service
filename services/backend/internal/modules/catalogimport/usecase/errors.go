package usecase

import "errors"

var (
	ErrNilJobRepository          = errors.New("import job repository is nil")
	ErrNilParserRegistry         = errors.New("parser registry is nil")
	ErrNilGiftIDGenerator        = errors.New("gift id generator is nil")
	ErrNilImportJobIDGenerator   = errors.New("import job id generator is nil")
	ErrNilImportErrorIDGenerator = errors.New("import error id generator is nil")
	ErrNilClock                  = errors.New("clock is nil")
)
