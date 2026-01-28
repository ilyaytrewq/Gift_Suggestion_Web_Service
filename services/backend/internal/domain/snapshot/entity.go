package snapshot

import (
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/shared"
)

type Snapshot struct {
	Occasion     shared.Occasion
	Relation     shared.Relation
	Budget       shared.Money
	InterestTags []shared.TagID
	Age          shared.AgeLimit
}

func NewSnapshot(occasion shared.Occasion, relation shared.Relation, budget shared.Money, interestTags []shared.TagID, age shared.AgeLimit) (*Snapshot, error) {
	if isBlank(string(occasion)) {
		return nil, ErrInvalidOccasion
	}
	if isBlank(string(relation)) {
		return nil, ErrInvalidRelation
	}
	if !budget.IsNonNegative() {
		return nil, ErrInvalidBudget
	}
	if !age.IsValid() {
		return nil, ErrInvalidAge
	}
	for _, tag := range interestTags {
		if isBlank(string(tag)) {
			return nil, ErrInvalidTag
		}
	}

	return &Snapshot{
		Occasion:     occasion,
		Relation:     relation,
		Budget:       budget,
		InterestTags: shared.UniqTags(interestTags),
		Age:          age,
	}, nil
}
