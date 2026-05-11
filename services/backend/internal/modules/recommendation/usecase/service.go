package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	recommendationdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/recommendation/domain"
	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

const (
	candidateFetchLimit = 400
	maxRankCandidates   = 400
	maxTopN             = maxRankCandidates
	alternativesPerItem = 2
)

type Service struct {
	candidateReader CandidateReader
	requestRepo     RequestRepository
	userReader      UserReader
	wishlistReader  WishlistReader
	giftReader      GiftReader
	ranker          RankingGateway
	requestIDGen    RequestIDGenerator
	resultIDGen     ResultIDGenerator
	clock           Clock
}

type scoredCandidate struct {
	gift                   catalogdomain.Gift
	score                  float64
	explanations           []recommendationdomain.Explanation
	alreadyInWishlist      bool
	preferredCategoryMatch bool
	interestMatches        int
	budgetGap              int64
}

type primarySelection struct {
	candidate      scoredCandidate
	source         recommendationdomain.RankingSource
	score          *float64
	explanations   []recommendationdomain.Explanation
	alternativeIDs []catalogdomain.GiftID
}

func NewService(
	candidateReader CandidateReader,
	requestRepo RequestRepository,
	userReader UserReader,
	wishlistReader WishlistReader,
	giftReader GiftReader,
	ranker RankingGateway,
	requestIDGen RequestIDGenerator,
	resultIDGen ResultIDGenerator,
	clock Clock,
) (*Service, error) {
	switch {
	case candidateReader == nil:
		return nil, ErrNilCandidateReader
	case requestRepo == nil:
		return nil, ErrNilRequestRepository
	case userReader == nil:
		return nil, ErrNilUserReader
	case wishlistReader == nil:
		return nil, ErrNilWishlistReader
	case giftReader == nil:
		return nil, ErrNilGiftReader
	case ranker == nil:
		return nil, ErrNilRankingGateway
	case requestIDGen == nil:
		return nil, ErrNilRequestIDGenerator
	case resultIDGen == nil:
		return nil, ErrNilResultIDGenerator
	case clock == nil:
		return nil, ErrNilClock
	}

	return &Service{
		candidateReader: candidateReader,
		requestRepo:     requestRepo,
		userReader:      userReader,
		wishlistReader:  wishlistReader,
		giftReader:      giftReader,
		ranker:          ranker,
		requestIDGen:    requestIDGen,
		resultIDGen:     resultIDGen,
		clock:           clock,
	}, nil
}

