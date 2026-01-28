package snapshot

import "errors"

var (
	ErrInvalidOccasion = errors.New("invalid occasion")
	ErrInvalidRelation = errors.New("invalid relation")
	ErrInvalidBudget   = errors.New("invalid budget")
	ErrInvalidAge      = errors.New("invalid age")
	ErrInvalidTag      = errors.New("invalid interest tag")
)
