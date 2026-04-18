package domain

import "errors"

var (
	ErrWishlistIDEmpty       = errors.New("wishlist id is empty")
	ErrInvalidWishlistID     = errors.New("wishlist id has invalid format")
	ErrWishlistItemIDEmpty   = errors.New("wishlist item id is empty")
	ErrInvalidWishlistItemID = errors.New("wishlist item id has invalid format")
	ErrWishlistNameEmpty     = errors.New("wishlist name is empty")
	ErrWishlistNameTooLong   = errors.New("wishlist name is too long")
	ErrWishlistNotFound      = errors.New("wishlist not found")
	ErrWishlistAlreadyExists = errors.New("wishlist already exists")
	ErrWishlistItemNotFound  = errors.New("wishlist item not found")
	ErrWishlistItemExists    = errors.New("wishlist item already exists")
)
