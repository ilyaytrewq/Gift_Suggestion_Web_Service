package parser

import (
	"bytes"
	"context"

	"github.com/xuri/excelize/v2"

	catalogimportusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalogimport/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

type xlsxParser struct{}

func (xlsxParser) Parse(_ context.Context, source []byte) ([]catalogimportusecase.ImportRowRaw, error) {
	workbook, err := excelize.OpenReader(bytes.NewReader(source))
	if err != nil {
		return nil, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_import_file",
			"xlsx import file is invalid",
			err,
		)
	}

	sheets := workbook.GetSheetList()
	if len(sheets) == 0 {
		if closeErr := workbook.Close(); closeErr != nil {
			return nil, apperrors.Wrap(
				apperrors.KindValidation,
				"invalid_import_file",
				"xlsx workbook close failed",
				closeErr,
			)
		}

		return nil, errInvalidWorkbook()
	}

	rows, err := workbook.GetRows(sheets[0])
	if err != nil {
		if closeErr := workbook.Close(); closeErr != nil {
			return nil, apperrors.Wrap(
				apperrors.KindValidation,
				"invalid_import_file",
				"xlsx workbook close failed",
				closeErr,
			)
		}

		return nil, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_import_file",
			"xlsx import file is invalid",
			err,
		)
	}
	if closeErr := workbook.Close(); closeErr != nil {
		return nil, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_import_file",
			"xlsx workbook close failed",
			closeErr,
		)
	}
	if len(rows) == 0 {
		return nil, nil
	}

	index, err := indexHeaders(rows[0])
	if err != nil {
		return nil, err
	}

	output := make([]catalogimportusecase.ImportRowRaw, 0, len(rows)-1)
	for rowIndex := 1; rowIndex < len(rows); rowIndex++ {
		output = append(output, rowFromColumns(rowIndex+1, index, rows[rowIndex]))
	}

	return output, nil
}
