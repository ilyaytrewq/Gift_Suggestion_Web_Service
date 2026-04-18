package domain

import "github.com/google/uuid"

type WishlistID struct {
	value uuid.UUID
}

func NewWishlistID(raw string) (WishlistID, error) {
	if raw == "" {
		return WishlistID{}, ErrWishlistIDEmpty
	}
	if uuid.Validate(raw) != nil {
		return WishlistID{}, ErrInvalidWishlistID
	}

	parsed, err := uuid.Parse(raw)
	if err != nil {
		return WishlistID{}, ErrInvalidWishlistID
	}

	return WishlistID{value: parsed}, nil
}

func (id WishlistID) String() string {
	return id.value.String()
}
