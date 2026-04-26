package usecase

import (
	"time"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	wishlistdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/wishlist/domain"
)

type CreateWishlistInput struct {
	UserID string
	Name   string
}

type ListWishlistsInput struct {
	UserID string
	Limit  int
	Offset int
}

type GetWishlistInput struct {
	UserID     string
	WishlistID string
}

type AddWishlistItemInput struct {
	UserID     string
	WishlistID string
	GiftID     string
}

type RemoveWishlistItemInput struct {
	UserID     string
	WishlistID string
	GiftID     string
}

type DeleteWishlistInput struct {
	UserID     string
	WishlistID string
}

type Page struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

type GiftPreview struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	Price          string        `json:"price"`
	StoreLink      string        `json:"store_link"`
	Image          *string       `json:"image,omitempty"`
	AgeRestriction *int          `json:"age_restriction,omitempty"`
	Category       *GiftCategory `json:"category,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

type GiftCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type WishlistItem struct {
	ID        string      `json:"id"`
	CreatedAt time.Time   `json:"created_at"`
	Gift      GiftPreview `json:"gift"`
}

type WishlistSummary struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ItemCount int       `json:"item_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Wishlist struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	ItemCount int            `json:"item_count"`
	Items     []WishlistItem `json:"items"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type CreateWishlistOutput struct {
	Wishlist Wishlist `json:"wishlist"`
}

type ListWishlistsOutput struct {
	Items []WishlistSummary `json:"items"`
	Page  Page              `json:"page"`
}

type GetWishlistOutput struct {
	Wishlist Wishlist `json:"wishlist"`
}

type AddWishlistItemOutput struct {
	AlreadyInWishlist bool         `json:"already_in_wishlist"`
	Item              WishlistItem `json:"item"`
}

type RemoveWishlistItemOutput struct {
	Removed bool `json:"removed"`
}

type DeleteWishlistOutput struct {
	Deleted bool `json:"deleted"`
}

func newWishlistSummary(record WishlistSummaryRecord) WishlistSummary {
	wishlist := record.Wishlist

	return WishlistSummary{
		ID:        wishlist.ID().String(),
		Name:      wishlist.Name(),
		ItemCount: record.ItemCount,
		CreatedAt: wishlist.CreatedAt(),
		UpdatedAt: wishlist.UpdatedAt(),
	}
}

func newWishlist(wishlist wishlistdomain.Wishlist, items []WishlistItem) Wishlist {
	return Wishlist{
		ID:        wishlist.ID().String(),
		Name:      wishlist.Name(),
		ItemCount: len(items),
		Items:     items,
		CreatedAt: wishlist.CreatedAt(),
		UpdatedAt: wishlist.UpdatedAt(),
	}
}

func newWishlistItem(item wishlistdomain.WishlistItem, gift catalogdomain.Gift) WishlistItem {
	return WishlistItem{
		ID:        item.ID().String(),
		CreatedAt: item.CreatedAt(),
		Gift:      newGiftPreview(gift),
	}
}

func newGiftPreview(gift catalogdomain.Gift) GiftPreview {
	var ageRestriction *int
	if gift.AgeRestriction() != nil {
		value := gift.AgeRestriction().Int()
		ageRestriction = &value
	}

	var category *GiftCategory
	if gift.CategoryID() != nil {
		category = &GiftCategory{
			ID:   gift.CategoryID().String(),
			Name: gift.CategoryName(),
		}
	}

	return GiftPreview{
		ID:             gift.ID().String(),
		Name:           gift.Name(),
		Description:    gift.Description(),
		Price:          gift.Price().DecimalString(),
		StoreLink:      gift.StoreLink(),
		Image:          gift.Image(),
		AgeRestriction: ageRestriction,
		Category:       category,
		CreatedAt:      gift.CreatedAt(),
		UpdatedAt:      gift.UpdatedAt(),
	}
}
