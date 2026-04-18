package parser

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"io"

	catalogimportusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalogimport/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

type csvParser struct{}

func (csvParser) Parse(_ context.Context, source []byte) ([]catalogimportusecase.ImportRowRaw, error) {
	reader := csv.NewReader(bytes.NewReader(source))
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if errors.Is(err, io.EOF) {
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_import_file",
			"csv import file is invalid",
			err,
		)
	}

	index, err := indexHeaders(headers)
	if err != nil {
		return nil, err
	}

	rows := make([]catalogimportusecase.ImportRowRaw, 0)
	rowNumber := 2
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, apperrors.Wrap(
				apperrors.KindValidation,
				"invalid_import_file",
				"csv import file is invalid",
				err,
			)
		}

		rows = append(rows, rowFromColumns(rowNumber, index, record))
		rowNumber++
	}

	return rows, nil
}
