package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	catalogimportdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalogimport/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateJob(ctx context.Context, job *catalogimportdomain.ImportJob) error {
	const query = `
		INSERT INTO import_jobs (
			id,
			requested_by_user_id,
			status,
			source_format,
			source_filename,
			source_label,
			source_size_bytes,
			total_rows,
			processed_rows,
			imported_rows,
			updated_rows,
			skipped_rows,
			duplicate_in_file_rows,
			duplicate_in_catalog_rows,
			error_rows,
			failure_code,
			failure_message,
			started_at,
			finished_at,
			created_at,
			updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21
		)
	`

	summary := job.Summary()
	_, err := r.db.ExecContext(
		ctx,
		query,
		job.ID().String(),
		userIDValue(job.RequestedByUserID()),
		string(job.Status()),
		string(job.SourceFormat()),
		job.SourceFilename(),
		nullString(job.SourceLabel()),
		job.SourceSizeBytes(),
		summary.TotalRows,
		summary.ProcessedRows,
		summary.ImportedRows,
		summary.UpdatedRows,
		summary.SkippedRows,
		summary.DuplicateInFileRows,
		summary.DuplicateInCatalogRows,
		summary.ErrorRows,
		nullString(job.FailureCode()),
		nullString(job.FailureMessage()),
		nullTime(job.StartedAt()),
		nullTime(job.FinishedAt()),
		job.CreatedAt(),
		job.UpdatedAt(),
	)

	return err
}

func (r *Repository) UpdateJob(ctx context.Context, job *catalogimportdomain.ImportJob) error {
	const query = `
		UPDATE import_jobs
		SET
			requested_by_user_id = $2,
			status = $3,
			source_format = $4,
			source_filename = $5,
			source_label = $6,
			source_size_bytes = $7,
			total_rows = $8,
			processed_rows = $9,
			imported_rows = $10,
			updated_rows = $11,
			skipped_rows = $12,
			duplicate_in_file_rows = $13,
			duplicate_in_catalog_rows = $14,
			error_rows = $15,
			failure_code = $16,
			failure_message = $17,
			started_at = $18,
			finished_at = $19,
			updated_at = $20
		WHERE id = $1
	`

	summary := job.Summary()
	result, err := r.db.ExecContext(
		ctx,
		query,
		job.ID().String(),
		userIDValue(job.RequestedByUserID()),
		string(job.Status()),
		string(job.SourceFormat()),
		job.SourceFilename(),
		nullString(job.SourceLabel()),
		job.SourceSizeBytes(),
		summary.TotalRows,
		summary.ProcessedRows,
		summary.ImportedRows,
		summary.UpdatedRows,
		summary.SkippedRows,
		summary.DuplicateInFileRows,
		summary.DuplicateInCatalogRows,
		summary.ErrorRows,
		nullString(job.FailureCode()),
		nullString(job.FailureMessage()),
		nullTime(job.StartedAt()),
		nullTime(job.FinishedAt()),
		job.UpdatedAt(),
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("import job not found")
	}

	return nil
}

