package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	catalogimportdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalogimport/domain"
	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type Service struct {
	repo                 Repository
	parsers              ParserRegistry
	giftIDGenerator      GiftIDGenerator
	importJobIDGenerator ImportJobIDGenerator
	importErrIDGenerator ImportErrorIDGenerator
	maxFileSizeBytes     int64
	clock                Clock
}

func NewService(
	repo Repository,
	parsers ParserRegistry,
	giftIDGenerator GiftIDGenerator,
	importJobIDGenerator ImportJobIDGenerator,
	importErrIDGenerator ImportErrorIDGenerator,
	maxFileSizeBytes int64,
	clock Clock,
) (*Service, error) {
	switch {
	case repo == nil:
		return nil, ErrNilJobRepository
	case parsers == nil:
		return nil, ErrNilParserRegistry
	case giftIDGenerator == nil:
		return nil, ErrNilGiftIDGenerator
	case importJobIDGenerator == nil:
		return nil, ErrNilImportJobIDGenerator
	case importErrIDGenerator == nil:
		return nil, ErrNilImportErrorIDGenerator
	case clock == nil:
		return nil, ErrNilClock
	}

	return &Service{
		repo:                 repo,
		parsers:              parsers,
		giftIDGenerator:      giftIDGenerator,
		importJobIDGenerator: importJobIDGenerator,
		importErrIDGenerator: importErrIDGenerator,
		maxFileSizeBytes:     maxFileSizeBytes,
		clock:                clock,
	}, nil
}

func (s *Service) RunImport(ctx context.Context, input RunImportInput) (RunImportOutput, error) {
	requestedByUserID, sourceFormat, parser, err := s.prepareImport(input)
	if err != nil {
		return RunImportOutput{}, err
	}

	job, err := s.createRunningJob(ctx, requestedByUserID, sourceFormat, input)
	if err != nil {
		return RunImportOutput{}, err
	}

	rows, err := parser.Parse(ctx, input.File)
	if err != nil {
		return s.failJobWithParseError(ctx, &job, err)
	}
	if len(rows) == 0 {
		return s.failJobWithParseError(ctx, &job, apperrors.New(
			apperrors.KindValidation,
			"empty_import_file",
			"import file does not contain any records",
		))
	}

	summary := job.Summary()
	summary.TotalRows = len(rows)
	job.UpdateProgress(summary, s.clock.Now())
	if err := s.repo.UpdateJob(ctx, &job); err != nil {
		return RunImportOutput{}, err
	}

	if err := s.processRows(ctx, &job, input.SourceLabel, rows, &summary); err != nil {
		return RunImportOutput{}, err
	}

	if summary.ErrorRows > 0 {
		job.CompleteWithErrors(summary, s.clock.Now())
	} else {
		job.Complete(summary, s.clock.Now())
	}
	if err := s.repo.UpdateJob(ctx, &job); err != nil {
		return RunImportOutput{}, err
	}

	return RunImportOutput{Job: newJob(job)}, nil
}

func (s *Service) prepareImport(
	input RunImportInput,
) (userdomain.UserID, catalogimportdomain.SourceFormat, Parser, error) {
	requestedByUserID, err := userdomain.NewUserID(input.RequestedByUserID)
	if err != nil {
		return userdomain.UserID{}, "", nil, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_requested_by_user_id",
			"requested by user id is invalid",
			err,
		)
	}

	if input.FileSizeBytes < 1 || len(input.File) == 0 {
		return userdomain.UserID{}, "", nil, apperrors.New(
			apperrors.KindValidation,
			"empty_import_file",
			"import file must not be empty",
		)
	}
	if s.maxFileSizeBytes > 0 && input.FileSizeBytes > s.maxFileSizeBytes {
		return userdomain.UserID{}, "", nil, apperrors.New(
			apperrors.KindValidation,
			"file_too_large",
			"import file exceeds size limit",
		)
	}

	sourceFormat, err := catalogimportdomain.DetectSourceFormat(input.Filename)
	if err != nil {
		return userdomain.UserID{}, "", nil, apperrors.Wrap(
			apperrors.KindValidation,
			"unsupported_import_format",
			"import format is not supported",
			err,
		)
	}

	parser, err := s.parsers.ParserFor(sourceFormat)
	if err != nil {
		return userdomain.UserID{}, "", nil, err
	}

	return requestedByUserID, sourceFormat, parser, nil
}

