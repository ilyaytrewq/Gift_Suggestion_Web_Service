package recipient

import (
	"strings"

	"github.com/google/uuid"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/services/backend/internal/domain/shared"
)

func (id RecipientProfileID) IsValid() bool {
	return isValidRecipientProfileID(string(id))
}

func isValidRecipientProfileID(id string) bool {
	return uuid.Validate(id) == nil
}

func isBlank(value string) bool {
	return strings.TrimSpace(value) == ""
}

func isValidOccasion(occasion *shared.Occasion) bool {
	if occasion == nil {
		return false
	}
	return !isBlank(string(*occasion))
}

func isValidRelation(relation *shared.Relation) bool {
	if relation == nil {
		return false
	}
	return !isBlank(string(*relation))
}

func isValidInterestTag(tag shared.TagID) bool {
	return !isBlank(string(tag))
}
