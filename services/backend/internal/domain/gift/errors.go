package gift

import (
	"errors"
)

var (
	ErrInvalidCategory = errors.New("invalid category")
	ErrInvalidGiftID   = errors.New("invalid gift id")
	ErrInvalidTag      = errors.New("invalid tag")
	ErrInvalidTitle    = errors.New("invalid title")
	ErrInvalidShopURL  = errors.New("invalid shop url")
)