func (s *Service) createRunningJob(
	ctx context.Context,
	requestedByUserID userdomain.UserID,
	sourceFormat catalogimportdomain.SourceFormat,
	input RunImportInput,
) (catalogimportdomain.ImportJob, error) {
	jobID, err := s.importJobIDGenerator.NewImportJobID()
	if err != nil {
		return catalogimportdomain.ImportJob{}, err
	}

	now := s.clock.Now()
	job, err := catalogimportdomain.NewImportJob(
		jobID,
		&requestedByUserID,
		sourceFormat,
		input.Filename,
		input.SourceLabel,
		input.FileSizeBytes,
		now,
	)
	if err != nil {
		return catalogimportdomain.ImportJob{}, err
	}

	if err := s.repo.CreateJob(ctx, &job); err != nil {
		return catalogimportdomain.ImportJob{}, err
	}

	job.MarkRunning(now)
	if err := s.repo.UpdateJob(ctx, &job); err != nil {
		return catalogimportdomain.ImportJob{}, err
	}

	return job, nil
}

func (s *Service) processRows(
	ctx context.Context,
	job *catalogimportdomain.ImportJob,
	defaultSourceLabel string,
	rows []ImportRowRaw,
	summary *catalogimportdomain.Summary,
) error {
	seenKeys := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		if err := s.processRow(ctx, job, defaultSourceLabel, seenKeys, row, summary); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) processRow(
	ctx context.Context,
	job *catalogimportdomain.ImportJob,
	defaultSourceLabel string,
	seenKeys map[string]struct{},
	row ImportRowRaw,
	summary *catalogimportdomain.Summary,
) error {
	summary.ProcessedRows++

	record, recordKey, rowErr, err := s.normalizeRow(ctx, row, defaultSourceLabel)
	if err != nil {
		return err
	}
	if rowErr != nil {
		return s.persistRowFailure(ctx, job, row, recordKey, *rowErr, summary)
	}

	if _, exists := seenKeys[*recordKey]; exists {
		return s.persistRowFailure(ctx, job, row, recordKey, rowError{
			Field:   "name",
			Code:    "duplicate_file_record",
			Message: "duplicate record inside import file",
		}, summary)
	}
	seenKeys[*recordKey] = struct{}{}

	exists, err := s.repo.GiftExists(ctx, record.NormalizedName, record.NormalizedStoreLink)
	if err != nil {
		return err
	}
	if exists {
		return s.persistRowFailure(ctx, job, row, recordKey, rowError{
			Field:   "name",
			Code:    "duplicate_catalog_entry",
			Message: "gift already exists in catalog",
		}, summary)
	}

	if err := s.repo.InsertGift(ctx, record.Gift, record.SourceName); err != nil {
		return err
	}

	if len(record.Offers) > 0 {
		if err := s.repo.InsertOffers(ctx, record.Offers); err != nil {
			return err
		}
	}

	summary.ImportedRows++
	job.UpdateProgress(*summary, s.clock.Now())
	return s.repo.UpdateJob(ctx, job)
}

func (s *Service) persistRowFailure(
	ctx context.Context,
	job *catalogimportdomain.ImportJob,
	row ImportRowRaw,
	recordKey *string,
	rowErr rowError,
	summary *catalogimportdomain.Summary,
) error {
	summary.ErrorRows++
	summary.SkippedRows++
	if rowErr.Code == "duplicate_file_record" {
		summary.DuplicateInFileRows++
	}
	if rowErr.Code == "duplicate_catalog_entry" {
		summary.DuplicateInCatalogRows++
	}

	if err := s.persistRowError(ctx, job.ID(), row, recordKey, rowErr); err != nil {
		return err
	}

	job.UpdateProgress(*summary, s.clock.Now())
	return s.repo.UpdateJob(ctx, job)
}

func (s *Service) GetImportJob(ctx context.Context, input GetImportJobInput) (GetImportJobOutput, error) {
	jobID, err := catalogimportdomain.NewImportJobID(input.JobID)
	if err != nil {
		return GetImportJobOutput{}, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_import_job_id",
			"import job id is invalid",
			err,
		)
	}

	job, err := s.repo.GetJob(ctx, jobID)
	if err != nil {
		return GetImportJobOutput{}, err
	}
	if job == nil {
		return GetImportJobOutput{}, apperrors.New(
			apperrors.KindNotFound,
			"import_job_not_found",
			"import job not found",
		)
	}

	return GetImportJobOutput{Job: newJob(*job)}, nil
}

func (s *Service) ListImportErrors(ctx context.Context, input ListImportErrorsInput) (ListImportErrorsOutput, error) {
	jobID, err := catalogimportdomain.NewImportJobID(input.JobID)
	if err != nil {
		return ListImportErrorsOutput{}, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_import_job_id",
			"import job id is invalid",
			err,
		)
	}

	limit, err := normalizeLimit(input.Limit)
	if err != nil {
		return ListImportErrorsOutput{}, err
	}
	offset, err := normalizeOffset(input.Offset)
	if err != nil {
		return ListImportErrorsOutput{}, err
	}

	job, err := s.repo.GetJob(ctx, jobID)
	if err != nil {
		return ListImportErrorsOutput{}, err
	}
	if job == nil {
		return ListImportErrorsOutput{}, apperrors.New(
			apperrors.KindNotFound,
			"import_job_not_found",
			"import job not found",
		)
	}

	items, total, err := s.repo.ListErrors(ctx, jobID, limit, offset)
	if err != nil {
		return ListImportErrorsOutput{}, err
	}

	outputItems := make([]ImportError, 0, len(items))
	for _, item := range items {
		outputItems = append(outputItems, newImportError(item))
	}

	return ListImportErrorsOutput{
		Items: outputItems,
		Page: Page{
			Limit:  limit,
			Offset: offset,
			Total:  total,
		},
	}, nil
}

