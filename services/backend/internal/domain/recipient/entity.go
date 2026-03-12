package recipient

import (
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/shared"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/user"
)

type (
	RecipientProfileID string
	RecipientProfile   struct {
		RecipientID  RecipientProfileID
		OwnerUserID  *user.UserID
		Occasion     *shared.Occasion
		Relation     *shared.Relation
		Age          shared.AgeLimit
		Budget       shared.Money
		InterestTags []shared.TagID
	}
)

func NewRecipientProfile(id RecipientProfileID, ownerUserID *user.UserID, occasion *shared.Occasion, relation *shared.Relation, age shared.AgeLimit, budget shared.Money, interestTags []shared.TagID) (*RecipientProfile, error) {
	if isBlank(string(id)) {
		return nil, ErrRecipientProfileIDEmpty
	}
	if !id.IsValid() {
		return nil, ErrInvalidRecipientProfileID
	}
	if ownerUserID != nil && !ownerUserID.IsValid() {
		return nil, ErrInvalidOwnerUserID
	}
	if occasion != nil && !isValidOccasion(occasion) {
		return nil, ErrInvalidOccasion
	}
	if relation != nil && !isValidRelation(relation) {
		return nil, ErrInvalidRelation
	}
	if !age.IsValid() {
		return nil, ErrInvalidAge
	}
	if !budget.IsNonNegative() {
		return nil, ErrInvalidBudget
	}
	for _, tag := range interestTags {
		if !isValidInterestTag(tag) {
			return nil, ErrInvalidInterestTag
		}
	}

	normalizedTags := shared.UniqTags(interestTags)

	return &RecipientProfile{
		RecipientID:  id,
		OwnerUserID:  ownerUserID,
		Occasion:     occasion,
		Relation:     relation,
		Age:          age,
		Budget:       budget,
		InterestTags: normalizedTags,
	}, nil
}
