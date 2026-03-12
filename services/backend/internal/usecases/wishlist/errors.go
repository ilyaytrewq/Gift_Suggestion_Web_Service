package wishlist

import "errors"

var (
	ErrNilWishlistRepository = errors.New("wishlist repository is nil")
	ErrNilUserRepository     = errors.New("user repository is nil")
	ErrNilGiftRepository     = errors.New("gift repository is nil")

	ErrUserNotFound = errors.New("user not found")
	ErrGiftNotFound = errors.New("gift not found")
)
