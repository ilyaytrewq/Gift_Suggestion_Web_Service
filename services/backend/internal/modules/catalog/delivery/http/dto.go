package http

import (
	"net/url"
	"strconv"

	catalogusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

var allowedGiftQueryParams = map[string]struct{}{
	"q":               {},
	"category_id":     {},
	"min_price":       {},
	"max_price":       {},
	"age_restriction": {},
	"has_image":       {},
	"limit":           {},
	"offset":          {},
	"sort":            {},
}

var allowedCategoryQueryParams = map[string]struct{}{
	"q":         {},
	"limit":     {},
	"offset":    {},
	"sort":      {},
	"has_gifts": {},
}

func parseListGiftsInput(values url.Values) (catalogusecase.ListGiftsInput, error) {
	if err := validateQueryParams(values, allowedGiftQueryParams); err != nil {
		return catalogusecase.ListGiftsInput{}, err
	}

	limit, err := intParam(values, "limit")
	if err != nil {
		return catalogusecase.ListGiftsInput{}, err
	}
	offset, err := intParam(values, "offset")
	if err != nil {
		return catalogusecase.ListGiftsInput{}, err
	}
	ageRestriction, err := optionalIntParam(values, "age_restriction")
	if err != nil {
		return catalogusecase.ListGiftsInput{}, err
	}
	hasImage, err := optionalBoolParam(values, "has_image")
	if err != nil {
		return catalogusecase.ListGiftsInput{}, err
	}

	return catalogusecase.ListGiftsInput{
		Search:         values.Get("q"),
		CategoryID:     values.Get("category_id"),
		MinPrice:       values.Get("min_price"),
		MaxPrice:       values.Get("max_price"),
		AgeRestriction: ageRestriction,
		HasImage:       hasImage,
		Limit:          limit,
		Offset:         offset,
		Sort:           values.Get("sort"),
	}, nil
}

func parseListCategoriesInput(values url.Values) (catalogusecase.ListCategoriesInput, error) {
	if err := validateQueryParams(values, allowedCategoryQueryParams); err != nil {
		return catalogusecase.ListCategoriesInput{}, err
	}

	limit, err := intParam(values, "limit")
	if err != nil {
		return catalogusecase.ListCategoriesInput{}, err
	}
	offset, err := intParam(values, "offset")
	if err != nil {
		return catalogusecase.ListCategoriesInput{}, err
	}
	hasGifts, err := optionalBoolParam(values, "has_gifts")
	if err != nil {
		return catalogusecase.ListCategoriesInput{}, err
	}

	return catalogusecase.ListCategoriesInput{
		Search:   values.Get("q"),
		HasGifts: hasGifts,
		Limit:    limit,
		Offset:   offset,
		Sort:     values.Get("sort"),
	}, nil
}

func validateQueryParams(values url.Values, allowed map[string]struct{}) error {
	for key := range values {
		if _, ok := allowed[key]; ok {
			continue
		}

		return apperrors.New(
			apperrors.KindValidation,
			"invalid_query_parameter",
			"query parameter is not supported",
		)
	}

	return nil
}

func intParam(values url.Values, key string) (int, error) {
	raw := values.Get(key)
	if raw == "" {
		return 0, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, apperrors.New(
			apperrors.KindValidation,
			"invalid_query_parameter",
			"query parameter must be an integer",
		)
	}

	return value, nil
}

func optionalIntParam(values url.Values, key string) (*int, error) {
	raw := values.Get(key)
	if raw == "" {
		return nil, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return nil, apperrors.New(
			apperrors.KindValidation,
			"invalid_query_parameter",
			"query parameter must be an integer",
		)
	}

	return &value, nil
}

func optionalBoolParam(values url.Values, key string) (*bool, error) {
	raw := values.Get(key)
	if raw == "" {
		return nil, nil
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, apperrors.New(
			apperrors.KindValidation,
			"invalid_query_parameter",
			"query parameter must be a boolean",
		)
	}

	return &value, nil
}
