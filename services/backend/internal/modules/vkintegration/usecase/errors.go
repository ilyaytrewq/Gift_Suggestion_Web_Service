package usecase

import "errors"

var (
	ErrNilConnectionRepository = errors.New("vk connection repository is nil")
	ErrNilUserReader           = errors.New("vk user reader is nil")
	ErrNilTokenProtector       = errors.New("vk token protector is nil")
	ErrNilInterestImporter     = errors.New("vk interest importer is nil")
	ErrNilConnectionIDGen      = errors.New("vk connection id generator is nil")
	ErrNilClock                = errors.New("vk clock is nil")

	ErrInterestImportNotImplemented = errors.New("vk interest import is not implemented")
	ErrInterestImportUnavailable    = errors.New("vk interest import is unavailable")
	ErrInvalidInterestImportResult  = errors.New("vk interest import result is invalid")
	ErrTokenProtectionUnavailable   = errors.New("vk token protection is unavailable")
	ErrTokenCiphertextCorrupted     = errors.New("vk token ciphertext is corrupted")
	ErrVKUserAlreadyConnected       = errors.New("vk user is already connected to another account")

	ErrVKTokenInvalid       = errors.New("vk token is invalid or expired, reconnection required")
	ErrVKRateLimited        = errors.New("vk api rate limit exceeded")
	ErrVKGroupsAccessDenied = errors.New("vk groups list access denied by privacy settings")
	ErrVKGroupsScopeRequired = errors.New("vk groups scope is not granted for this token")
)
