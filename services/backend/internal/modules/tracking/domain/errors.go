package domain

import "errors"

var (
	ErrEventIDEmpty                    = errors.New("tracking event id is empty")
	ErrInvalidEventID                  = errors.New("tracking event id has invalid format")
	ErrInvalidEventType                = errors.New("tracking event type is invalid")
	ErrClientEventIDTooLong            = errors.New("tracking client event id is too long")
	ErrOccurredAtZero                  = errors.New("tracking event occurred at is zero")
	ErrRecommendationRequestIDRequired = errors.New("recommendation request id is required")
	ErrGiftIDRequired                  = errors.New("gift id is required")
	ErrWishlistIDRequired              = errors.New("wishlist id is required")
	ErrGiftIDForbidden                 = errors.New("gift id is forbidden for event type")
	ErrWishlistIDForbidden             = errors.New("wishlist id is forbidden for event type")
	ErrMetadataSurfaceTooLong          = errors.New("tracking metadata surface is too long")
	ErrInvalidMetadataSurface          = errors.New("tracking metadata surface is invalid")
	ErrInvalidMetadataPosition         = errors.New("tracking metadata position is invalid")
)
