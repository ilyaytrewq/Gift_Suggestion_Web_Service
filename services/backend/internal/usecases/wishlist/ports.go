package wishlist

import (
	"context"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/gift"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/user"
)

type WishlistRepository interface {
	Save(ctx context.Context, userID *user.UserID, giftID *gift.GiftID) error
	GetAll(ctx context.Context, userID *user.UserID) ([]*gift.Gift, error)
	Delete(ctx context.Context, userID *user.UserID, giftID *gift.GiftID) error
}

type UserRepository interface {
	FindByID(ctx context.Context, userID *user.UserID) (*user.User, error)
}

type GiftRepository interface {
	FindByID(ctx context.Context, gift *gift.GiftID) (*gift.Gift, error)
}
