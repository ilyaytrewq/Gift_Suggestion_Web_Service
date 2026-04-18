package usecase

import "errors"

var (
	ErrNilWishlistRepository      = errors.New("wishlist repository is nil")
	ErrNilUserReader              = errors.New("user reader is nil")
	ErrNilGiftReader              = errors.New("gift reader is nil")
	ErrNilWishlistIDGenerator     = errors.New("wishlist id generator is nil")
	ErrNilWishlistItemIDGenerator = errors.New("wishlist item id generator is nil")
	ErrNilClock                   = errors.New("clock is nil")
)
