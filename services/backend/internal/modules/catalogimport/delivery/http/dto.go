package http

import (
	"net/url"
	"strconv"

	catalogimportusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalogimport/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

var allowedErrorsQueryParams = map[string]struct{}{
	"limit":  {},
	"offset": {},
}

func parseListImportErrorsInput(values url.Values) (catalogimportusecase.ListImportErrorsInput, error) {
	for key := range values {
		if _, ok := allowedErrorsQueryParams[key]; ok {
			continue
		}

		return catalogimportusecase.ListImportErrorsInput{}, apperrors.New(
			apperrors.KindValidation,
			"invalid_query_parameter",
			"query parameter is not supported",
		)
	}

	limit, err := parseIntQuery(values, "limit")
	if err != nil {
		return catalogimportusecase.ListImportErrorsInput{}, err
	}
	offset, err := parseIntQuery(values, "offset")
	if err != nil {
		return catalogimportusecase.ListImportErrorsInput{}, err
	}

	return catalogimportusecase.ListImportErrorsInput{
		Limit:  limit,
		Offset: offset,
	}, nil
}

func parseIntQuery(values url.Values, key string) (int, error) {
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