func (s *Service) Recommend(ctx context.Context, input RecommendInput) (RecommendOutput, error) {
	requestedByUserID, useWishlistContext, err := s.resolveRecommendationUser(ctx, input.UserID, input.UseWishlistContext)
	if err != nil {
		return RecommendOutput{}, err
	}

	topN, err := normalizeTopN(input.TopN)
	if err != nil {
		return RecommendOutput{}, err
	}

	questionnaire, err := recommendationdomain.NewQuestionnaire(
		input.Occasion,
		input.Relationship,
		input.RecipientAge,
		input.RecipientGender,
		input.BudgetMax,
		input.PreferredCategoryIDs,
		input.Interests,
		topN,
		useWishlistContext,
	)
	if err != nil {
		return RecommendOutput{}, mapQuestionnaireError(err)
	}

	requestID, err := s.requestIDGen.NewRecommendationRequestID()
	if err != nil {
		return RecommendOutput{}, err
	}

	request := recommendationdomain.NewRecommendationRequest(requestID, requestedByUserID, questionnaire, s.clock.Now())
	if err := s.requestRepo.CreateRequest(ctx, &request); err != nil {
		return RecommendOutput{}, err
	}

	wishlistGiftIDs := make(map[string]struct{})
	if questionnaire.UseWishlistContext() && requestedByUserID != nil {
		wishlistGiftIDs, err = s.loadWishlistGiftIDs(ctx, *requestedByUserID)
		if err != nil {
			return RecommendOutput{}, s.failRequest(ctx, &request, "wishlist_context_load_failed", "failed to load wishlist context", err)
		}
	}

	candidates, totalCandidates, err := s.candidateReader.SelectCandidates(ctx, CandidateSelection{
		BudgetMax:            questionnaire.BudgetMax(),
		RecipientAge:         questionnaire.RecipientAge(),
		PreferredCategoryIDs: questionnaire.PreferredCategoryIDs(),
		Limit:                candidateFetchLimit,
	})
	if err != nil {
		return RecommendOutput{}, s.failRequest(ctx, &request, "candidate_selection_failed", "failed to load recommendation candidates", err)
	}

	fallbackCandidates := buildFallbackCandidates(questionnaire, candidates, wishlistGiftIDs)
	sortFallbackCandidates(fallbackCandidates)
	if len(fallbackCandidates) > maxRankCandidates {
		fallbackCandidates = fallbackCandidates[:maxRankCandidates]
	}

	candidateCountAfterFilters := len(fallbackCandidates)
	if candidateCountAfterFilters == 0 {
		reason := "no_candidates_after_filters"
		request.CompleteEmpty(&reason, totalCandidates, 0, s.clock.Now())
		if err := s.requestRepo.CompleteRequest(ctx, &request, nil); err != nil {
			return RecommendOutput{}, err
		}

		return RecommendOutput{
			Recommendation: newRecommendationSet(request, nil, map[string]catalogdomain.Gift{}),
		}, nil
	}

	primarySelections, alternativeSelections, rankingSource, fallbackUsed, fallbackReason := s.rankAndShapeRecommendations(
		ctx,
		request.ID().String(),
		userIDValueString(requestedByUserID),
		questionnaire,
		fallbackCandidates,
	)

	results, err := s.buildResults(request.ID(), primarySelections, alternativeSelections)
	if err != nil {
		return RecommendOutput{}, s.failRequest(ctx, &request, "recommendation_result_build_failed", "failed to build recommendation results", err)
	}

	if err := request.Complete(
		rankingSource,
		fallbackUsed,
		fallbackReason,
		totalCandidates,
		candidateCountAfterFilters,
		len(primarySelections),
		countAlternativeResults(alternativeSelections),
		s.clock.Now(),
	); err != nil {
		return RecommendOutput{}, s.failRequest(ctx, &request, "recommendation_completion_failed", "failed to finalize recommendation request", err)
	}

	if err := s.requestRepo.CompleteRequest(ctx, &request, results); err != nil {
		return RecommendOutput{}, err
	}

	return RecommendOutput{
		Recommendation: newRecommendationSet(request, results, candidateGiftMap(fallbackCandidates)),
	}, nil
}

func (s *Service) resolveRecommendationUser(
	ctx context.Context,
	rawUserID string,
	useWishlistContext *bool,
) (*userdomain.UserID, bool, error) {
	if strings.TrimSpace(rawUserID) == "" {
		return nil, false, nil
	}

	userID, err := s.ensureUserExists(ctx, rawUserID)
	if err != nil {
		return nil, false, err
	}

	return &userID, normalizeUseWishlistContext(useWishlistContext), nil
}

func (s *Service) GetRecommendation(ctx context.Context, input GetRecommendationInput) (GetRecommendationOutput, error) {
	userID, err := s.ensureUserExists(ctx, input.UserID)
	if err != nil {
		return GetRecommendationOutput{}, err
	}

	requestID, err := recommendationdomain.NewRequestID(input.RequestID)
	if err != nil {
		return GetRecommendationOutput{}, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_recommendation_request_id",
			"recommendation request id is invalid",
			err,
		)
	}

	request, err := s.requestRepo.GetRequest(ctx, requestID)
	if err != nil {
		return GetRecommendationOutput{}, err
	}
	if request == nil || request.RequestedByUserID() == nil || request.RequestedByUserID().String() != userID.String() {
		return GetRecommendationOutput{}, apperrors.New(
			apperrors.KindNotFound,
			"recommendation_request_not_found",
			"recommendation request not found",
		)
	}

	results, err := s.requestRepo.ListResults(ctx, requestID)
	if err != nil {
		return GetRecommendationOutput{}, err
	}

	gifts, err := s.loadResultGifts(ctx, results)
	if err != nil {
		return GetRecommendationOutput{}, err
	}

	return GetRecommendationOutput{
		Recommendation: newRecommendationSet(*request, results, gifts),
	}, nil
}

