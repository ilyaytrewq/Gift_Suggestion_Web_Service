package domain

import "errors"

var (
	ErrGiftIDEmpty           = errors.New("gift id is empty")
	ErrInvalidGiftID         = errors.New("gift id has invalid format")
	ErrCategoryIDEmpty       = errors.New("category id is empty")
	ErrInvalidCategoryID     = errors.New("category id has invalid format")
	ErrGiftNameEmpty         = errors.New("gift name is empty")
	ErrCategoryNameEmpty     = errors.New("category name is empty")
	ErrPriceEmpty            = errors.New("price is empty")
	ErrInvalidPrice          = errors.New("price has invalid format")
	ErrNegativePrice         = errors.New("price must be non-negative")
	ErrStoreLinkEmpty        = errors.New("store link is empty")
	ErrInvalidStoreLink      = errors.New("store link has invalid format")
	ErrInvalidAgeRestriction = errors.New("age restriction is invalid")
	ErrOfferStoreNameEmpty   = errors.New("offer store name is empty")
	ErrOfferCurrencyEmpty    = errors.New("offer currency is empty")
)
