package domain

import "strings"

type ImportError struct {
	id        ImportErrorID
	jobID     ImportJobID
	rowNumber *int
	recordKey *string
	fieldName *string
	code      string
	message   string
	rawRecord *string
}

func NewImportError(
	id ImportErrorID,
	jobID ImportJobID,
	rowNumber *int,
	recordKey *string,
	fieldName *string,
	code string,
	message string,
	rawRecord *string,
) (ImportError, error) {
	normalizedCode := strings.TrimSpace(code)
	if normalizedCode == "" {
		return ImportError{}, ErrImportErrorCodeEmpty
	}

	normalizedMessage := strings.TrimSpace(message)
	if normalizedMessage == "" {
		return ImportError{}, ErrImportErrorMessageEmpty
	}

	return ImportError{
		id:        id,
		jobID:     jobID,
		rowNumber: cloneIntPtr(rowNumber),
		recordKey: cloneStringPtr(recordKey),
		fieldName: cloneStringPtr(fieldName),
		code:      normalizedCode,
		message:   normalizedMessage,
		rawRecord: cloneStringPtr(rawRecord),
	}, nil
}

func RestoreImportError(
	id string,
	jobID string,
	rowNumber *int,
	recordKey *string,
	fieldName *string,
	code string,
	message string,
	rawRecord *string,
) (ImportError, error) {
	importErrorID, err := NewImportErrorID(id)
	if err != nil {
		return ImportError{}, err
	}

	importJobID, err := NewImportJobID(jobID)
	if err != nil {
		return ImportError{}, err
	}

	return NewImportError(importErrorID, importJobID, rowNumber, recordKey, fieldName, code, message, rawRecord)
}

func (e ImportError) ID() ImportErrorID {
	return e.id
}

func (e ImportError) JobID() ImportJobID {
	return e.jobID
}

func (e ImportError) RowNumber() *int {
	return cloneIntPtr(e.rowNumber)
}

func (e ImportError) RecordKey() *string {
	return cloneStringPtr(e.recordKey)
}

func (e ImportError) FieldName() *string {
	return cloneStringPtr(e.fieldName)
}

func (e ImportError) Code() string {
	return e.code
}

func (e ImportError) Message() string {
	return e.message
}

func (e ImportError) RawRecord() *string {
	return cloneStringPtr(e.rawRecord)
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}

	cloned := strings.TrimSpace(*value)
	return &cloned
}

func cloneIntPtr(value *int) *int {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}
