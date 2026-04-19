package domain

import (
	"math"
	"strings"
	"time"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
)

type ResultKind string

const (
	ResultKindPrimary     ResultKind = "primary"
	ResultKindAlternative ResultKind = "alternative"
)

type Explanation struct {
	code string
	text string
}

type RecommendationResult struct {
	id              ResultID
	requestID       RequestID
	giftID          catalogdomain.GiftID
	slotPosition    int
	resultKind      ResultKind
	alternativeRank *int
	rankingSource   RankingSource
	score           *float64
	explanations    []Explanation
	createdAt       time.Time
}

func NewExplanation(code, text string) (Explanation, error) {
	if strings.TrimSpace(code) == "" {
		return Explanation{}, ErrExplanationCodeEmpty
	}
	if strings.TrimSpace(text) == "" {
		return Explanation{}, ErrExplanationTextEmpty
	}

	return Explanation{
		code: strings.TrimSpace(code),
		text: strings.TrimSpace(text),
	}, nil
}

func RestoreRecommendationResult(
	id string,
	requestID string,
	giftID string,
	slotPosition int,
	resultKind string,
	alternativeRank *int,
	rankingSource string,
	score *float64,
	explanations []Explanation,
	createdAt time.Time,
) (RecommendationResult, error) {
	resultID, err := NewResultID(id)
	if err != nil {
		return RecommendationResult{}, err
	}

	reqID, err := NewRequestID(requestID)
	if err != nil {
		return RecommendationResult{}, err
	}

	catalogGiftID, err := catalogdomain.NewGiftID(giftID)
	if err != nil {
		return RecommendationResult{}, err
	}

	kind, err := parseResultKind(resultKind)
	if err != nil {
		return RecommendationResult{}, err
	}

	source, err := parseRankingSource(rankingSource)
	if err != nil {
		return RecommendationResult{}, err
	}

	return newRecommendationResult(
		resultID,
		reqID,
		catalogGiftID,
		slotPosition,
		kind,
		alternativeRank,
		source,
		score,
		explanations,
		createdAt,
	)
}

func NewPrimaryResult(
	id ResultID,
	requestID RequestID,
	giftID catalogdomain.GiftID,
	slotPosition int,
	rankingSource RankingSource,
	score *float64,
	explanations []Explanation,
	createdAt time.Time,
) (RecommendationResult, error) {
	return newRecommendationResult(
		id,
		requestID,
		giftID,
		slotPosition,
		ResultKindPrimary,
		nil,
		rankingSource,
		score,
		explanations,
		createdAt,
	)
}

func NewAlternativeResult(
	id ResultID,
	requestID RequestID,
	giftID catalogdomain.GiftID,
	slotPosition int,
	alternativeRank int,
	rankingSource RankingSource,
	score *float64,
	explanations []Explanation,
	createdAt time.Time,
) (RecommendationResult, error) {
	return newRecommendationResult(
		id,
		requestID,
		giftID,
		slotPosition,
		ResultKindAlternative,
		&alternativeRank,
		rankingSource,
		score,
		explanations,
		createdAt,
	)
}

func newRecommendationResult(
	id ResultID,
	requestID RequestID,
	giftID catalogdomain.GiftID,
	slotPosition int,
	resultKind ResultKind,
	alternativeRank *int,
	rankingSource RankingSource,
	score *float64,
	explanations []Explanation,
	createdAt time.Time,
) (RecommendationResult, error) {
	if slotPosition < 1 {
		return RecommendationResult{}, ErrSlotPositionInvalid
	}
	if rankingSource != RankingSourceML && rankingSource != RankingSourceFallback {
		return RecommendationResult{}, ErrInvalidRankingSource
	}
	if resultKind == ResultKindAlternative && (alternativeRank == nil || *alternativeRank < 1) {
		return RecommendationResult{}, ErrAlternativeRankInvalid
	}
	if resultKind == ResultKindPrimary && alternativeRank != nil {
		return RecommendationResult{}, ErrAlternativeRankInvalid
	}
	if score != nil && (math.IsNaN(*score) || math.IsInf(*score, 0)) {
		return RecommendationResult{}, ErrInvalidRankingSource
	}

	clonedExplanations := make([]Explanation, len(explanations))
	copy(clonedExplanations, explanations)

	return RecommendationResult{
		id:              id,
		requestID:       requestID,
		giftID:          giftID,
		slotPosition:    slotPosition,
		resultKind:      resultKind,
		alternativeRank: cloneIntPtr(alternativeRank),
		rankingSource:   rankingSource,
		score:           cloneFloat64Ptr(score),
		explanations:    clonedExplanations,
		createdAt:       createdAt.UTC(),
	}, nil
}

func (r RecommendationResult) ID() ResultID {
	return r.id
}

func (r RecommendationResult) RequestID() RequestID {
	return r.requestID
}

func (r RecommendationResult) GiftID() catalogdomain.GiftID {
	return r.giftID
}

func (r RecommendationResult) SlotPosition() int {
	return r.slotPosition
}

func (r RecommendationResult) ResultKind() ResultKind {
	return r.resultKind
}

func (r RecommendationResult) AlternativeRank() *int {
	return cloneIntPtr(r.alternativeRank)
}

func (r RecommendationResult) RankingSource() RankingSource {
	return r.rankingSource
}

func (r RecommendationResult) Score() *float64 {
	return cloneFloat64Ptr(r.score)
}

func (r RecommendationResult) Explanations() []Explanation {
	values := make([]Explanation, len(r.explanations))
	copy(values, r.explanations)
	return values
}

func (r RecommendationResult) CreatedAt() time.Time {
	return r.createdAt
}

func (e Explanation) Code() string {
	return e.code
}

func (e Explanation) Text() string {
	return e.text
}

func parseResultKind(raw string) (ResultKind, error) {
	switch ResultKind(raw) {
	case ResultKindPrimary, ResultKindAlternative:
		return ResultKind(raw), nil
	default:
		return "", ErrInvalidResultKind
	}
}

func cloneIntPtr(value *int) *int {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}

func cloneFloat64Ptr(value *float64) *float64 {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}
