package usecase

import (
	"context"
	"time"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	recommendationdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/recommendation/domain"
	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	wishlistdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/wishlist/domain"
)

type CandidateSelection struct {
	BudgetMax            catalogdomain.Price
	RecipientAge         *int
	PreferredCategoryIDs []catalogdomain.CategoryID
	Limit                int
}

type CandidateReader interface {
	SelectCandidates(ctx context.Context, selection CandidateSelection) ([]catalogdomain.Gift, int, error)
}

type RequestRepository interface {
	CreateRequest(ctx context.Context, request *recommendationdomain.RecommendationRequest) error
	CompleteRequest(ctx context.Context, request *recommendationdomain.RecommendationRequest, results []recommendationdomain.RecommendationResult) error
	FailRequest(ctx context.Context, request *recommendationdomain.RecommendationRequest) error
	GetRequest(ctx context.Context, id recommendationdomain.RequestID) (*recommendationdomain.RecommendationRequest, error)
	ListResults(ctx context.Context, requestID recommendationdomain.RequestID) ([]recommendationdomain.RecommendationResult, error)
}

type UserReader interface {
	GetByID(ctx context.Context, id userdomain.UserID) (*userdomain.User, error)
}

type WishlistReader interface {
	GetWishlistByUserID(ctx context.Context, userID userdomain.UserID) (*wishlistdomain.Wishlist, error)
	ListWishlistItems(ctx context.Context, wishlistID wishlistdomain.WishlistID) ([]wishlistdomain.WishlistItem, error)
}

type GiftReader interface {
	GetGift(ctx context.Context, id catalogdomain.GiftID) (*catalogdomain.Gift, error)
}

type RankCandidate struct {
	Gift              catalogdomain.Gift
	AlreadyInWishlist bool
}

type RankInput struct {
	RequestID       string
	UserID          string
	Occasion        string
	Relationship    string
	RecipientAge    *int
	RecipientGender *string
	BudgetMax       catalogdomain.Price
	Interests       []string
	TopN            int
	Candidates      []RankCandidate
}

type RankedItem struct {
	GiftID             catalogdomain.GiftID
	Score              float64
	Explanations       []recommendationdomain.Explanation
	AlternativeGiftIDs []catalogdomain.GiftID
}

type RankOutput struct {
	Items        []RankedItem
	ModelVersion string
}

type RankingGateway interface {
	Rank(ctx context.Context, input RankInput) (RankOutput, error)
}

type RequestIDGenerator interface {
	NewRecommendationRequestID() (recommendationdomain.RequestID, error)
}

type ResultIDGenerator interface {
	NewRecommendationResultID() (recommendationdomain.ResultID, error)
}

type Clock interface {
	Now() time.Time
}
