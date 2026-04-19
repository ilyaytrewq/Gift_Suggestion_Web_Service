package usecase

import (
	"context"
	"time"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	recommendationdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/recommendation/domain"
	trackingdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/tracking/domain"
	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	wishlistdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/wishlist/domain"
)

type EventRepository interface {
	CreateEvent(ctx context.Context, event *trackingdomain.Event) (trackingdomain.Event, bool, error)
}

type UserReader interface {
	GetByID(ctx context.Context, id userdomain.UserID) (*userdomain.User, error)
}

type GiftReader interface {
	GetGift(ctx context.Context, id catalogdomain.GiftID) (*catalogdomain.Gift, error)
}

type WishlistReader interface {
	GetWishlistByID(ctx context.Context, id wishlistdomain.WishlistID) (*wishlistdomain.Wishlist, error)
}

type RecommendationRequestReader interface {
	GetRequest(ctx context.Context, id recommendationdomain.RequestID) (*recommendationdomain.RecommendationRequest, error)
}

type EventIDGenerator interface {
	NewTrackingEventID() (trackingdomain.EventID, error)
}

type Clock interface {
	Now() time.Time
}
