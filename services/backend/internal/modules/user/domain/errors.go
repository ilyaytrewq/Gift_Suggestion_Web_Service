package domain

import "errors"

var (
	ErrUserIDEmpty   = errors.New("user id is empty")
	ErrInvalidUserID = errors.New("user id has invalid format")
	ErrUserNotFound  = errors.New("user not found")
	ErrUserExists    = errors.New("user already exists")

	ErrEmailEmpty   = errors.New("email is empty")
	ErrInvalidEmail = errors.New("email has invalid format")

	ErrInvalidRole = errors.New("invalid role value")
	ErrRoleEmpty   = errors.New("role is empty")

	ErrDisplayNameTooLong = errors.New("display name is too long")
	ErrPasswordHashEmpty  = errors.New("password hash is empty")
)

var (
	ErrPasswordTooShort      = errors.New("password must be at least 8 characters long")
	ErrPasswordTooLong       = errors.New("password must be at most 72 characters long")
	ErrPasswordNoLower       = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoUpper       = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoDigit       = errors.New("password must contain at least one digit")
	ErrPasswordNoSpecial     = errors.New("password must contain at least one special character")
	ErrPasswordContainsSpace = errors.New("password must not contain spaces")
)
