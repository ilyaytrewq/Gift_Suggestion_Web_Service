package usecase

import (
	"context"
	"time"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	catalogimportdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalogimport/domain"
)

type Parser interface {
	Parse(ctx context.Context, source []byte) ([]ImportRowRaw, error)
}

type ParserRegistry interface {
	ParserFor(format catalogimportdomain.SourceFormat) (Parser, error)
}

type Repository interface {
	CreateJob(ctx context.Context, job *catalogimportdomain.ImportJob) error
	UpdateJob(ctx context.Context, job *catalogimportdomain.ImportJob) error
	SaveErrors(ctx context.Context, errors []catalogimportdomain.ImportError) error
	FindCategoryByName(ctx context.Context, name string) (*catalogdomain.Category, error)
	GiftExists(ctx context.Context, normalizedName, normalizedStoreLink string) (bool, error)
	InsertGift(ctx context.Context, gift catalogdomain.Gift, sourceName *string) error
	InsertOffers(ctx context.Context, offers []catalogdomain.Offer) error
	GetJob(ctx context.Context, id catalogimportdomain.ImportJobID) (*catalogimportdomain.ImportJob, error)
	ListErrors(
		ctx context.Context,
		jobID catalogimportdomain.ImportJobID,
		limit int,
		offset int,
	) ([]catalogimportdomain.ImportError, int, error)
}

type GiftIDGenerator interface {
	NewGiftID() (catalogdomain.GiftID, error)
}

type ImportJobIDGenerator interface {
	NewImportJobID() (catalogimportdomain.ImportJobID, error)
}

type ImportErrorIDGenerator interface {
	NewImportErrorID() (catalogimportdomain.ImportErrorID, error)
}

type Clock interface {
	Now() time.Time
}
