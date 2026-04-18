package domain

import (
	"strings"
	"time"

	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
)

type Status string

const (
	StatusPending             Status = "pending"
	StatusRunning             Status = "running"
	StatusCompleted           Status = "completed"
	StatusCompletedWithErrors Status = "completed_with_errors"
	StatusFailed              Status = "failed"
)

type Summary struct {
	TotalRows              int
	ProcessedRows          int
	ImportedRows           int
	UpdatedRows            int
	SkippedRows            int
	DuplicateInFileRows    int
	DuplicateInCatalogRows int
	ErrorRows              int
}

type ImportJob struct {
	id                ImportJobID
	requestedByUserID *userdomain.UserID
	status            Status
	sourceFormat      SourceFormat
	sourceFilename    string
	sourceLabel       string
	sourceSizeBytes   int64
	summary           Summary
	failureCode       string
	failureMessage    string
	startedAt         *time.Time
	finishedAt        *time.Time
	createdAt         time.Time
	updatedAt         time.Time
}

func NewImportJob(
	id ImportJobID,
	requestedByUserID *userdomain.UserID,
	format SourceFormat,
	filename string,
	sourceLabel string,
	sourceSizeBytes int64,
	now time.Time,
) (ImportJob, error) {
	normalizedFilename := strings.TrimSpace(filename)
	if normalizedFilename == "" {
		return ImportJob{}, ErrSourceFilenameEmpty
	}
	if sourceSizeBytes < 0 {
		return ImportJob{}, ErrNegativeSourceSize
	}

	timestamp := now.UTC()

	return ImportJob{
		id:                id,
		requestedByUserID: cloneUserID(requestedByUserID),
		status:            StatusPending,
		sourceFormat:      format,
		sourceFilename:    normalizedFilename,
		sourceLabel:       strings.TrimSpace(sourceLabel),
		sourceSizeBytes:   sourceSizeBytes,
		createdAt:         timestamp,
		updatedAt:         timestamp,
	}, nil
}

func RestoreImportJob(
	id string,
	requestedByUserID *string,
	status string,
	sourceFormat string,
	sourceFilename string,
	sourceLabel string,
	sourceSizeBytes int64,
	summary Summary,
	failureCode string,
	failureMessage string,
	startedAt *time.Time,
	finishedAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) (ImportJob, error) {
	jobID, err := NewImportJobID(id)
	if err != nil {
		return ImportJob{}, err
	}

	format, err := RestoreSourceFormat(sourceFormat)
	if err != nil {
		return ImportJob{}, err
	}

	var userID *userdomain.UserID
	if requestedByUserID != nil && strings.TrimSpace(*requestedByUserID) != "" {
		value, err := userdomain.NewUserID(strings.TrimSpace(*requestedByUserID))
		if err != nil {
			return ImportJob{}, err
		}
		userID = &value
	}

	job, err := NewImportJob(
		jobID,
		userID,
		format,
		sourceFilename,
		sourceLabel,
		sourceSizeBytes,
		createdAt,
	)
	if err != nil {
		return ImportJob{}, err
	}

	job.status = Status(strings.TrimSpace(status))
	job.summary = summary
	job.failureCode = strings.TrimSpace(failureCode)
	job.failureMessage = strings.TrimSpace(failureMessage)
	job.startedAt = cloneTimePtr(startedAt)
	job.finishedAt = cloneTimePtr(finishedAt)
	job.createdAt = createdAt.UTC()
	job.updatedAt = updatedAt.UTC()

	return job, nil
}

func (j *ImportJob) MarkRunning(now time.Time) {
	if j == nil {
		return
	}

	timestamp := now.UTC()
	j.status = StatusRunning
	j.startedAt = &timestamp
	j.updatedAt = timestamp
}

func (j *ImportJob) UpdateProgress(summary Summary, now time.Time) {
	if j == nil {
		return
	}

	j.summary = summary
	j.updatedAt = now.UTC()
}

func (j *ImportJob) Complete(summary Summary, now time.Time) {
	j.finish(StatusCompleted, summary, "", "", now)
}

func (j *ImportJob) CompleteWithErrors(summary Summary, now time.Time) {
	j.finish(StatusCompletedWithErrors, summary, "", "", now)
}

func (j *ImportJob) Fail(summary Summary, code, message string, now time.Time) {
	j.finish(StatusFailed, summary, code, message, now)
}

func (j *ImportJob) finish(status Status, summary Summary, code, message string, now time.Time) {
	if j == nil {
		return
	}

	timestamp := now.UTC()
	j.status = status
	j.summary = summary
	j.failureCode = strings.TrimSpace(code)
	j.failureMessage = strings.TrimSpace(message)
	j.finishedAt = &timestamp
	j.updatedAt = timestamp
	if j.startedAt == nil {
		j.startedAt = &timestamp
	}
}

func (j ImportJob) ID() ImportJobID {
	return j.id
}

func (j ImportJob) RequestedByUserID() *userdomain.UserID {
	return cloneUserID(j.requestedByUserID)
}

func (j ImportJob) Status() Status {
	return j.status
}

func (j ImportJob) SourceFormat() SourceFormat {
	return j.sourceFormat
}

func (j ImportJob) SourceFilename() string {
	return j.sourceFilename
}

func (j ImportJob) SourceLabel() string {
	return j.sourceLabel
}

func (j ImportJob) SourceSizeBytes() int64 {
	return j.sourceSizeBytes
}

func (j ImportJob) Summary() Summary {
	return j.summary
}

func (j ImportJob) FailureCode() string {
	return j.failureCode
}

func (j ImportJob) FailureMessage() string {
	return j.failureMessage
}

func (j ImportJob) StartedAt() *time.Time {
	return cloneTimePtr(j.startedAt)
}

func (j ImportJob) FinishedAt() *time.Time {
	return cloneTimePtr(j.finishedAt)
}

func (j ImportJob) CreatedAt() time.Time {
	return j.createdAt
}

func (j ImportJob) UpdatedAt() time.Time {
	return j.updatedAt
}

func cloneUserID(value *userdomain.UserID) *userdomain.UserID {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}

func cloneTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	cloned := value.UTC()
	return &cloned
}
