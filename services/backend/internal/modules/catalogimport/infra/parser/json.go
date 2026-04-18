package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"

	catalogimportusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalogimport/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

type jsonParser struct{}

func (jsonParser) Parse(_ context.Context, source []byte) ([]catalogimportusecase.ImportRowRaw, error) {
	decoder := json.NewDecoder(bytes.NewReader(source))
	decoder.UseNumber()

	var payload any
	if err := decoder.Decode(&payload); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil
		}

		return nil, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_import_file",
			"json import file is invalid",
			err,
		)
	}

	var items []any
	switch value := payload.(type) {
	case []any:
		items = value
	case map[string]any:
		rawItems, ok := value["items"]
		if !ok {
			return nil, errInvalidJSONShape()
		}

		array, ok := rawItems.([]any)
		if !ok {
			return nil, errInvalidJSONShape()
		}
		items = array
	default:
		return nil, errInvalidJSONShape()
	}

	rows := make([]catalogimportusecase.ImportRowRaw, 0, len(items))
	for index, item := range items {
		record, ok := item.(map[string]any)
		if !ok {
			return nil, errInvalidJSONShape()
		}

		row, err := rowFromMap(index+1, record)
		if err != nil {
			return nil, err
		}

		rows = append(rows, row)
	}

	return rows, nil
}

func scalarToString(value any) (string, error) {
	switch typed := value.(type) {
	case nil:
		return "", nil
	case string:
		return typed, nil
	case json.Number:
		return typed.String(), nil
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(typed), nil
	default:
		return "", fmt.Errorf("unsupported value type %T", value)
	}
}
