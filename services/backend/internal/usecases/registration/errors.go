package registration

import "errors"

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrNilUserRepository  = errors.New("user repository is nil")
	ErrNilIDGenerator     = errors.New("id generator is nil")
)
