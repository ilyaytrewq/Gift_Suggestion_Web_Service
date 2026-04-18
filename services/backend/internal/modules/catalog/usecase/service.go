package usecase

import (
	"context"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) (*Service, error) {
	if repo == nil {
		return nil, ErrNilCatalogRepository
	}

	return &Service{repo: repo}, nil
}

func (s *Service) ListGifts(ctx context.Context, input ListGiftsInput) (ListGiftsOutput, error) {
	filter, err := normalizeGiftFilter(input)
	if err != nil {
		return ListGiftsOutput{}, err
	}

	gifts, total, err := s.repo.ListGifts(ctx, filter)
	if err != nil {
		return ListGiftsOutput{}, err
	}

	items := make([]Gift, 0, len(gifts))
	for _, gift := range gifts {
		items = append(items, newGift(gift))
	}

	return ListGiftsOutput{
		Items: items,
		Page: Page{
			Limit:  filter.Limit,
			Offset: filter.Offset,
			Total:  total,
		},
	}, nil
}

func (s *Service) GetGift(ctx context.Context, input GetGiftInput) (GetGiftOutput, error) {
	giftID, err := domain.NewGiftID(input.GiftID)
	if err != nil {
		return GetGiftOutput{}, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_gift_id",
			"gift id is invalid",
			err,
		)
	}

	gift, err := s.repo.GetGift(ctx, giftID)
	if err != nil {
		return GetGiftOutput{}, err
	}
	if gift == nil {
		return GetGiftOutput{}, apperrors.New(
			apperrors.KindNotFound,
			"gift_not_found",
			"gift not found",
		)
	}

	return GetGiftOutput{Gift: newGift(*gift)}, nil
}

func (s *Service) ListCategories(ctx context.Context, input ListCategoriesInput) (ListCategoriesOutput, error) {
	filter, err := normalizeCategoryFilter(input)
	if err != nil {
		return ListCategoriesOutput{}, err
	}

	categories, total, err := s.repo.ListCategories(ctx, filter)
	if err != nil {
		return ListCategoriesOutput{}, err
	}

	items := make([]Category, 0, len(categories))
	for _, category := range categories {
		items = append(items, newCategory(category))
	}

	return ListCategoriesOutput{
		Items: items,
		Page: Page{
			Limit:  filter.Limit,
			Offset: filter.Offset,
			Total:  total,
		},
	}, nil
}

func normalizeGiftFilter(input ListGiftsInput) (GiftFilter, error) {
	limit, err := normalizeLimit(input.Limit)
	if err != nil {
		return GiftFilter{}, err
	}
	offset, err := normalizeOffset(input.Offset)
	if err != nil {
		return GiftFilter{}, err
	}
	sort, err := normalizeGiftSort(input.Sort)
	if err != nil {
		return GiftFilter{}, err
	}

	filter := GiftFilter{
		Search:   normalizeSearch(input.Search),
		Limit:    limit,
		Offset:   offset,
		Sort:     sort,
		HasImage: input.HasImage,
	}

	if input.CategoryID != "" {
		categoryID, err := domain.NewCategoryID(input.CategoryID)
		if err != nil {
			return GiftFilter{}, apperrors.Wrap(
				apperrors.KindValidation,
				"invalid_category_id",
				"category id is invalid",
				err,
			)
		}

		filter.CategoryID = &categoryID
	}

	if input.MinPrice != "" {
		minPrice, err := domain.NewPrice(input.MinPrice)
		if err != nil {
			return GiftFilter{}, apperrors.Wrap(
				apperrors.KindValidation,
				"invalid_min_price",
				"min price is invalid",
				err,
			)
		}

		filter.MinPrice = &minPrice
	}

	if input.MaxPrice != "" {
		maxPrice, err := domain.NewPrice(input.MaxPrice)
		if err != nil {
			return GiftFilter{}, apperrors.Wrap(
				apperrors.KindValidation,
				"invalid_max_price",
				"max price is invalid",
				err,
			)
		}

		filter.MaxPrice = &maxPrice
	}

	if filter.MinPrice != nil && filter.MaxPrice != nil && filter.MinPrice.Cents() > filter.MaxPrice.Cents() {
		return GiftFilter{}, apperrors.New(
			apperrors.KindValidation,
			"invalid_price_range",
			"min price must be less than or equal to max price",
		)
	}

	if input.AgeRestriction != nil {
		ageRestriction, err := domain.NewAgeRestriction(*input.AgeRestriction)
		if err != nil {
			return GiftFilter{}, apperrors.Wrap(
				apperrors.KindValidation,
				"invalid_age_restriction",
				"age restriction is invalid",
				err,
			)
		}

		filter.AgeRestriction = &ageRestriction
	}

	return filter, nil
}

func normalizeCategoryFilter(input ListCategoriesInput) (CategoryFilter, error) {
	limit, err := normalizeLimit(input.Limit)
	if err != nil {
		return CategoryFilter{}, err
	}
	offset, err := normalizeOffset(input.Offset)
	if err != nil {
		return CategoryFilter{}, err
	}
	sort, err := normalizeCategorySort(input.Sort)
	if err != nil {
		return CategoryFilter{}, err
	}

	return CategoryFilter{
		Search: normalizeSearch(input.Search),
		Limit:  limit,
		Offset: offset,
		Sort:   sort,
	}, nil
}

func normalizeLimit(limit int) (int, error) {
	if limit == 0 {
		return defaultLimit, nil
	}
	if limit < 1 || limit > maxLimit {
		return 0, apperrors.New(
			apperrors.KindValidation,
			"invalid_limit",
			"limit must be between 1 and 100",
		)
	}

	return limit, nil
}

func normalizeOffset(offset int) (int, error) {
	if offset < 0 {
		return 0, apperrors.New(
			apperrors.KindValidation,
			"invalid_offset",
			"offset must be greater than or equal to zero",
		)
	}

	return offset, nil
}

func normalizeGiftSort(raw string) (GiftSort, error) {
	switch raw {
	case "", string(GiftSortNewest):
		return GiftSortNewest, nil
	case string(GiftSortNameAsc):
		return GiftSortNameAsc, nil
	case string(GiftSortNameDesc):
		return GiftSortNameDesc, nil
	case string(GiftSortPriceAsc):
		return GiftSortPriceAsc, nil
	case string(GiftSortPriceDesc):
		return GiftSortPriceDesc, nil
	default:
		return "", apperrors.New(
			apperrors.KindValidation,
			"invalid_sort",
			"sort is invalid",
		)
	}
}

func normalizeCategorySort(raw string) (CategorySort, error) {
	switch raw {
	case "", string(CategorySortNameAsc):
		return CategorySortNameAsc, nil
	case string(CategorySortNameDesc):
		return CategorySortNameDesc, nil
	case string(CategorySortNewest):
		return CategorySortNewest, nil
	default:
		return "", apperrors.New(
			apperrors.KindValidation,
			"invalid_sort",
			"sort is invalid",
		)
	}
}
