package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

const (
	testCatalogGiftID     = "550e8400-e29b-41d4-a716-446655440010"
	testCatalogCategoryID = "550e8400-e29b-41d4-a716-446655440011"
)

func TestServiceListGiftsRejectsInvalidPriceRange(t *testing.T) {
	t.Parallel()

	service, err := NewService(fakeRepository{})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.ListGifts(context.Background(), ListGiftsInput{
		MinPrice: "200.00",
		MaxPrice: "100.00",
	})
	if err == nil {
		t.Fatal("ListGifts() expected validation error")
	}

	appErr := apperrors.From(err)
	if appErr.Code() != "invalid_price_range" {
		t.Fatalf("ListGifts() error code = %q, want %q", appErr.Code(), "invalid_price_range")
	}
}

func TestServiceGetGiftReturnsNotFound(t *testing.T) {
	t.Parallel()

	service, err := NewService(fakeRepository{})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.GetGift(context.Background(), GetGiftInput{GiftID: testCatalogGiftID})
	if err == nil {
		t.Fatal("GetGift() expected not found error")
	}

	appErr := apperrors.From(err)
	if appErr.Kind() != apperrors.KindNotFound {
		t.Fatalf("GetGift() error kind = %q, want %q", appErr.Kind(), apperrors.KindNotFound)
	}
}

func TestServiceListCategoriesSuccess(t *testing.T) {
	t.Parallel()

	repo := fakeRepository{
		categories: []domain.Category{
			mustCategory(t, testCatalogCategoryID, "Books"),
		},
		categoryTotal: 1,
	}

	service, err := NewService(repo)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	output, err := service.ListCategories(context.Background(), ListCategoriesInput{})
	if err != nil {
		t.Fatalf("ListCategories() error = %v", err)
	}

	if len(output.Items) != 1 {
		t.Fatalf("expected one category, got %d", len(output.Items))
	}
	if output.Items[0].Name != "Books" {
		t.Fatalf("ListCategories() first item name = %q, want %q", output.Items[0].Name, "Books")
	}
	if output.Page.Total != 1 {
		t.Fatalf("ListCategories() page total = %d, want %d", output.Page.Total, 1)
	}
}

type fakeRepository struct {
	gifts         []domain.Gift
	gift          *domain.Gift
	giftTotal     int
	categories    []domain.Category
	categoryTotal int
}

func (r fakeRepository) ListGifts(context.Context, GiftFilter) ([]domain.Gift, int, error) {
	return r.gifts, r.giftTotal, nil
}

func (r fakeRepository) GetGift(context.Context, domain.GiftID) (*domain.Gift, error) {
	return r.gift, nil
}

func (r fakeRepository) ListCategories(context.Context, CategoryFilter) ([]domain.Category, int, error) {
	return r.categories, r.categoryTotal, nil
}

func (r fakeRepository) ListOffersByGiftID(context.Context, domain.GiftID) ([]domain.Offer, error) {
	return nil, nil
}

func (r fakeRepository) ListSimilarGifts(context.Context, domain.GiftID, *domain.CategoryID, int64, int) ([]domain.Gift, error) {
	return nil, nil
}

func mustCategory(t *testing.T, id, name string) domain.Category {
	t.Helper()

	category, err := domain.RestoreCategory(id, name, time.Date(2026, 4, 18, 10, 0, 0, 0, time.UTC), time.Date(2026, 4, 18, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("RestoreCategory() error = %v", err)
	}

	return category
}
