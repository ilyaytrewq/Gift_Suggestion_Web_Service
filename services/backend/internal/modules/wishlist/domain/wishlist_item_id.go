package domain

import "github.com/google/uuid"

type WishlistItemID struct {
	value uuid.UUID
}

func NewWishlistItemID(raw string) (WishlistItemID, error) {
	if raw == "" {
		return WishlistItemID{}, ErrWishlistItemIDEmpty
	}
	if uuid.Validate(raw) != nil {
		return WishlistItemID{}, ErrInvalidWishlistItemID
	}

	parsed, err := uuid.Parse(raw)
	if err != nil {
		return WishlistItemID{}, ErrInvalidWishlistItemID
	}

	return WishlistItemID{value: parsed}, nil
}

func (id WishlistItemID) String() string {
	return id.value.String()
}
