package domain

import (
	"strings"
	"time"

	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
)

const maxWishlistNameLength = 120

type Wishlist struct {
	id        WishlistID
	userID    userdomain.UserID
	name      string
	createdAt time.Time
	updatedAt time.Time
}

func NewWishlist(id WishlistID, userID userdomain.UserID, name string, now time.Time) (Wishlist, error) {
	normalizedName, err := normalizeWishlistName(name)
	if err != nil {
		return Wishlist{}, err
	}

	timestamp := now.UTC()

	return Wishlist{
		id:        id,
		userID:    userID,
		name:      normalizedName,
		createdAt: timestamp,
		updatedAt: timestamp,
	}, nil
}

func RestoreWishlist(
	id string,
	userID string,
	name string,
	createdAt time.Time,
	updatedAt time.Time,
) (Wishlist, error) {
	wishlistID, err := NewWishlistID(id)
	if err != nil {
		return Wishlist{}, err
	}

	ownerID, err := userdomain.NewUserID(userID)
	if err != nil {
		return Wishlist{}, err
	}

	normalizedName, err := normalizeWishlistName(name)
	if err != nil {
		return Wishlist{}, err
	}

	return Wishlist{
		id:        wishlistID,
		userID:    ownerID,
		name:      normalizedName,
		createdAt: createdAt.UTC(),
		updatedAt: updatedAt.UTC(),
	}, nil
}

func (w Wishlist) ID() WishlistID {
	return w.id
}

func (w Wishlist) UserID() userdomain.UserID {
	return w.userID
}

func (w Wishlist) Name() string {
	return w.name
}

func (w Wishlist) CreatedAt() time.Time {
	return w.createdAt
}

func (w Wishlist) UpdatedAt() time.Time {
	return w.updatedAt
}

func normalizeWishlistName(raw string) (string, error) {
	normalized := strings.TrimSpace(raw)
	if normalized == "" {
		return "", ErrWishlistNameEmpty
	}
	if len([]rune(normalized)) > maxWishlistNameLength {
		return "", ErrWishlistNameTooLong
	}

	return normalized, nil
}