func (s *Service) rankAndShapeRecommendations(
	ctx context.Context,
	requestID string,
	userID string,
	questionnaire recommendationdomain.Questionnaire,
	candidates []scoredCandidate,
) (
	[]primarySelection,
	map[int][]primarySelection,
	recommendationdomain.RankingSource,
	bool,
	*string,
) {
	rankInput := buildRankInput(requestID, userID, questionnaire, candidates)
	candidateByID := indexCandidatesByID(candidates)

	rankOutput, rankErr := s.ranker.Rank(ctx, rankInput)

	rankingSource := recommendationdomain.RankingSourceFallback
	fallbackUsed := false
	var fallbackReason *string

	primary := make([]primarySelection, 0, questionnaire.TopN())
	usedPrimary := make(map[string]struct{}, questionnaire.TopN())
	if rankErr == nil {
		rankingSource = recommendationdomain.RankingSourceML
		primary, usedPrimary = buildMLPrimarySelections(rankOutput, candidateByID, questionnaire.TopN())
	} else {
		reason := fallbackReasonForRankingError(rankErr)
		fallbackReason = &reason
		fallbackUsed = true
	}

	primary, fallbackUsed, fallbackReason = appendFallbackPrimaries(
		primary,
		usedPrimary,
		candidates,
		questionnaire.TopN(),
		fallbackUsed,
		fallbackReason,
		rankErr == nil,
	)

	if len(primary) == 0 {
		return nil, nil, recommendationdomain.RankingSourceFallback, true, fallbackReason
	}

	alternatives, fallbackUsed, fallbackReason := buildAlternativeSelections(
		primary,
		candidateByID,
		candidates,
		usedPrimary,
		fallbackUsed,
		fallbackReason,
		rankErr == nil,
	)

	return primary, alternatives, rankingSource, fallbackUsed, fallbackReason
}

func (s *Service) buildResults(
	requestID recommendationdomain.RequestID,
	primary []primarySelection,
	alternatives map[int][]primarySelection,
) ([]recommendationdomain.RecommendationResult, error) {
	results := make([]recommendationdomain.RecommendationResult, 0, len(primary)*(alternativesPerItem+1))
	now := s.clock.Now()

	for index, item := range primary {
		slot := index + 1

		resultID, err := s.resultIDGen.NewRecommendationResultID()
		if err != nil {
			return nil, err
		}

		result, err := recommendationdomain.NewPrimaryResult(
			resultID,
			requestID,
			item.candidate.gift.ID(),
			slot,
			item.source,
			item.score,
			item.explanations,
			now,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, result)

		slotAlternatives := alternatives[slot]
		for alternativeIndex, alternative := range slotAlternatives {
			alternativeID, err := s.resultIDGen.NewRecommendationResultID()
			if err != nil {
				return nil, err
			}

			alternativeResult, err := recommendationdomain.NewAlternativeResult(
				alternativeID,
				requestID,
				alternative.candidate.gift.ID(),
				slot,
				alternativeIndex+1,
				alternative.source,
				alternative.score,
				alternative.explanations,
				now,
			)
			if err != nil {
				return nil, err
			}

			results = append(results, alternativeResult)
		}
	}

	return results, nil
}

