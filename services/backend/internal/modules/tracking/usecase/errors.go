package usecase

import "errors"

var (
	ErrNilEventRepository             = errors.New("tracking event repository is nil")
	ErrNilUserReader                  = errors.New("user reader is nil")
	ErrNilGiftReader                  = errors.New("gift reader is nil")
	ErrNilWishlistReader              = errors.New("wishlist reader is nil")
	ErrNilRecommendationRequestReader = errors.New("recommendation request reader is nil")
	ErrNilEventIDGenerator            = errors.New("tracking event id generator is nil")
	ErrNilClock                       = errors.New("clock is nil")
)
