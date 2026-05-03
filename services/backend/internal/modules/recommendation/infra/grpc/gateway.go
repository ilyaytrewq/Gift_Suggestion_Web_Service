package grpc

import (
	"context"
	"errors"
	"math"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	recommendationdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/recommendation/domain"
	recommendationusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/recommendation/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/config"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/mlgrpc"
)

type rankClient interface {
	RankCandidates(ctx context.Context, request mlgrpc.RankRequest) (mlgrpc.RankResponse, error)
}

type Gateway struct {
	client         rankClient
	requestTimeout time.Duration
	maxRetries     int
}

func NewGateway(client rankClient, cfg config.MLConfig) *Gateway {
	return &Gateway{
		client:         client,
		requestTimeout: cfg.RequestTimeout,
		maxRetries:     cfg.MaxRetries,
	}
}

func (g *Gateway) Rank(ctx context.Context, input recommendationusecase.RankInput) (recommendationusecase.RankOutput, error) {
	request := mapRankRequest(input)

	var lastErr error
	for attempt := 0; attempt <= g.maxRetries; attempt++ {
		callCtx := ctx
		cancel := func() {}
		if g.requestTimeout > 0 {
			callCtx, cancel = context.WithTimeout(ctx, g.requestTimeout)
		}

		response, err := g.client.RankCandidates(callCtx, request)
		cancel()
		if err == nil {
			return mapRankResponse(input, response)
		}
		if errors.Is(err, context.Canceled) && ctx.Err() != nil {
			return recommendationusecase.RankOutput{}, err
		}

		lastErr = err
		if !shouldRetry(err) || attempt == g.maxRetries {
			break
		}
	}

	return recommendationusecase.RankOutput{}, mapRankError(lastErr)
}

func mapRankRequest(input recommendationusecase.RankInput) mlgrpc.RankRequest {
	request := mlgrpc.RankRequest{
		SelectionID: input.RequestID,
		UserID:      input.UserID,
		TopN:        input.TopN,
		Query: mlgrpc.QueryContext{
			Occasion:        input.Occasion,
			Relationship:    input.Relationship,
			Interests:       append([]string(nil), input.Interests...),
			BudgetCents:     input.BudgetMax.Cents(),
			RecipientAge:    cloneIntPtr(input.RecipientAge),
			RecipientGender: derefStringPtr(input.RecipientGender),
		},
		Candidates: make([]mlgrpc.Candidate, 0, len(input.Candidates)),
	}

	for _, candidate := range input.Candidates {
		var categoryID *string
		if candidate.Gift.CategoryID() != nil {
			value := candidate.Gift.CategoryID().String()
			categoryID = &value
		}

		var ageRestriction *int
		if candidate.Gift.AgeRestriction() != nil {
			value := candidate.Gift.AgeRestriction().Int()
			ageRestriction = &value
		}

		request.Candidates = append(request.Candidates, mlgrpc.Candidate{
			ID:             candidate.Gift.ID().String(),
			CategoryID:     categoryID,
			CategoryName:   candidate.Gift.CategoryName(),
			PriceCents:     candidate.Gift.Price().Cents(),
			AgeRestriction: ageRestriction,
			Title:          candidate.Gift.Name(),
			Description:    candidate.Gift.Description(),
		})
	}

	return request
}

func mapRankResponse(
	input recommendationusecase.RankInput,
	response mlgrpc.RankResponse,
) (recommendationusecase.RankOutput, error) {
	candidateSet := make(map[string]struct{}, len(input.Candidates))
	for _, candidate := range input.Candidates {
		candidateSet[candidate.Gift.ID().String()] = struct{}{}
	}

	items := make([]recommendationusecase.RankedItem, 0, len(response.Items))
	seen := make(map[string]struct{}, len(response.Items))
	for _, item := range response.Items {
		if _, ok := candidateSet[item.CandidateID]; !ok {
			continue
		}
		if _, exists := seen[item.CandidateID]; exists {
			continue
		}
		if math.IsNaN(item.Score) || math.IsInf(item.Score, 0) {
			continue
		}

		giftID, err := catalogdomain.NewGiftID(item.CandidateID)
		if err != nil {
			continue
		}

		explanations := make([]recommendationdomain.Explanation, 0, len(item.Explanations))
		for _, explanation := range item.Explanations {
			if explanation.Code == "" || explanation.Text == "" {
				continue
			}

			value, err := recommendationdomain.NewExplanation(explanation.Code, explanation.Text)
			if err != nil {
				continue
			}

			explanations = append(explanations, value)
		}

		alternativeIDs := make([]catalogdomain.GiftID, 0, len(item.AlternativeCandidateIDs))
		seenAlternatives := make(map[string]struct{}, len(item.AlternativeCandidateIDs))
		for _, raw := range item.AlternativeCandidateIDs {
			if raw == item.CandidateID {
				continue
			}
			if _, ok := candidateSet[raw]; !ok {
				continue
			}
			if _, exists := seenAlternatives[raw]; exists {
				continue
			}

			alternativeID, err := catalogdomain.NewGiftID(raw)
			if err != nil {
				continue
			}

			seenAlternatives[raw] = struct{}{}
			alternativeIDs = append(alternativeIDs, alternativeID)
		}

		items = append(items, recommendationusecase.RankedItem{
			GiftID:             giftID,
			Score:              item.Score,
			Explanations:       explanations,
			AlternativeGiftIDs: alternativeIDs,
		})
		seen[item.CandidateID] = struct{}{}
	}

	if len(items) == 0 {
		return recommendationusecase.RankOutput{}, recommendationusecase.ErrInvalidRankingResponse
	}

	return recommendationusecase.RankOutput{
		Items:        items,
		ModelVersion: response.ModelVersion,
	}, nil
}

func mapRankError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, mlgrpc.ErrRankingNotImplemented):
		return recommendationusecase.ErrRankingNotImplemented
	case errors.Is(err, context.DeadlineExceeded):
		return recommendationusecase.ErrRankingTimedOut
	case status.Code(err) == codes.DeadlineExceeded:
		return recommendationusecase.ErrRankingTimedOut
	case status.Code(err) == codes.Unavailable:
		return recommendationusecase.ErrRankingUnavailable
	default:
		return recommendationusecase.ErrRankingUnavailable
	}
}

func shouldRetry(err error) bool {
	return status.Code(err) == codes.Unavailable
}

func cloneIntPtr(value *int) *int {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}

func derefStringPtr(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