func (r *Repository) SaveErrors(ctx context.Context, importErrors []catalogimportdomain.ImportError) error {
	if len(importErrors) == 0 {
		return nil
	}

	const query = `
		INSERT INTO import_errors (
			id,
			job_id,
			row_number,
			record_key,
			field_name,
			error_code,
			message,
			raw_record
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb)
	`

	for _, importError := range importErrors {
		_, err := r.db.ExecContext(
			ctx,
			query,
			importError.ID().String(),
			importError.JobID().String(),
			nullInt(importError.RowNumber()),
			nullStringPtr(importError.RecordKey()),
			nullStringPtr(importError.FieldName()),
			importError.Code(),
			importError.Message(),
			nullStringPtr(importError.RawRecord()),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) FindCategoryByName(ctx context.Context, name string) (*catalogdomain.Category, error) {
	const query = `
		SELECT id, name, created_at, updated_at
		FROM categories
		WHERE LOWER(name) = LOWER($1)
		LIMIT 1
	`

	var (
		id        string
		category  string
		createdAt sql.NullTime
		updatedAt sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, query, name).Scan(&id, &category, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	result, err := catalogdomain.RestoreCategory(id, category, createdAt.Time, updatedAt.Time)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *Repository) GiftExists(ctx context.Context, normalizedName, normalizedStoreLink string) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM gifts
			WHERE LOWER(name) = $1 AND LOWER(store_link) = $2
		)
	`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, normalizedName, normalizedStoreLink).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *Repository) InsertGift(ctx context.Context, gift catalogdomain.Gift, sourceName *string) error {
	const query = `
		INSERT INTO gifts (
			id,
			category_id,
			name,
			description,
			price,
			store_link,
			image,
			age_restriction,
			source_name,
			created_at,
			updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`

	var categoryID any
	if gift.CategoryID() != nil {
		categoryID = gift.CategoryID().String()
	}

	var ageRestriction any
	if gift.AgeRestriction() != nil {
		ageRestriction = gift.AgeRestriction().Int()
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		gift.ID().String(),
		categoryID,
		gift.Name(),
		gift.Description(),
		gift.Price().DecimalString(),
		gift.StoreLink(),
		nullStringPtr(gift.Image()),
		ageRestriction,
		nullStringPtr(sourceName),
		gift.CreatedAt(),
		gift.UpdatedAt(),
	)

	return err
}

func (r *Repository) GetJob(ctx context.Context, id catalogimportdomain.ImportJobID) (*catalogimportdomain.ImportJob, error) {
	const query = `
		SELECT
			id,
			requested_by_user_id,
			status,
			source_format,
			source_filename,
			source_label,
			source_size_bytes,
			total_rows,
			processed_rows,
			imported_rows,
			updated_rows,
			skipped_rows,
			duplicate_in_file_rows,
			duplicate_in_catalog_rows,
			error_rows,
			failure_code,
			failure_message,
			started_at,
			finished_at,
			created_at,
			updated_at
		FROM import_jobs
		WHERE id = $1
	`

	var (
		jobID                  string
		requestedByUserID      sql.NullString
		status                 string
		sourceFormat           string
		sourceFilename         string
		sourceLabel            sql.NullString
		sourceSizeBytes        int64
		totalRows              int
		processedRows          int
		importedRows           int
		updatedRows            int
		skippedRows            int
		duplicateInFileRows    int
		duplicateInCatalogRows int
		errorRows              int
		failureCode            sql.NullString
		failureMessage         sql.NullString
		startedAt              sql.NullTime
		finishedAt             sql.NullTime
		createdAt              sql.NullTime
		updatedAt              sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&jobID,
		&requestedByUserID,
		&status,
		&sourceFormat,
		&sourceFilename,
		&sourceLabel,
		&sourceSizeBytes,
		&totalRows,
		&processedRows,
		&importedRows,
		&updatedRows,
		&skippedRows,
		&duplicateInFileRows,
		&duplicateInCatalogRows,
		&errorRows,
		&failureCode,
		&failureMessage,
		&startedAt,
		&finishedAt,
		&createdAt,
		&updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	summary := catalogimportdomain.Summary{
		TotalRows:              totalRows,
		ProcessedRows:          processedRows,
		ImportedRows:           importedRows,
		UpdatedRows:            updatedRows,
		SkippedRows:            skippedRows,
		DuplicateInFileRows:    duplicateInFileRows,
		DuplicateInCatalogRows: duplicateInCatalogRows,
		ErrorRows:              errorRows,
	}

	job, err := catalogimportdomain.RestoreImportJob(
		jobID,
		stringPtrFromNull(requestedByUserID),
		status,
		sourceFormat,
		sourceFilename,
		sourceLabel.String,
		sourceSizeBytes,
		summary,
		failureCode.String,
		failureMessage.String,
		timePtrFromNull(startedAt),
		timePtrFromNull(finishedAt),
		createdAt.Time,
		updatedAt.Time,
	)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (r *Repository) ListErrors(
	ctx context.Context,
	jobID catalogimportdomain.ImportJobID,
	limit int,
	offset int,
) (items []catalogimportdomain.ImportError, total int, err error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM import_errors
		WHERE job_id = $1
	`

	if err := r.db.QueryRowContext(ctx, countQuery, jobID.String()).Scan(&total); err != nil {
		return nil, 0, err
	}

	const query = `
		SELECT id, job_id, row_number, record_key, field_name, error_code, message, raw_record
		FROM import_errors
		WHERE job_id = $1
		ORDER BY COALESCE(row_number, 0) ASC, id ASC
		LIMIT $2
		OFFSET $3
	`

	var rows *sql.Rows
	rows, err = r.db.QueryContext(ctx, query, jobID.String(), limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		closeErr := rows.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	items = make([]catalogimportdomain.ImportError, 0, limit)
	for rows.Next() {
		var (
			id        string
			rawJobID  string
			rowNumber sql.NullInt32
			recordKey sql.NullString
			fieldName sql.NullString
			code      string
			message   string
			rawRecord sql.NullString
		)

		if err := rows.Scan(&id, &rawJobID, &rowNumber, &recordKey, &fieldName, &code, &message, &rawRecord); err != nil {
			return nil, 0, err
		}

		item, err := catalogimportdomain.RestoreImportError(
			id,
			rawJobID,
			intPtrFromNull(rowNumber),
			stringPtrFromNull(recordKey),
			stringPtrFromNull(fieldName),
			code,
			message,
			stringPtrFromNull(rawRecord),
		)
		if err != nil {
			return nil, 0, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func userIDValue(id any) any {
	if id == nil {
		return nil
	}

	type stringer interface{ String() string }
	if value, ok := id.(stringer); ok {
		return value.String()
	}

	return nil
}

func nullString(value string) any {
	if value == "" {
		return nil
	}

	return value
}

func nullStringPtr(value *string) any {
	if value == nil || *value == "" {
		return nil
	}

	return *value
}

func nullTime(value *time.Time) any {
	if value == nil {
		return nil
	}

	return value.UTC()
}

func nullInt(value *int) any {
	if value == nil {
		return nil
	}

	return *value
}

func stringPtrFromNull(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	result := value.String
	return &result
}

func intPtrFromNull(value sql.NullInt32) *int {
	if !value.Valid {
		return nil
	}

	result := int(value.Int32)
	return &result
}

func timePtrFromNull(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}

	result := value.Time.UTC()
	return &result
}
