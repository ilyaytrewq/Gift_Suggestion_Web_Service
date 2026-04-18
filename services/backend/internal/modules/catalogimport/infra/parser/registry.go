package parser

import (
	"errors"

	catalogimportdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalogimport/domain"
	catalogimportusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalogimport/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

type Registry struct {
	parsers map[catalogimportdomain.SourceFormat]catalogimportusecase.Parser
}

func NewRegistry() *Registry {
	return &Registry{
		parsers: map[catalogimportdomain.SourceFormat]catalogimportusecase.Parser{
			catalogimportdomain.SourceFormatCSV:  csvParser{},
			catalogimportdomain.SourceFormatJSON: jsonParser{},
			catalogimportdomain.SourceFormatXLSX: xlsxParser{},
		},
	}
}

func (r *Registry) ParserFor(format catalogimportdomain.SourceFormat) (catalogimportusecase.Parser, error) {
	if r == nil {
		return nil, errors.New("parser registry is nil")
	}

	parser, ok := r.parsers[format]
	if !ok {
		return nil, apperrors.New(
			apperrors.KindValidation,
			"unsupported_import_format",
			"import format is not supported",
		)
	}

	return parser, nil
}