func (s *Service) ensureUserExists(ctx context.Context, rawUserID string) (userdomain.UserID, error) {
	userID, err := userdomain.NewUserID(rawUserID)
	if err != nil {
		return userdomain.UserID{}, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_user_id",
			"user id is invalid",
			err,
		)
	}

	user, err := s.userReader.GetByID(ctx, userID)
	if err != nil {
		return userdomain.UserID{}, err
	}
	if user == nil {
		return userdomain.UserID{}, apperrors.New(
			apperrors.KindNotFound,
			"user_not_found",
			"user not found",
		)
	}

	return userID, nil
}

func (s *Service) loadWishlistGiftIDs(ctx context.Context, userID userdomain.UserID) (map[string]struct{}, error) {
	result := make(map[string]struct{})

	wishlist, err := s.wishlistReader.GetWishlistByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if wishlist == nil {
		return result, nil
	}

	items, err := s.wishlistReader.ListWishlistItems(ctx, wishlist.ID())
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		result[item.GiftID().String()] = struct{}{}
	}

	return result, nil
}

func (s *Service) loadResultGifts(
	ctx context.Context,
	results []recommendationdomain.RecommendationResult,
) (map[string]catalogdomain.Gift, error) {
	gifts := make(map[string]catalogdomain.Gift, len(results))
	for _, giftID := range uniqueGiftIDs(results) {
		gift, err := s.giftReader.GetGift(ctx, giftID)
		if err != nil {
			return nil, err
		}
		if gift == nil {
			return nil, apperrors.New(
				apperrors.KindInternal,
				"recommendation_result_inconsistent",
				"recommendation result references a missing gift",
			)
		}

		gifts[giftID.String()] = *gift
	}

	return gifts, nil
}

func (s *Service) failRequest(
	ctx context.Context,
	request *recommendationdomain.RecommendationRequest,
	code string,
	message string,
	cause error,
) error {
	request.Fail(code, message, s.clock.Now())
	if err := s.requestRepo.FailRequest(ctx, request); err != nil {
		return err
	}

	return apperrors.Wrap(
		apperrors.KindInternal,
		code,
		message,
		cause,
	)
}

func buildFallbackCandidates(
	questionnaire recommendationdomain.Questionnaire,
	candidates []catalogdomain.Gift,
	wishlistGiftIDs map[string]struct{},
) []scoredCandidate {
	result := make([]scoredCandidate, 0, len(candidates))
	for _, gift := range candidates {
		_, alreadyInWishlist := wishlistGiftIDs[gift.ID().String()]
		categoryMatch := hasPreferredCategory(questionnaire.PreferredCategoryIDs(), gift.CategoryID())
		interestMatches := countInterestMatches(questionnaire.Interests(), gift)
		budgetGap := questionnaire.BudgetMax().Cents() - gift.Price().Cents()

		score := 3.0
		if questionnaire.BudgetMax().Cents() > 0 {
			score += 1 - float64(budgetGap)/float64(questionnaire.BudgetMax().Cents())
		}
		if categoryMatch {
			score += 2.0
		}
		score += float64(interestMatches) * 0.75
		if gift.Image() != nil {
			score += 0.25
		}
		if alreadyInWishlist {
			score -= 2.0
		} else {
			score += 0.5
		}

		result = append(result, scoredCandidate{
			gift:                   gift,
			score:                  score,
			explanations:           synthesizeExplanations(questionnaire, categoryMatch, interestMatches),
			alreadyInWishlist:      alreadyInWishlist,
			preferredCategoryMatch: categoryMatch,
			interestMatches:        interestMatches,
			budgetGap:              budgetGap,
		})
	}

	return result
}

func sortFallbackCandidates(candidates []scoredCandidate) {
	sort.SliceStable(candidates, func(i, j int) bool {
		left := candidates[i]
		right := candidates[j]

		if left.score != right.score {
			return left.score > right.score
		}
		if left.alreadyInWishlist != right.alreadyInWishlist {
			return !left.alreadyInWishlist
		}
		if left.budgetGap != right.budgetGap {
			return left.budgetGap < right.budgetGap
		}
		if !left.gift.CreatedAt().Equal(right.gift.CreatedAt()) {
			return left.gift.CreatedAt().After(right.gift.CreatedAt())
		}

		return left.gift.ID().String() < right.gift.ID().String()
	})
}

