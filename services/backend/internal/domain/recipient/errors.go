package recipient

import "errors"

var (
	ErrRecipientProfileIDEmpty   = errors.New("recipient profile id is empty")
	ErrInvalidRecipientProfileID = errors.New("recipient profile id has invalid format")
	ErrInvalidOwnerUserID        = errors.New("owner user id has invalid format")
	ErrInvalidOccasion           = errors.New("invalid occasion")
	ErrInvalidRelation           = errors.New("invalid relation")
	ErrInvalidAge                = errors.New("invalid age")
	ErrInvalidBudget             = errors.New("invalid budget")
	ErrInvalidInterestTag        = errors.New("invalid interest tag")
)
