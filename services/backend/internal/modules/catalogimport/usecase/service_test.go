package usecase

import (
	"context"
	"strings"
	"testing"
	"time"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	catalogimportdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalogimport/domain"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

const (
	testImportUserID   = "550e8400-e29b-41d4-a716-446655440900"
	testImportJobID    = "550e8400-e29b-41d4-a716-446655440901"
	testImportErrorID  = "550e8400-e29b-41d4-a716-446655440902"
	testImportGiftID   = "550e8400-e29b-41d4-a716-446655440903"
	testImportCategory = "Books"
)

func TestServiceRunImportSuccess(t *testing.T) {
	t.Parallel()

	service := mustImportService(t, importServiceDeps{
		repo: &fakeImportRepository{
			categories: map[string]*catalogdomain.Category{
				"books": mustCategory(t, testImportCategory),
			},
		},
		parsers: fakeParserRegistry{
			parser: fakeParser{
				rows: []ImportRowRaw{
					{
						RowNumber:   1,
						Name:        "Book",
						Category:    testImportCategory,
						PriceRaw:    "1499.90",
						Description: "Gift book",
						StoreLink:   "https://example.com/book",
					},
				},
			},
		},
	})

	output, err := service.RunImport(context.Background(), RunImportInput{
		RequestedByUserID: testImportUserID,
		Filename:          "catalog.csv",
		FileSizeBytes:     42,
		File:              []byte("csv"),
	})
	if err != nil {
		t.Fatalf("RunImport() error = %v", err)
	}

	if output.Job.Status != string(catalogimportdomain.StatusCompleted) {
		t.Fatalf("RunImport() status = %q, want %q", output.Job.Status, catalogimportdomain.StatusCompleted)
	}
	if output.Job.Summary.ImportedRows != 1 {
		t.Fatalf("RunImport() imported rows = %d, want 1", output.Job.Summary.ImportedRows)
	}
}

func TestServiceRunImportRejectsLargeFile(t *testing.T) {
	t.Parallel()

	service := mustImportService(t, importServiceDeps{
		repo:             &fakeImportRepository{},
		parsers:          fakeParserRegistry{parser: fakeParser{}},
		maxFileSizeBytes: 4,
	})

	_, err := service.RunImport(context.Background(), RunImportInput{
		RequestedByUserID: testImportUserID,
		Filename:          "catalog.csv",
		FileSizeBytes:     5,
		File:              []byte("12345"),
	})
	if err == nil {
		t.Fatal("RunImport() expected file size error")
	}

	appErr := apperrors.From(err)
	if appErr.Code() != "file_too_large" {
		t.Fatalf("RunImport() code = %q, want %q", appErr.Code(), "file_too_large")
	}
}

func TestServiceRunImportUnsupportedFormat(t *testing.T) {
	t.Parallel()

	service := mustImportService(t, importServiceDeps{
		repo:    &fakeImportRepository{},
		parsers: fakeParserRegistry{parser: fakeParser{}},
	})

	_, err := service.RunImport(context.Background(), RunImportInput{
		RequestedByUserID: testImportUserID,
		Filename:          "catalog.txt",
		FileSizeBytes:     4,
		File:              []byte("data"),
	})
	if err == nil {
		t.Fatal("RunImport() expected unsupported format error")
	}

	appErr := apperrors.From(err)
	if appErr.Code() != "unsupported_import_format" {
		t.Fatalf("RunImport() code = %q, want %q", appErr.Code(), "unsupported_import_format")
	}
}

func TestServiceRunImportEmptyFile(t *testing.T) {
	t.Parallel()

	service := mustImportService(t, importServiceDeps{
		repo:    &fakeImportRepository{},
		parsers: fakeParserRegistry{parser: fakeParser{}},
	})

	_, err := service.RunImport(context.Background(), RunImportInput{
		RequestedByUserID: testImportUserID,
		Filename:          "catalog.csv",
		FileSizeBytes:     0,
		File:              nil,
	})
	if err == nil {
		t.Fatal("RunImport() expected empty file error")
	}

	appErr := apperrors.From(err)
	if appErr.Code() != "empty_import_file" {
		t.Fatalf("RunImport() code = %q, want %q", appErr.Code(), "empty_import_file")
	}
}

func TestServiceRunImportPartialValidFile(t *testing.T) {
	t.Parallel()

	repo := &fakeImportRepository{
		categories: map[string]*catalogdomain.Category{
			"books": mustCategory(t, testImportCategory),
		},
		existingKeys: map[string]bool{
			"book existing|https://example.com/existing": true,
		},
	}
	service := mustImportService(t, importServiceDeps{
		repo: repo,
		parsers: fakeParserRegistry{
			parser: fakeParser{
				rows: []ImportRowRaw{
					{
						RowNumber:   1,
						Name:        "Book Fresh",
						Category:    testImportCategory,
						PriceRaw:    "1499.90",
						Description: "Gift book",
						StoreLink:   "https://example.com/fresh",
					},
					{
						RowNumber:   2,
						Name:        "Broken Price",
						Category:    testImportCategory,
						PriceRaw:    "bad",
						Description: "Gift book",
						StoreLink:   "https://example.com/bad",
					},
					{
						RowNumber:   3,
						Name:        "Book Existing",
						Category:    testImportCategory,
						PriceRaw:    "100.00",
						Description: "Duplicate in catalog",
						StoreLink:   "https://example.com/existing",
					},
					{
						RowNumber:   4,
						Name:        "Unknown Category",
						Category:    "Unknown",
						PriceRaw:    "100.00",
						Description: "Unknown",
						StoreLink:   "https://example.com/unknown",
					},
					{
						RowNumber:   5,
						Name:        "Book Fresh",
						Category:    testImportCategory,
						PriceRaw:    "1499.90",
						Description: "Duplicate in file",
						StoreLink:   "https://example.com/fresh",
					},
				},
			},
		},
	})

	output, err := service.RunImport(context.Background(), RunImportInput{
		RequestedByUserID: testImportUserID,
		Filename:          "catalog.csv",
		FileSizeBytes:     100,
		File:              []byte("csv"),
	})
	if err != nil {
		t.Fatalf("RunImport() error = %v", err)
	}

	if output.Job.Status != string(catalogimportdomain.StatusCompletedWithErrors) {
		t.Fatalf("RunImport() status = %q, want %q", output.Job.Status, catalogimportdomain.StatusCompletedWithErrors)
	}
	if output.Job.Summary.ImportedRows != 1 {
		t.Fatalf("RunImport() imported rows = %d, want 1", output.Job.Summary.ImportedRows)
	}
	if output.Job.Summary.ErrorRows != 4 {
		t.Fatalf("RunImport() error rows = %d, want 4", output.Job.Summary.ErrorRows)
	}
	if output.Job.Summary.DuplicateInCatalogRows != 1 {
		t.Fatalf("RunImport() duplicate in catalog rows = %d, want 1", output.Job.Summary.DuplicateInCatalogRows)
	}
	if output.Job.Summary.DuplicateInFileRows != 1 {
		t.Fatalf("RunImport() duplicate in file rows = %d, want 1", output.Job.Summary.DuplicateInFileRows)
	}
	if len(repo.savedErrors) != 4 {
		t.Fatalf("RunImport() saved errors = %d, want 4", len(repo.savedErrors))
	}
}

func TestServiceRunImportBrokenFileMarksJobFailed(t *testing.T) {
	t.Parallel()

	service := mustImportService(t, importServiceDeps{
		repo: &fakeImportRepository{},
		parsers: fakeParserRegistry{
			parser: fakeParser{err: apperrors.New(apperrors.KindValidation, "invalid_import_file", "broken csv")},
		},
	})

	output, err := service.RunImport(context.Background(), RunImportInput{
		RequestedByUserID: testImportUserID,
		Filename:          "catalog.csv",
		FileSizeBytes:     10,
		File:              []byte("csv"),
	})
	if err != nil {
		t.Fatalf("RunImport() error = %v", err)
	}

	if output.Job.Status != string(catalogimportdomain.StatusFailed) {
		t.Fatalf("RunImport() status = %q, want %q", output.Job.Status, catalogimportdomain.StatusFailed)
	}
	if output.Job.FailureCode == nil || *output.Job.FailureCode != "invalid_import_file" {
		t.Fatalf("RunImport() failure code = %v, want invalid_import_file", output.Job.FailureCode)
	}
}

type importServiceDeps struct {
	repo             *fakeImportRepository
	parsers          fakeParserRegistry
	maxFileSizeBytes int64
}

func mustImportService(t *testing.T, deps importServiceDeps) *Service {
	t.Helper()

	maxFileSize := deps.maxFileSizeBytes
	if maxFileSize == 0 {
		maxFileSize = 1024
	}

	service, err := NewService(
		deps.repo,
		deps.parsers,
		fakeGiftIDGenerator{id: testImportGiftID},
		fakeImportJobIDGenerator{id: testImportJobID},
		fakeImportErrorIDGenerator{id: testImportErrorID},
		maxFileSize,
		fixedImportClock{now: time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)},
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	return service
}

type fakeImportRepository struct {
	categories   map[string]*catalogdomain.Category
	job          *catalogimportdomain.ImportJob
	savedErrors  []catalogimportdomain.ImportError
	existingKeys map[string]bool
	inserted     []catalogdomain.Gift
}

func (r *fakeImportRepository) CreateJob(_ context.Context, job *catalogimportdomain.ImportJob) error {
	clone := *job
	r.job = &clone
	return nil
}

func (r *fakeImportRepository) UpdateJob(_ context.Context, job *catalogimportdomain.ImportJob) error {
	clone := *job
	r.job = &clone
	return nil
}

func (r *fakeImportRepository) SaveErrors(_ context.Context, importErrors []catalogimportdomain.ImportError) error {
	r.savedErrors = append(r.savedErrors, importErrors...)
	return nil
}

func (r *fakeImportRepository) FindCategoryByName(_ context.Context, name string) (*catalogdomain.Category, error) {
	if r.categories == nil {
		return nil, nil
	}

	return r.categories[strings.ToLower(name)], nil
}

func (r *fakeImportRepository) GiftExists(_ context.Context, normalizedName, normalizedStoreLink string) (bool, error) {
	return r.existingKeys[normalizedName+"|"+normalizedStoreLink], nil
}

func (r *fakeImportRepository) InsertGift(_ context.Context, gift catalogdomain.Gift, _ *string) error {
	r.inserted = append(r.inserted, gift)
	return nil
}

func (r *fakeImportRepository) InsertOffers(context.Context, []catalogdomain.Offer) error {
	return nil
}

func (r *fakeImportRepository) GetJob(context.Context, catalogimportdomain.ImportJobID) (*catalogimportdomain.ImportJob, error) {
	return r.job, nil
}

func (r *fakeImportRepository) ListErrors(context.Context, catalogimportdomain.ImportJobID, int, int) ([]catalogimportdomain.ImportError, int, error) {
	return r.savedErrors, len(r.savedErrors), nil
}

type fakeParserRegistry struct {
	parser Parser
	err    error
}

func (r fakeParserRegistry) ParserFor(catalogimportdomain.SourceFormat) (Parser, error) {
	if r.err != nil {
		return nil, r.err
	}

	return r.parser, nil
}

type fakeParser struct {
	rows []ImportRowRaw
	err  error
}

func (p fakeParser) Parse(context.Context, []byte) ([]ImportRowRaw, error) {
	return p.rows, p.err
}

type fakeGiftIDGenerator struct {
	id string
}

func (g fakeGiftIDGenerator) NewGiftID() (catalogdomain.GiftID, error) {
	return catalogdomain.NewGiftID(g.id)
}

type fakeImportJobIDGenerator struct {
	id string
}

func (g fakeImportJobIDGenerator) NewImportJobID() (catalogimportdomain.ImportJobID, error) {
	return catalogimportdomain.NewImportJobID(g.id)
}

type fakeImportErrorIDGenerator struct {
	id string
}

func (g fakeImportErrorIDGenerator) NewImportErrorID() (catalogimportdomain.ImportErrorID, error) {
	return catalogimportdomain.NewImportErrorID(g.id)
}

type fixedImportClock struct {
	now time.Time
}

func (c fixedImportClock) Now() time.Time {
	return c.now.UTC()
}

func mustCategory(t *testing.T, name string) *catalogdomain.Category {
	t.Helper()

	category, err := catalogdomain.RestoreCategory(
		"550e8400-e29b-41d4-a716-446655440904",
		name,
		time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("RestoreCategory() error = %v", err)
	}

	return &category
}