func normalizeTopN(value int) (int, error) {
	if value == 0 {
		return maxTopN, nil
	}
	if value < 1 || value > maxTopN {
		return 0, apperrors.New(
			apperrors.KindValidation,
			"invalid_top_n",
			fmt.Sprintf("top_n must be between 1 and %d, or omitted for the full list", maxTopN),
		)
	}

	return value, nil
}

func normalizeUseWishlistContext(value *bool) bool {
	if value == nil {
		return true
	}

	return *value
}

func userIDValueString(value *userdomain.UserID) string {
	if value == nil {
		return ""
	}

	return value.String()
}

func mapQuestionnaireError(err error) error {
	switch {
	case errors.Is(err, catalogdomain.ErrPriceEmpty), errors.Is(err, recommendationdomain.ErrBudgetEmpty):
		return apperrors.New(
			apperrors.KindValidation,
			"budget_required",
			"budget_max is required",
		)
	case errors.Is(err, catalogdomain.ErrInvalidPrice), errors.Is(err, catalogdomain.ErrNegativePrice):
		return apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_budget",
			"budget_max is invalid",
			err,
		)
	case errors.Is(err, recommendationdomain.ErrOccasionTooLong):
		return apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_occasion",
			"occasion is invalid",
			err,
		)
	case errors.Is(err, recommendationdomain.ErrRelationshipTooLong):
		return apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_relationship",
			"relationship is invalid",
			err,
		)
	case errors.Is(err, recommendationdomain.ErrInvalidGender):
		return apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_recipient_gender",
			"recipient_gender is invalid; allowed: male, female, other",
			err,
		)
	case errors.Is(err, recommendationdomain.ErrRecipientAgeInvalid), errors.Is(err, catalogdomain.ErrInvalidAgeRestriction):
		return apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_recipient_age",
			"recipient_age is invalid",
			err,
		)
	case errors.Is(err, recommendationdomain.ErrTopNInvalid):
		return apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_top_n",
			"top_n is invalid",
			err,
		)
	case errors.Is(err, recommendationdomain.ErrTooManyPreferredCategories):
		return apperrors.Wrap(
			apperrors.KindValidation,
			"too_many_preferred_categories",
			"preferred_category_ids contains too many values",
			err,
		)
	case errors.Is(err, recommendationdomain.ErrTooManyInterests), errors.Is(err, recommendationdomain.ErrInterestTooLong):
		return apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_interests",
			"interests are invalid",
			err,
		)
	default:
		return apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_recommendation_request",
			"recommendation request is invalid",
			err,
		)
	}
}

func countAlternativeResults(items map[int][]primarySelection) int {
	total := 0
	for _, selection := range items {
		total += len(selection)
	}

	return total
}

func candidateGiftMap(candidates []scoredCandidate) map[string]catalogdomain.Gift {
	result := make(map[string]catalogdomain.Gift, len(candidates))
	for _, candidate := range candidates {
		result[candidate.gift.ID().String()] = candidate.gift
	}

	return result
}

func fallbackReasonForRankingError(err error) string {
	switch {
	case errors.Is(err, ErrRankingTimedOut):
		return "ml_timeout"
	case errors.Is(err, ErrRankingUnavailable):
		return "ml_unavailable"
	case errors.Is(err, ErrRankingNotImplemented):
		return "ml_not_implemented"
	case errors.Is(err, ErrInvalidRankingResponse):
		return "ml_response_invalid"
	default:
		return "ml_integration_error"
	}
}

func hasPreferredCategory(preferred []catalogdomain.CategoryID, categoryID *catalogdomain.CategoryID) bool {
	if categoryID == nil || len(preferred) == 0 {
		return false
	}

	for _, candidate := range preferred {
		if candidate.String() == categoryID.String() {
			return true
		}
	}

	return false
}

