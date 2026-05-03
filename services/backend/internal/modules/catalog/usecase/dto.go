package usecase

import (
	"strings"
	"time"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
)

type ListGiftsInput struct {
	Search         string
	CategoryID     string
	MinPrice       string
	MaxPrice       string
	AgeRestriction *int
	HasImage       *bool
	Limit          int
	Offset         int
	Sort           string
}

type GetGiftInput struct {
	GiftID string
}

type GetSimilarGiftsInput struct {
	GiftID string
	Limit  int
}

type GetSimilarGiftsOutput struct {
	Items []Gift `json:"items"`
}

type ListCategoriesInput struct {
	Search   string
	HasGifts *bool
	Limit    int
	Offset   int
	Sort     string
}

type Page struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

type Category struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GiftCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Offer struct {
	ID        string `json:"id"`
	StoreName string `json:"store_name"`
	StoreURL  string `json:"store_url"`
	Price     string `json:"price"`
	Currency  string `json:"currency"`
	Available bool   `json:"available"`
}

type Gift struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	Price          string        `json:"price"`
	StoreLink      string        `json:"store_link"`
	Image          *string       `json:"image,omitempty"`
	AgeRestriction *int          `json:"age_restriction,omitempty"`
	Category       *GiftCategory `json:"category,omitempty"`
	Offers         []Offer       `json:"offers,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

type ListGiftsOutput struct {
	Items []Gift `json:"items"`
	Page  Page   `json:"page"`
}

type GetGiftOutput struct {
	Gift Gift `json:"gift"`
}

type ListCategoriesOutput struct {
	Items []Category `json:"items"`
	Page  Page       `json:"page"`
}

func newOffer(o domain.Offer) Offer {
	return Offer{
		ID:        o.ID(),
		StoreName: o.StoreName(),
		StoreURL:  o.StoreURL(),
		Price:     o.Price().DecimalString(),
		Currency:  o.Currency(),
		Available: o.Available(),
	}
}

func newGift(gift domain.Gift) Gift {
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

	return Gift{
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

func newCategory(category domain.Category) Category {
	return Category{
		ID:        category.ID().String(),
		Name:      category.Name(),
		CreatedAt: category.CreatedAt(),
		UpdatedAt: category.UpdatedAt(),
	}
}

func normalizeSearch(raw string) string {
	return strings.TrimSpace(raw)
}
