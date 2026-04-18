package domain

import (
	"time"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
)

type WishlistItem struct {
	id         WishlistItemID
	wishlistID WishlistID
	giftID     catalogdomain.GiftID
	createdAt  time.Time
}

func NewWishlistItem(
	id WishlistItemID,
	wishlistID WishlistID,
	giftID catalogdomain.GiftID,
	now time.Time,
) WishlistItem {
	return WishlistItem{
		id:         id,
		wishlistID: wishlistID,
		giftID:     giftID,
		createdAt:  now.UTC(),
	}
}

func RestoreWishlistItem(
	id string,
	wishlistID string,
	giftID string,
	createdAt time.Time,
) (WishlistItem, error) {
	itemID, err := NewWishlistItemID(id)
	if err != nil {
		return WishlistItem{}, err
	}

	listID, err := NewWishlistID(wishlistID)
	if err != nil {
		return WishlistItem{}, err
	}

	catalogGiftID, err := catalogdomain.NewGiftID(giftID)
	if err != nil {
		return WishlistItem{}, err
	}

	return WishlistItem{
		id:         itemID,
		wishlistID: listID,
		giftID:     catalogGiftID,
		createdAt:  createdAt.UTC(),
	}, nil
}

func (i WishlistItem) ID() WishlistItemID {
	return i.id
}

func (i WishlistItem) WishlistID() WishlistID {
	return i.wishlistID
}

func (i WishlistItem) GiftID() catalogdomain.GiftID {
	return i.giftID
}

func (i WishlistItem) CreatedAt() time.Time {
	return i.createdAt
}
