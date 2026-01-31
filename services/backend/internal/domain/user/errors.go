package user

import "errors"

var (
	ErrUserIDEmpty   = errors.New("user id is empty")
	ErrInvalidUserID = errors.New("user id has invalid format")

	ErrEmailEmpty   = errors.New("email is empty")
	ErrInvalidEmail = errors.New("email has invalid format")

	ErrPasswordEmpty    = errors.New("password is empty")
	ErrPasswordTooShort = errors.New("password is too short")
	ErrPasswordTooLong  = errors.New("password is too long")

	ErrInvalidRole = errors.New("invalid role value")
	ErrRoleEmpty   = errors.New("role is empty")
)