func countInterestMatches(interests []string, gift catalogdomain.Gift) int {
	if len(interests) == 0 {
		return 0
	}

	corpus := strings.ToLower(strings.Join([]string{
		gift.Name(),
		gift.Description(),
		gift.CategoryName(),
	}, " "))

	count := 0
	for _, interest := range interests {
		if strings.Contains(corpus, strings.ToLower(interest)) {
			count++
		}
	}

	return count
}

func synthesizeExplanations(
	questionnaire recommendationdomain.Questionnaire,
	categoryMatch bool,
	interestMatches int,
) []recommendationdomain.Explanation {
	explanations := make([]recommendationdomain.Explanation, 0, 4)
	explanations = append(explanations, mustExplanation("fits_budget", "Подходит по бюджету."))

	if questionnaire.RecipientAge() != nil {
		explanations = append(explanations, mustExplanation("age_appropriate", "Учитывает возрастное ограничение."))
	}
	if categoryMatch {
		explanations = append(explanations, mustExplanation("matches_selected_category", "Совпадает с выбранной категорией."))
	}
	if interestMatches > 0 {
		explanations = append(explanations, mustExplanation("matches_interests", "Соотносится с указанными интересами."))
	}
	if len(explanations) == 0 {
		explanations = append(explanations, mustExplanation("matches_request", "Соответствует заданным параметрам."))
	}

	return explanations
}

func mustExplanation(code, text string) recommendationdomain.Explanation {
	explanation, err := recommendationdomain.NewExplanation(code, text)
	if err != nil {
		panic(err)
	}

	return explanation
}

func cloneStructSet(source map[string]struct{}) map[string]struct{} {
	result := make(map[string]struct{}, len(source))
	for key := range source {
		result[key] = struct{}{}
	}

	return result
}

func slicesCloneGiftIDs(values []catalogdomain.GiftID) []catalogdomain.GiftID {
	result := make([]catalogdomain.GiftID, len(values))
	copy(result, values)
	return result
}

func buildRankInput(
	requestID string,
	userID string,
	questionnaire recommendationdomain.Questionnaire,
	candidates []scoredCandidate,
) RankInput {
	input := RankInput{
		RequestID:       requestID,
		UserID:          userID,
		Occasion:        questionnaire.Occasion(),
		Relationship:    questionnaire.Relationship(),
		RecipientAge:    questionnaire.RecipientAge(),
		RecipientGender: questionnaire.RecipientGender(),
		BudgetMax:       questionnaire.BudgetMax(),
		Interests:       questionnaire.Interests(),
		TopN:            questionnaire.TopN(),
		Candidates:      make([]RankCandidate, 0, len(candidates)),
	}
	for _, candidate := range candidates {
		input.Candidates = append(input.Candidates, RankCandidate{
			Gift:              candidate.gift,
			AlreadyInWishlist: candidate.alreadyInWishlist,
		})
	}

	return input
}

func indexCandidatesByID(candidates []scoredCandidate) map[string]scoredCandidate {
	index := make(map[string]scoredCandidate, len(candidates))
	for _, candidate := range candidates {
		index[candidate.gift.ID().String()] = candidate
	}

	return index
}

func buildMLPrimarySelections(
	output RankOutput,
	candidateByID map[string]scoredCandidate,
	limit int,
) ([]primarySelection, map[string]struct{}) {
	primary := make([]primarySelection, 0, limit)
	usedPrimary := make(map[string]struct{}, limit)

	for _, item := range output.Items {
		candidate, ok := candidateByID[item.GiftID.String()]
		if !ok {
			continue
		}
		if _, exists := usedPrimary[item.GiftID.String()]; exists {
			continue
		}

		explanations := item.Explanations
		if len(explanations) == 0 {
			explanations = candidate.explanations
		}

		score := item.Score
		primary = append(primary, primarySelection{
			candidate:      candidate,
			source:         recommendationdomain.RankingSourceML,
			score:          &score,
			explanations:   explanations,
			alternativeIDs: slicesCloneGiftIDs(item.AlternativeGiftIDs),
		})
		usedPrimary[item.GiftID.String()] = struct{}{}
		if len(primary) == limit {
			break
		}
	}

	return primary, usedPrimary
}

