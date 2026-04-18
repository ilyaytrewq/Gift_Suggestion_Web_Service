package usecase

import (
	"time"

	catalogimportdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalogimport/domain"
)

type RunImportInput struct {
	RequestedByUserID string
	Filename          string
	SourceLabel       string
	FileSizeBytes     int64
	File              []byte
}

type GetImportJobInput struct {
	JobID string
}

type ListImportErrorsInput struct {
	JobID  string
	Limit  int
	Offset int
}

type Page struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

type Summary struct {
	TotalRows              int `json:"total_rows"`
	ProcessedRows          int `json:"processed_rows"`
	ImportedRows           int `json:"imported_rows"`
	UpdatedRows            int `json:"updated_rows"`
	SkippedRows            int `json:"skipped_rows"`
	DuplicateInFileRows    int `json:"duplicate_in_file_rows"`
	DuplicateInCatalogRows int `json:"duplicate_in_catalog_rows"`
	ErrorRows              int `json:"error_rows"`
}

type Job struct {
	ID                string     `json:"id"`
	RequestedByUserID *string    `json:"requested_by_user_id,omitempty"`
	Status            string     `json:"status"`
	SourceFormat      string     `json:"source_format"`
	SourceFilename    string     `json:"source_filename"`
	SourceLabel       *string    `json:"source_label,omitempty"`
	SourceSizeBytes   int64      `json:"source_size_bytes"`
	Summary           Summary    `json:"summary"`
	FailureCode       *string    `json:"failure_code,omitempty"`
	FailureMessage    *string    `json:"failure_message,omitempty"`
	StartedAt         *time.Time `json:"started_at,omitempty"`
	FinishedAt        *time.Time `json:"finished_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type ImportError struct {
	ID        string  `json:"id"`
	RowNumber *int    `json:"row_number,omitempty"`
	RecordKey *string `json:"record_key,omitempty"`
	FieldName *string `json:"field_name,omitempty"`
	Code      string  `json:"code"`
	Message   string  `json:"message"`
	RawRecord *string `json:"raw_record,omitempty"`
}

type RunImportOutput struct {
	Job Job `json:"job"`
}

type GetImportJobOutput struct {
	Job Job `json:"job"`
}

type ListImportErrorsOutput struct {
	Items []ImportError `json:"items"`
	Page  Page          `json:"page"`
}

func newJob(job catalogimportdomain.ImportJob) Job {
	var requestedByUserID *string
	if job.RequestedByUserID() != nil {
		value := job.RequestedByUserID().String()
		requestedByUserID = &value
	}

	var sourceLabel *string
	if job.SourceLabel() != "" {
		value := job.SourceLabel()
		sourceLabel = &value
	}

	var failureCode *string
	if job.FailureCode() != "" {
		value := job.FailureCode()
		failureCode = &value
	}

	var failureMessage *string
	if job.FailureMessage() != "" {
		value := job.FailureMessage()
		failureMessage = &value
	}

	summary := job.Summary()

	return Job{
		ID:                job.ID().String(),
		RequestedByUserID: requestedByUserID,
		Status:            string(job.Status()),
		SourceFormat:      string(job.SourceFormat()),
		SourceFilename:    job.SourceFilename(),
		SourceLabel:       sourceLabel,
		SourceSizeBytes:   job.SourceSizeBytes(),
		Summary: Summary{
			TotalRows:              summary.TotalRows,
			ProcessedRows:          summary.ProcessedRows,
			ImportedRows:           summary.ImportedRows,
			UpdatedRows:            summary.UpdatedRows,
			SkippedRows:            summary.SkippedRows,
			DuplicateInFileRows:    summary.DuplicateInFileRows,
			DuplicateInCatalogRows: summary.DuplicateInCatalogRows,
			ErrorRows:              summary.ErrorRows,
		},
		FailureCode:    failureCode,
		FailureMessage: failureMessage,
		StartedAt:      job.StartedAt(),
		FinishedAt:     job.FinishedAt(),
		CreatedAt:      job.CreatedAt(),
		UpdatedAt:      job.UpdatedAt(),
	}
}

func newImportError(importError catalogimportdomain.ImportError) ImportError {
	return ImportError{
		ID:        importError.ID().String(),
		RowNumber: importError.RowNumber(),
		RecordKey: importError.RecordKey(),
		FieldName: importError.FieldName(),
		Code:      importError.Code(),
		Message:   importError.Message(),
		RawRecord: importError.RawRecord(),
	}
}
