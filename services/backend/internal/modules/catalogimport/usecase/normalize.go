package usecase

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
)

type normalizedRecord struct {
	Gift                catalogdomain.Gift
	NormalizedName      string
	NormalizedStoreLink string
	SourceName          *string
}

type rowError struct {
	Field   string
	Code    string
	Message string
}

func (s *Service) normalizeRow(
	ctx context.Context,
	row ImportRowRaw,
	defaultSourceLabel string,
) (normalizedRecord, *string, *rowError, error) {
	if strings.TrimSpace(row.Name) == "" {
		return normalizedRecord{}, nil, &rowError{
			Field:   "name",
			Code:    "missing_required_field",
			Message: "name is required",
		}, nil
	}
	categoryName := strings.TrimSpace(row.Category)
	if categoryName == "" {
		return normalizedRecord{}, nil, &rowError{
			Field:   "category",
			Code:    "missing_required_field",
			Message: "category is required",
		}, nil
	}
	if strings.TrimSpace(row.Description) == "" {
		return normalizedRecord{}, nil, &rowError{
			Field:   "description",
			Code:    "missing_required_field",
			Message: "description is required",
		}, nil
	}
	if strings.TrimSpace(row.PriceRaw) == "" {
		return normalizedRecord{}, nil, &rowError{
			Field:   "price",
			Code:    "missing_required_field",
			Message: "price is required",
		}, nil
	}
	if strings.TrimSpace(row.StoreLink) == "" {
		return normalizedRecord{}, nil, &rowError{
			Field:   "store_link",
			Code:    "missing_required_field",
			Message: "store link is required",
		}, nil
	}

	category, err := s.repo.FindCategoryByName(ctx, categoryName)
	if err != nil {
		return normalizedRecord{}, nil, nil, err
	}
	if category == nil {
		return normalizedRecord{}, nil, &rowError{
			Field:   "category",
			Code:    "unknown_category",
			Message: "category is not known to catalog",
		}, nil
	}

	giftID, err := s.giftIDGenerator.NewGiftID()
	if err != nil {
		return normalizedRecord{}, nil, nil, err
	}

	categoryID := category.ID().String()
	categoryDisplayName := category.Name()
	ageRestriction, rowErr := normalizeAgeRestriction(row.AgeRestrictionRaw)
	if rowErr != nil {
		return normalizedRecord{}, nil, rowErr, nil
	}
	image, rowErr := normalizeImageLink(row.Image)
	if rowErr != nil {
		return normalizedRecord{}, nil, rowErr, nil
	}

	gift, err := catalogdomain.RestoreGift(
		giftID.String(),
		&categoryID,
		&categoryDisplayName,
		strings.TrimSpace(row.Name),
		strings.TrimSpace(row.Description),
		strings.TrimSpace(row.PriceRaw),
		strings.TrimSpace(row.StoreLink),
		image,
		ageRestriction,
		s.clock.Now(),
		s.clock.Now(),
	)
	if err != nil {
		mapped := mapCatalogValidationError(err)
		return normalizedRecord{}, nil, &mapped, nil
	}

	recordKey := strings.ToLower(gift.Name()) + "|" + strings.ToLower(gift.StoreLink())

	sourceName := normalizeOptionalString(row.SourceName)
	if sourceName == nil {
		sourceName = normalizeOptionalString(defaultSourceLabel)
	}

	return normalizedRecord{
		Gift:                gift,
		NormalizedName:      strings.ToLower(gift.Name()),
		NormalizedStoreLink: strings.ToLower(gift.StoreLink()),
		SourceName:          sourceName,
	}, &recordKey, nil, nil
}

func normalizeAgeRestriction(raw string) (*int, *rowError) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	trimmed = strings.TrimSuffix(trimmed, "+")
	switch strings.ToLower(trimmed) {
	case "all", "none":
		value := 0
		return &value, nil
	}

	value, err := strconv.Atoi(trimmed)
	if err != nil {
		return nil, &rowError{
			Field:   "age_restriction",
			Code:    "invalid_age_restriction",
			Message: "age restriction is invalid",
		}
	}

	return &value, nil
}

func normalizeOptionalString(raw string) *string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func normalizeImageLink(raw string) (*string, *rowError) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	parsed, err := url.ParseRequestURI(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, &rowError{
			Field:   "image",
			Code:    "invalid_image_link",
			Message: "image link is invalid",
		}
	}

	return &trimmed, nil
}
