package domain

import "errors"

var (
	ErrRequestIDEmpty             = errors.New("recommendation request id is empty")
	ErrInvalidRequestID           = errors.New("recommendation request id has invalid format")
	ErrResultIDEmpty              = errors.New("recommendation result id is empty")
	ErrInvalidResultID            = errors.New("recommendation result id has invalid format")
	ErrBudgetEmpty                = errors.New("budget is empty")
	ErrOccasionTooLong            = errors.New("occasion is too long")
	ErrRelationshipTooLong        = errors.New("relationship is too long")
	ErrRecipientAgeInvalid        = errors.New("recipient age is invalid")
	ErrInvalidGender              = errors.New("recipient gender is invalid; allowed values: male, female, other")
	ErrTopNInvalid                = errors.New("top n is invalid")
	ErrTooManyPreferredCategories = errors.New("too many preferred categories")
	ErrTooManyInterests           = errors.New("too many interests")
	ErrInterestTooLong            = errors.New("interest is too long")
	ErrExplanationCodeEmpty       = errors.New("recommendation explanation code is empty")
	ErrExplanationTextEmpty       = errors.New("recommendation explanation text is empty")
	ErrInvalidRequestStatus       = errors.New("recommendation request status is invalid")
	ErrInvalidRankingSource       = errors.New("recommendation ranking source is invalid")
	ErrInvalidResultKind          = errors.New("recommendation result kind is invalid")
	ErrSlotPositionInvalid        = errors.New("slot position is invalid")
	ErrAlternativeRankInvalid     = errors.New("alternative rank is invalid")
)