func (s *Service) failJobWithParseError(
	ctx context.Context,
	job *catalogimportdomain.ImportJob,
	parseErr error,
) (RunImportOutput, error) {
	if job == nil {
		return RunImportOutput{}, parseErr
	}

	summary := job.Summary()
	summary.ErrorRows = 1
	summary.SkippedRows = 1

	importErr, err := s.newImportError(
		job.ID(),
		nil,
		nil,
		nil,
		"invalid_import_file",
		"import file is invalid",
		nil,
	)
	if err != nil {
		return RunImportOutput{}, err
	}

	if err := s.repo.SaveErrors(ctx, []catalogimportdomain.ImportError{importErr}); err != nil {
		return RunImportOutput{}, err
	}

	job.Fail(summary, "invalid_import_file", parseErr.Error(), s.clock.Now())
	if err := s.repo.UpdateJob(ctx, job); err != nil {
		return RunImportOutput{}, err
	}

	return RunImportOutput{Job: newJob(*job)}, nil
}

func (s *Service) persistRowError(
	ctx context.Context,
	jobID catalogimportdomain.ImportJobID,
	row ImportRowRaw,
	recordKey *string,
	rowErr rowError,
) error {
	rawRecord, err := marshalRawRow(row)
	if err != nil {
		return err
	}

	importErr, err := s.newImportError(
		jobID,
		&row.RowNumber,
		recordKey,
		optionalString(rowErr.Field),
		rowErr.Code,
		rowErr.Message,
		rawRecord,
	)
	if err != nil {
		return err
	}

	return s.repo.SaveErrors(ctx, []catalogimportdomain.ImportError{importErr})
}

func (s *Service) newImportError(
	jobID catalogimportdomain.ImportJobID,
	rowNumber *int,
	recordKey *string,
	fieldName *string,
	code string,
	message string,
	rawRecord *string,
) (catalogimportdomain.ImportError, error) {
	importErrorID, err := s.importErrIDGenerator.NewImportErrorID()
	if err != nil {
		return catalogimportdomain.ImportError{}, err
	}

	return catalogimportdomain.NewImportError(
		importErrorID,
		jobID,
		rowNumber,
		recordKey,
		fieldName,
		code,
		message,
		rawRecord,
	)
}

func marshalRawRow(row ImportRowRaw) (*string, error) {
	payload, err := json.Marshal(row)
	if err != nil {
		return nil, err
	}

	value := string(payload)
	return &value, nil
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func normalizeLimit(limit int) (int, error) {
	if limit == 0 {
		return defaultLimit, nil
	}
	if limit < 1 || limit > maxLimit {
		return 0, apperrors.New(
			apperrors.KindValidation,
			"invalid_limit",
			"limit must be between 1 and 100",
		)
	}

	return limit, nil
}

func normalizeOffset(offset int) (int, error) {
	if offset < 0 {
		return 0, apperrors.New(
			apperrors.KindValidation,
			"invalid_offset",
			"offset must be greater than or equal to zero",
		)
	}

	return offset, nil
}

func mapCatalogValidationError(err error) rowError {
	switch {
	case errors.Is(err, catalogdomain.ErrGiftNameEmpty):
		return rowError{Field: "name", Code: "missing_required_field", Message: "gift name is required"}
	case errors.Is(err, catalogdomain.ErrPriceEmpty):
		return rowError{Field: "price", Code: "missing_required_field", Message: "price is required"}
	case errors.Is(err, catalogdomain.ErrInvalidPrice), errors.Is(err, catalogdomain.ErrNegativePrice):
		return rowError{Field: "price", Code: "invalid_price", Message: "price is invalid"}
	case errors.Is(err, catalogdomain.ErrStoreLinkEmpty):
		return rowError{Field: "store_link", Code: "missing_required_field", Message: "store link is required"}
	case errors.Is(err, catalogdomain.ErrInvalidStoreLink):
		return rowError{Field: "store_link", Code: "invalid_store_link", Message: "store link is invalid"}
	case errors.Is(err, catalogdomain.ErrInvalidAgeRestriction):
		return rowError{Field: "age_restriction", Code: "invalid_age_restriction", Message: "age restriction is invalid"}
	default:
		return rowError{Code: "invalid_record", Message: "record is invalid"}
	}
}
