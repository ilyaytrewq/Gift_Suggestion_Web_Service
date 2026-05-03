package domain

import (
	"strings"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
)

const (
	maxTextLength          = 120
	maxInterestLength      = 80
	maxPreferredCategories = 10
	maxInterests           = 10
)

type Questionnaire struct {
	occasion             string
	relationship         string
	recipientAge         *int
	recipientGender      *string
	budgetMax            catalogdomain.Price
	preferredCategoryIDs []catalogdomain.CategoryID
	interests            []string
	topN                 int
	useWishlistContext   bool
}

var validGenders = map[string]struct{}{"male": {}, "female": {}, "other": {}}

func NewQuestionnaire(
	occasion string,
	relationship string,
	recipientAge *int,
	recipientGender *string,
	budgetMax string,
	preferredCategoryIDs []string,
	interests []string,
	topN int,
	useWishlistContext bool,
) (Questionnaire, error) {
	normalizedOccasion := strings.TrimSpace(occasion)
	if len([]rune(normalizedOccasion)) > maxTextLength {
		return Questionnaire{}, ErrOccasionTooLong
	}

	normalizedRelationship := strings.TrimSpace(relationship)
	if len([]rune(normalizedRelationship)) > maxTextLength {
		return Questionnaire{}, ErrRelationshipTooLong
	}

	if topN < 1 {
		return Questionnaire{}, ErrTopNInvalid
	}

	price, err := catalogdomain.NewPrice(budgetMax)
	if err != nil {
		return Questionnaire{}, err
	}

	var normalizedRecipientAge *int
	if recipientAge != nil {
		if *recipientAge < 0 || *recipientAge > 120 {
			return Questionnaire{}, ErrRecipientAgeInvalid
		}

		value := *recipientAge
		normalizedRecipientAge = &value
	}

	var normalizedGender *string
	if recipientGender != nil {
		trimmed := strings.ToLower(strings.TrimSpace(*recipientGender))
		if trimmed != "" {
			if _, ok := validGenders[trimmed]; !ok {
				return Questionnaire{}, ErrInvalidGender
			}
			normalizedGender = &trimmed
		}
	}

	if len(preferredCategoryIDs) > maxPreferredCategories {
		return Questionnaire{}, ErrTooManyPreferredCategories
	}

	categorySet := make(map[string]struct{}, len(preferredCategoryIDs))
	categoryIDs := make([]catalogdomain.CategoryID, 0, len(preferredCategoryIDs))
	for _, raw := range preferredCategoryIDs {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}

		categoryID, err := catalogdomain.NewCategoryID(trimmed)
		if err != nil {
			return Questionnaire{}, err
		}
		if _, exists := categorySet[categoryID.String()]; exists {
			continue
		}

		categorySet[categoryID.String()] = struct{}{}
		categoryIDs = append(categoryIDs, categoryID)
	}

	if len(interests) > maxInterests {
		return Questionnaire{}, ErrTooManyInterests
	}

	interestSet := make(map[string]struct{}, len(interests))
	normalizedInterests := make([]string, 0, len(interests))
	for _, raw := range interests {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		if len([]rune(trimmed)) > maxInterestLength {
			return Questionnaire{}, ErrInterestTooLong
		}

		key := strings.ToLower(trimmed)
		if _, exists := interestSet[key]; exists {
			continue
		}

		interestSet[key] = struct{}{}
		normalizedInterests = append(normalizedInterests, trimmed)
	}

	return Questionnaire{
		occasion:             normalizedOccasion,
		relationship:         normalizedRelationship,
		recipientAge:         normalizedRecipientAge,
		recipientGender:      normalizedGender,
		budgetMax:            price,
		preferredCategoryIDs: categoryIDs,
		interests:            normalizedInterests,
		topN:                 topN,
		useWishlistContext:   useWishlistContext,
	}, nil
}

func (q Questionnaire) RecipientGender() *string {
	if q.recipientGender == nil {
		return nil
	}
	value := *q.recipientGender
	return &value
}

func (q Questionnaire) Occasion() string {
	return q.occasion
}

func (q Questionnaire) Relationship() string {
	return q.relationship
}

func (q Questionnaire) RecipientAge() *int {
	if q.recipientAge == nil {
		return nil
	}

	value := *q.recipientAge
	return &value
}

func (q Questionnaire) BudgetMax() catalogdomain.Price {
	return q.budgetMax
}

func (q Questionnaire) PreferredCategoryIDs() []catalogdomain.CategoryID {
	values := make([]catalogdomain.CategoryID, len(q.preferredCategoryIDs))
	copy(values, q.preferredCategoryIDs)
	return values
}

func (q Questionnaire) Interests() []string {
	values := make([]string, len(q.interests))
	copy(values, q.interests)
	return values
}

func (q Questionnaire) TopN() int {
	return q.topN
}

func (q Questionnaire) UseWishlistContext() bool {
	return q.useWishlistContext
}