func appendFallbackPrimaries(
	primary []primarySelection,
	usedPrimary map[string]struct{},
	candidates []scoredCandidate,
	limit int,
	fallbackUsed bool,
	fallbackReason *string,
	partialML bool,
) ([]primarySelection, bool, *string) {
	for _, candidate := range candidates {
		if len(primary) == limit {
			break
		}
		if _, exists := usedPrimary[candidate.gift.ID().String()]; exists {
			continue
		}

		score := candidate.score
		primary = append(primary, primarySelection{
			candidate:    candidate,
			source:       recommendationdomain.RankingSourceFallback,
			score:        &score,
			explanations: candidate.explanations,
		})
		usedPrimary[candidate.gift.ID().String()] = struct{}{}
		fallbackUsed = true
		fallbackReason = ensurePartialFallbackReason(fallbackReason, partialML)
	}

	return primary, fallbackUsed, fallbackReason
}

func buildAlternativeSelections(
	primary []primarySelection,
	candidateByID map[string]scoredCandidate,
	candidates []scoredCandidate,
	usedPrimary map[string]struct{},
	fallbackUsed bool,
	fallbackReason *string,
	partialML bool,
) (map[int][]primarySelection, bool, *string) {
	usedResults := cloneStructSet(usedPrimary)
	alternatives := make(map[int][]primarySelection, len(primary))

	for index, item := range primary {
		slot := index + 1
		slotAlternatives, usedInSlot := buildMLAlternativeSelections(item, candidateByID, usedResults)
		for giftID := range usedInSlot {
			usedResults[giftID] = struct{}{}
		}

		for _, candidate := range candidates {
			if len(slotAlternatives) == alternativesPerItem {
				break
			}
			if _, exists := usedResults[candidate.gift.ID().String()]; exists {
				continue
			}

			score := candidate.score
			slotAlternatives = append(slotAlternatives, primarySelection{
				candidate:    candidate,
				source:       recommendationdomain.RankingSourceFallback,
				score:        &score,
				explanations: candidate.explanations,
			})
			usedResults[candidate.gift.ID().String()] = struct{}{}
			fallbackUsed = true
			fallbackReason = ensurePartialFallbackReason(fallbackReason, partialML)
		}

		alternatives[slot] = slotAlternatives
	}

	return alternatives, fallbackUsed, fallbackReason
}

func buildMLAlternativeSelections(
	item primarySelection,
	candidateByID map[string]scoredCandidate,
	usedResults map[string]struct{},
) ([]primarySelection, map[string]struct{}) {
	slotAlternatives := make([]primarySelection, 0, alternativesPerItem)
	usedInSlot := make(map[string]struct{}, alternativesPerItem)

	for _, giftID := range item.alternativeIDs {
		candidate, ok := candidateByID[giftID.String()]
		if !ok {
			continue
		}
		if _, exists := usedResults[giftID.String()]; exists {
			continue
		}

		slotAlternatives = append(slotAlternatives, primarySelection{
			candidate:    candidate,
			source:       recommendationdomain.RankingSourceML,
			score:        nil,
			explanations: candidate.explanations,
		})
		usedInSlot[giftID.String()] = struct{}{}
		if len(slotAlternatives) == alternativesPerItem {
			break
		}
	}

	return slotAlternatives, usedInSlot
}

func ensurePartialFallbackReason(fallbackReason *string, partialML bool) *string {
	if !partialML || fallbackReason != nil {
		return fallbackReason
	}

	reason := "partial_ml_response"
	return &reason
}
