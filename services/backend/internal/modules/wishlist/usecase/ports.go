package usecase

import (
	"context"
	"time"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	wishlistdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/wishlist/domain"
)

type WishlistSummaryRecord struct {
	Wishlist  wishlistdomain.Wishlist
	ItemCount int
}

type WishlistRepository interface {
	CreateWishlist(ctx context.Context, wishlist *wishlistdomain.Wishlist) error
	GetWishlistByID(ctx context.Context, id wishlistdomain.WishlistID) (*wishlistdomain.Wishlist, error)
	GetWishlistByUserID(ctx context.Context, userID userdomain.UserID) (*wishlistdomain.Wishlist, error)
	ListWishlistsByUser(
		ctx context.Context,
		userID userdomain.UserID,
		limit int,
		offset int,
	) ([]WishlistSummaryRecord, int, error)
	ListWishlistItems(ctx context.Context, wishlistID wishlistdomain.WishlistID) ([]wishlistdomain.WishlistItem, error)
	GetWishlistItemByGiftID(
		ctx context.Context,
		wishlistID wishlistdomain.WishlistID,
		giftID catalogdomain.GiftID,
	) (*wishlistdomain.WishlistItem, error)
	AddWishlistItem(ctx context.Context, item *wishlistdomain.WishlistItem) error
	RemoveWishlistItem(ctx context.Context, wishlistID wishlistdomain.WishlistID, giftID catalogdomain.GiftID) error
	DeleteWishlist(ctx context.Context, id wishlistdomain.WishlistID) error
}

type UserReader interface {
	GetByID(ctx context.Context, id userdomain.UserID) (*userdomain.User, error)
}

type GiftReader interface {
	GetGift(ctx context.Context, id catalogdomain.GiftID) (*catalogdomain.Gift, error)
}

type WishlistIDGenerator interface {
	NewWishlistID() (wishlistdomain.WishlistID, error)
}

type WishlistItemIDGenerator interface {
	NewWishlistItemID() (wishlistdomain.WishlistItemID, error)
}

type Clock interface {
	Now() time.Time
}
