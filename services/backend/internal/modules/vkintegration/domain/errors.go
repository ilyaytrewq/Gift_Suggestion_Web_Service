package domain

import "errors"

var (
	ErrConnectionIDEmpty        = errors.New("vk connection id is empty")
	ErrInvalidConnectionID      = errors.New("vk connection id is invalid")
	ErrInvalidConnectionState   = errors.New("vk connection state is invalid")
	ErrInvalidConsentState      = errors.New("vk consent state is invalid")
	ErrInvalidSyncState         = errors.New("vk sync state is invalid")
	ErrProviderUserIDEmpty      = errors.New("vk provider user id is empty")
	ErrProviderUserIDTooLong    = errors.New("vk provider user id is too long")
	ErrConsentVersionEmpty      = errors.New("vk consent version is empty")
	ErrConsentVersionTooLong    = errors.New("vk consent version is too long")
	ErrConnectedAtRequired      = errors.New("vk connected_at is required")
	ErrConsentGrantedAtRequired = errors.New("vk consent_granted_at is required")
	ErrScreenNameTooLong        = errors.New("vk screen_name is too long")
	ErrProfileURLTooLong        = errors.New("vk profile_url is too long")
	ErrInvalidProfileURL        = errors.New("vk profile_url is invalid")
	ErrInvalidScope             = errors.New("vk scope is invalid")
	ErrTokenCiphertextTooLong   = errors.New("vk token ciphertext is too long")
	ErrSyncErrorCodeTooLong     = errors.New("vk sync error code is too long")
	ErrImportedInterestEmpty    = errors.New("vk imported interest is empty")
	ErrImportedInterestTooLong  = errors.New("vk imported interest is too long")
	ErrNormalizedInterestEmpty  = errors.New("vk normalized interest is empty")
	ErrInvalidInterestSource    = errors.New("vk interest source is invalid")
	ErrInvalidInterestPosition  = errors.New("vk interest position is invalid")
	ErrImportedInterestAtZero   = errors.New("vk imported interest timestamp is zero")
)
