package wishlist

import "errors"

var (
	ErrInvalidUserID  = errors.New("invalid user id")
	ErrInvalidGiftID  = errors.New("invalid gift id")
	ErrNoteTooLong    = errors.New("note is too long")
)
