package http

import (
	"net/url"
	"strconv"

	wishlistusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/wishlist/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

var allowedListQueryParams = map[string]struct{}{
	"limit":  {},
	"offset": {},
}

type createWishlistRequest struct {
	Name string `json:"name"`
}

type addWishlistItemRequest struct {
	GiftID string `json:"gift_id"`
}

func parseListWishlistsInput(values url.Values) (wishlistusecase.ListWishlistsInput, error) {
	if err := validateQueryParams(values); err != nil {
		return wishlistusecase.ListWishlistsInput{}, err
	}

	limit, err := intParam(values, "limit")
	if err != nil {
		return wishlistusecase.ListWishlistsInput{}, err
	}
	offset, err := intParam(values, "offset")
	if err != nil {
		return wishlistusecase.ListWishlistsInput{}, err
	}

	return wishlistusecase.ListWishlistsInput{
		Limit:  limit,
		Offset: offset,
	}, nil
}

func validateQueryParams(values url.Values) error {
	for key := range values {
		if _, ok := allowedListQueryParams[key]; ok {
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
