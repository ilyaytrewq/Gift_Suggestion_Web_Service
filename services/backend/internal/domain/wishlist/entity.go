package wishlist

import (
	"time"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/models/gift"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/models/user"
)

type WishlistItem struct {
	UserID    user.UserID
	GiftID    gift.GiftID
	Note      string
	CreatedAt time.Time
}

func NewWishlistItem(id user.UserID, giftID gift.GiftID, note string) (*WishlistItem, error) {
	if !id.IsValid() {
		return nil, ErrInvalidUserID
	}
	if !giftID.IsValid() {
		return nil, ErrInvalidGiftID
	}
	if !isValidNote(note) {
		return nil, ErrNoteTooLong
	}

	return &WishlistItem{
		UserID:    id,
		GiftID:    giftID,
		Note:      note,
		CreatedAt: time.Now(),
	}, nil
}
