package usecase

import (
	"context"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
)

type GiftSort string

const (
	GiftSortNewest    GiftSort = "newest"
	GiftSortNameAsc   GiftSort = "name_asc"
	GiftSortNameDesc  GiftSort = "name_desc"
	GiftSortPriceAsc  GiftSort = "price_asc"
	GiftSortPriceDesc GiftSort = "price_desc"
)

type CategorySort string

const (
	CategorySortNameAsc  CategorySort = "name_asc"
	CategorySortNameDesc CategorySort = "name_desc"
	CategorySortNewest   CategorySort = "newest"
)

type GiftFilter struct {
	Search         string
	CategoryID     *domain.CategoryID
	MinPrice       *domain.Price
	MaxPrice       *domain.Price
	AgeRestriction *domain.AgeRestriction
	HasImage       *bool
	Limit          int
	Offset         int
	Sort           GiftSort
}

type CategoryFilter struct {
	Search    string
	HasGifts  *bool
	Limit     int
	Offset    int
	Sort      CategorySort
}

type Repository interface {
	ListGifts(ctx context.Context, filter GiftFilter) ([]domain.Gift, int, error)
	GetGift(ctx context.Context, id domain.GiftID) (*domain.Gift, error)
	ListCategories(ctx context.Context, filter CategoryFilter) ([]domain.Category, int, error)
	ListOffersByGiftID(ctx context.Context, giftID domain.GiftID) ([]domain.Offer, error)
	ListSimilarGifts(ctx context.Context, giftID domain.GiftID, categoryID *domain.CategoryID, priceCents int64, limit int) ([]domain.Gift, error)
}
