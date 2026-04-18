package apperrors

import (
	"errors"
	"net/http"
)

type Kind string

const (
	KindValidation   Kind = "validation"
	KindNotFound     Kind = "not_found"
	KindMethod       Kind = "method_not_allowed"
	KindConflict     Kind = "conflict"
	KindUnauthorized Kind = "unauthorized"
	KindForbidden    Kind = "forbidden"
	KindUnavailable  Kind = "unavailable"
	KindInternal     Kind = "internal"
)

type Error struct {
	kind    Kind
	code    string
	message string
	cause   error
}

func New(kind Kind, code, message string) *Error {
	return &Error{
		kind:    kind,
		code:    code,
		message: message,
	}
}

func Wrap(kind Kind, code, message string, cause error) *Error {
	return &Error{
		kind:    kind,
		code:    code,
		message: message,
		cause:   cause,
	}
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	return e.message
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.cause
}

func (e *Error) Code() string {
	if e == nil {
		return "internal_error"
	}

	return e.code
}

func (e *Error) Message() string {
	if e == nil {
		return "internal server error"
	}

	return e.message
}

func (e *Error) Kind() Kind {
	if e == nil {
		return KindInternal
	}

	return e.kind
}

func (e *Error) HTTPStatus() int {
	switch e.Kind() {
	case KindValidation:
		return http.StatusBadRequest
	case KindNotFound:
		return http.StatusNotFound
	case KindMethod:
		return http.StatusMethodNotAllowed
	case KindConflict:
		return http.StatusConflict
	case KindUnauthorized:
		return http.StatusUnauthorized
	case KindForbidden:
		return http.StatusForbidden
	case KindUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func From(err error) *Error {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}

	return Wrap(
		KindInternal,
		"internal_error",
		"internal server error",
		err,
	)
}
