package parser

import (
	"fmt"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

func errMissingHeader(header string) error {
	return apperrors.New(
		apperrors.KindValidation,
		"invalid_import_file",
		fmt.Sprintf("required header %q is missing", header),
	)
}

func errInvalidJSONShape() error {
	return apperrors.New(
		apperrors.KindValidation,
		"invalid_import_file",
		"json import file must contain an array of objects or an object with items",
	)
}

func errUnsupportedJSONValue(field string) error {
	return apperrors.New(
		apperrors.KindValidation,
		"invalid_import_file",
		fmt.Sprintf("json field %q must be a scalar value", field),
	)
}

func errInvalidWorkbook() error {
	return apperrors.New(
		apperrors.KindValidation,
		"invalid_import_file",
		"xlsx workbook is invalid",
	)
}
