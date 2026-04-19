package usecase

import (
	"sort"
	"time"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	recommendationdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/recommendation/domain"
)

type RecommendInput struct {
	UserID               string
	Occasion             string
	Relationship         string
	RecipientAge         *int
	BudgetMax            string
	PreferredCategoryIDs []string
	Interests            []string
	TopN                 int
	UseWishlistContext   *bool
}

type GetRecommendationInput struct {
	UserID    string
	RequestID string
}

type Explanation struct {
	Code string `json:"code"`
	Text string `json:"text"`
}

type GiftCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type GiftPreview struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	Price          string        `json:"price"`
	StoreLink      string        `json:"store_link"`
	Image          *string       `json:"image,omitempty"`
	AgeRestriction *int          `json:"age_restriction,omitempty"`
	Category       *GiftCategory `json:"category,omitempty"`
}

type RecommendationAlternative struct {
	Gift         GiftPreview   `json:"gift"`
	Source       string        `json:"source"`
	Score        *float64      `json:"score,omitempty"`
	Explanations []Explanation `json:"explanations"`
}

type RecommendationItem struct {
	Rank         int                         `json:"rank"`
	Gift         GiftPreview                 `json:"gift"`
	Source       string                      `json:"source"`
	Score        *float64                    `json:"score,omitempty"`
	Explanations []Explanation               `json:"explanations"`
	Alternatives []RecommendationAlternative `json:"alternatives"`
}

type RecommendationSet struct {
	RequestID                   string               `json:"request_id"`
	Status                      string               `json:"status"`
	Source                      string               `json:"source"`
	RequestedTopN               int                  `json:"requested_top_n"`
	CandidateCountBeforeFilters int                  `json:"candidate_count_before_filters"`
	CandidateCountAfterFilters  int                  `json:"candidate_count_after_filters"`
	FallbackUsed                bool                 `json:"fallback_used"`
	FallbackReason              *string              `json:"fallback_reason,omitempty"`
	GeneratedAt                 time.Time            `json:"generated_at"`
	Recommendations             []RecommendationItem `json:"recommendations"`
}

type RecommendOutput struct {
	Recommendation RecommendationSet `json:"recommendation"`
}

type GetRecommendationOutput struct {
	Recommendation RecommendationSet `json:"recommendation"`
}

func newRecommendationSet(
	request recommendationdomain.RecommendationRequest,
	results []recommendationdomain.RecommendationResult,
	gifts map[string]catalogdomain.Gift,
) RecommendationSet {
	grouped := make(map[int][]recommendationdomain.RecommendationResult)
	positions := make([]int, 0)
	for _, result := range results {
		slot := result.SlotPosition()
		if _, ok := grouped[slot]; !ok {
			positions = append(positions, slot)
		}

		grouped[slot] = append(grouped[slot], result)
	}
	sort.Ints(positions)

	items := make([]RecommendationItem, 0, len(positions))
	for _, slot := range positions {
		slotResults := grouped[slot]
		var primary *recommendationdomain.RecommendationResult
		alternatives := make([]recommendationdomain.RecommendationResult, 0)
		for _, result := range slotResults {
			if result.ResultKind() == recommendationdomain.ResultKindPrimary {
				current := result
				primary = &current
				continue
			}

			alternatives = append(alternatives, result)
		}
		if primary == nil {
			continue
		}

		gift, ok := gifts[primary.GiftID().String()]
		if !ok {
			continue
		}

		sort.SliceStable(alternatives, func(i, j int) bool {
			left := alternatives[i].AlternativeRank()
			right := alternatives[j].AlternativeRank()
			if left == nil {
				return false
			}
			if right == nil {
				return true
			}
			return *left < *right
		})

		itemAlternatives := make([]RecommendationAlternative, 0, len(alternatives))
		for _, alternative := range alternatives {
			alternativeGift, ok := gifts[alternative.GiftID().String()]
			if !ok {
				continue
			}

			itemAlternatives = append(itemAlternatives, RecommendationAlternative{
				Gift:         newGiftPreview(alternativeGift),
				Source:       string(alternative.RankingSource()),
				Score:        alternative.Score(),
				Explanations: newExplanations(alternative.Explanations()),
			})
		}

		items = append(items, RecommendationItem{
			Rank:         slot,
			Gift:         newGiftPreview(gift),
			Source:       string(primary.RankingSource()),
			Score:        primary.Score(),
			Explanations: newExplanations(primary.Explanations()),
			Alternatives: itemAlternatives,
		})
	}

	generatedAt := request.UpdatedAt()
	if finishedAt := request.FinishedAt(); finishedAt != nil {
		generatedAt = *finishedAt
	}

	source := string(request.RankingSource())
	if request.Status() == recommendationdomain.StatusCompletedEmpty {
		source = "empty"
	}

	return RecommendationSet{
		RequestID:                   request.ID().String(),
		Status:                      string(request.Status()),
		Source:                      source,
		RequestedTopN:               request.Questionnaire().TopN(),
		CandidateCountBeforeFilters: request.CandidateCountBeforeFilters(),
		CandidateCountAfterFilters:  request.CandidateCountAfterFilters(),
		FallbackUsed:                request.Status() == recommendationdomain.StatusCompletedWithFallback,
		FallbackReason:              request.FallbackReasonCode(),
		GeneratedAt:                 generatedAt,
		Recommendations:             items,
	}
}

func uniqueGiftIDs(results []recommendationdomain.RecommendationResult) []catalogdomain.GiftID {
	seen := make(map[string]struct{}, len(results))
	ids := make([]catalogdomain.GiftID, 0, len(results))
	for _, result := range results {
		key := result.GiftID().String()
		if _, exists := seen[key]; exists {
			continue
		}

		seen[key] = struct{}{}
		ids = append(ids, result.GiftID())
	}

	return ids
}

func newExplanations(items []recommendationdomain.Explanation) []Explanation {
	output := make([]Explanation, 0, len(items))
	for _, item := range items {
		output = append(output, Explanation{
			Code: item.Code(),
			Text: item.Text(),
		})
	}

	return output
}

func newGiftPreview(gift catalogdomain.Gift) GiftPreview {
	var ageRestriction *int
	if gift.AgeRestriction() != nil {
		value := gift.AgeRestriction().Int()
		ageRestriction = &value
	}

	var category *GiftCategory
	if gift.CategoryID() != nil {
		category = &GiftCategory{
			ID:   gift.CategoryID().String(),
			Name: gift.CategoryName(),
		}
	}

	return GiftPreview{
		ID:             gift.ID().String(),
		Name:           gift.Name(),
		Description:    gift.Description(),
		Price:          gift.Price().DecimalString(),
		StoreLink:      gift.StoreLink(),
		Image:          gift.Image(),
		AgeRestriction: ageRestriction,
		Category:       category,
	}
}
