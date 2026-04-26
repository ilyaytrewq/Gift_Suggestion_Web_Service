package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	recommendationdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/recommendation/domain"
	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	wishlistdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/wishlist/domain"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

const (
	testRecommendationUserID = "550e8400-e29b-41d4-a716-446655441000"
	testRecommendationReqID  = "550e8400-e29b-41d4-a716-446655441001"
)

var useWishlistContextDisabled = false

func TestServiceRecommendSuccessWithMLRanking(t *testing.T) {
	t.Parallel()

	gift1 := mustRecommendationGift(t, "550e8400-e29b-41d4-a716-446655441010", "Gift One", "Great book", "50.00")
	gift2 := mustRecommendationGift(t, "550e8400-e29b-41d4-a716-446655441011", "Gift Two", "Top pick", "70.00")

	explanation := mustExplanationForTest(t, "ml_reason", "Подходит по интересам.")
	repo := &fakeRecommendationRequestRepo{}
	ranker := &fakeRankingGateway{
		output: RankOutput{
			Items: []RankedItem{
				{
					GiftID:             gift2.ID(),
					Score:              0.91,
					Explanations:       []recommendationdomain.Explanation{explanation},
					AlternativeGiftIDs: []catalogdomain.GiftID{gift1.ID()},
				},
			},
		},
	}

	service := mustRecommendationService(t, recommendationServiceDeps{
		candidateReader: &fakeCandidateReader{items: []catalogdomain.Gift{gift1, gift2}, total: 2},
		requestRepo:     repo,
		userReader:      fakeUserReader{user: mustRecommendationUser(t)},
		wishlistReader:  &fakeWishlistReader{},
		giftReader:      fakeGiftReader{gifts: map[string]catalogdomain.Gift{gift1.ID().String(): gift1, gift2.ID().String(): gift2}},
		ranker:          ranker,
	})

	output, err := service.Recommend(context.Background(), RecommendInput{
		UserID:             testRecommendationUserID,
		BudgetMax:          "100.00",
		TopN:               1,
		UseWishlistContext: &useWishlistContextDisabled,
	})
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}

	if ranker.input.TopN != 1 {
		t.Fatalf("Rank() top_n = %d, want 1", ranker.input.TopN)
	}
	if output.Recommendation.Source != "ml" {
		t.Fatalf("Recommend() source = %q, want %q", output.Recommendation.Source, "ml")
	}
	if output.Recommendation.FallbackUsed {
		t.Fatal("Recommend() fallback_used = true, want false")
	}
	if len(output.Recommendation.Recommendations) != 1 {
		t.Fatalf("Recommend() recommendations = %d, want 1", len(output.Recommendation.Recommendations))
	}
	if output.Recommendation.Recommendations[0].Gift.ID != gift2.ID().String() {
		t.Fatalf("Recommend() first gift id = %q, want %q", output.Recommendation.Recommendations[0].Gift.ID, gift2.ID().String())
	}
	if len(output.Recommendation.Recommendations[0].Alternatives) != 1 {
		t.Fatalf("Recommend() alternatives = %d, want 1", len(output.Recommendation.Recommendations[0].Alternatives))
	}
	if output.Recommendation.Recommendations[0].Alternatives[0].Gift.ID != gift1.ID().String() {
		t.Fatalf("Recommend() alternative gift id = %q, want %q", output.Recommendation.Recommendations[0].Alternatives[0].Gift.ID, gift1.ID().String())
	}
	if repo.completed == nil || repo.completed.Status() != recommendationdomain.StatusCompleted {
		t.Fatalf("CompleteRequest() status = %v, want %v", repo.completed.Status(), recommendationdomain.StatusCompleted)
	}
}

func TestServiceRecommendAllowsGuestSession(t *testing.T) {
	t.Parallel()

	gift1 := mustRecommendationGift(t, "550e8400-e29b-41d4-a716-446655441012", "Gift One", "Great book", "50.00")
	repo := &fakeRecommendationRequestRepo{}
	wishlistReader := &fakeWishlistReader{}
	ranker := &fakeRankingGateway{
		err: ErrRankingNotImplemented,
	}

	service := mustRecommendationService(t, recommendationServiceDeps{
		candidateReader: &fakeCandidateReader{items: []catalogdomain.Gift{gift1}, total: 1},
		requestRepo:     repo,
		userReader:      fakeUserReader{user: nil},
		wishlistReader:  wishlistReader,
		giftReader:      fakeGiftReader{gifts: map[string]catalogdomain.Gift{gift1.ID().String(): gift1}},
		ranker:          ranker,
	})

	output, err := service.Recommend(context.Background(), RecommendInput{
		BudgetMax: "100.00",
		TopN:      1,
	})
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}

	if repo.created == nil {
		t.Fatal("CreateRequest() was not called")
	}
	if repo.created.RequestedByUserID() != nil {
		t.Fatalf("expected guest request without user id, got %v", repo.created.RequestedByUserID())
	}
	if ranker.input.UserID != "" {
		t.Fatalf("Rank() user id = %q, want empty for guest flow", ranker.input.UserID)
	}
	if wishlistReader.getWishlistCalls != 0 {
		t.Fatalf("expected wishlist context to be skipped for guest flow, got %d calls", wishlistReader.getWishlistCalls)
	}
	if output.Recommendation.Source != "fallback" {
		t.Fatalf("Recommend() source = %q, want %q", output.Recommendation.Source, "fallback")
	}
}

func TestServiceRecommendReturnsEmptyWhenNoCandidates(t *testing.T) {
	t.Parallel()

	repo := &fakeRecommendationRequestRepo{}
	ranker := &fakeRankingGateway{}
	service := mustRecommendationService(t, recommendationServiceDeps{
		candidateReader: &fakeCandidateReader{total: 0},
		requestRepo:     repo,
		userReader:      fakeUserReader{user: mustRecommendationUser(t)},
		wishlistReader:  &fakeWishlistReader{},
		giftReader:      fakeGiftReader{},
		ranker:          ranker,
	})

	output, err := service.Recommend(context.Background(), RecommendInput{
		UserID:             testRecommendationUserID,
		BudgetMax:          "100.00",
		UseWishlistContext: &useWishlistContextDisabled,
	})
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}

	if ranker.called {
		t.Fatal("Rank() called for empty candidate set")
	}
	if output.Recommendation.Source != "empty" {
		t.Fatalf("Recommend() source = %q, want %q", output.Recommendation.Source, "empty")
	}
	if len(output.Recommendation.Recommendations) != 0 {
		t.Fatalf("Recommend() recommendations = %d, want 0", len(output.Recommendation.Recommendations))
	}
	if repo.completed == nil || repo.completed.Status() != recommendationdomain.StatusCompletedEmpty {
		t.Fatalf("CompleteRequest() status = %v, want %v", repo.completed.Status(), recommendationdomain.StatusCompletedEmpty)
	}
}

func TestServiceRecommendFallsBackOnMLTimeout(t *testing.T) {
	t.Parallel()

	gift1 := mustRecommendationGift(t, "550e8400-e29b-41d4-a716-446655441020", "Gift One", "Great book", "50.00")
	gift2 := mustRecommendationGift(t, "550e8400-e29b-41d4-a716-446655441021", "Gift Two", "Top pick", "70.00")

	repo := &fakeRecommendationRequestRepo{}
	service := mustRecommendationService(t, recommendationServiceDeps{
		candidateReader: &fakeCandidateReader{items: []catalogdomain.Gift{gift1, gift2}, total: 2},
		requestRepo:     repo,
		userReader:      fakeUserReader{user: mustRecommendationUser(t)},
		wishlistReader:  &fakeWishlistReader{},
		giftReader:      fakeGiftReader{gifts: map[string]catalogdomain.Gift{gift1.ID().String(): gift1, gift2.ID().String(): gift2}},
		ranker:          &fakeRankingGateway{err: ErrRankingTimedOut},
	})

	output, err := service.Recommend(context.Background(), RecommendInput{
		UserID:             testRecommendationUserID,
		BudgetMax:          "100.00",
		TopN:               2,
		UseWishlistContext: &useWishlistContextDisabled,
	})
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}

	if output.Recommendation.Source != "fallback" {
		t.Fatalf("Recommend() source = %q, want %q", output.Recommendation.Source, "fallback")
	}
	if !output.Recommendation.FallbackUsed {
		t.Fatal("Recommend() fallback_used = false, want true")
	}
	if output.Recommendation.FallbackReason == nil || *output.Recommendation.FallbackReason != "ml_timeout" {
		t.Fatalf("Recommend() fallback_reason = %v, want ml_timeout", output.Recommendation.FallbackReason)
	}
	if len(output.Recommendation.Recommendations) != 2 {
		t.Fatalf("Recommend() recommendations = %d, want 2", len(output.Recommendation.Recommendations))
	}
	if repo.completed == nil || repo.completed.Status() != recommendationdomain.StatusCompletedWithFallback {
		t.Fatalf("CompleteRequest() status = %v, want %v", repo.completed.Status(), recommendationdomain.StatusCompletedWithFallback)
	}
}

func TestServiceRecommendUsesPartialMLResponseAndFallbackFill(t *testing.T) {
	t.Parallel()

	gift1 := mustRecommendationGift(t, "550e8400-e29b-41d4-a716-446655441030", "Gift One", "Great book", "50.00")
	gift2 := mustRecommendationGift(t, "550e8400-e29b-41d4-a716-446655441031", "Gift Two", "Top pick", "70.00")
	gift3 := mustRecommendationGift(t, "550e8400-e29b-41d4-a716-446655441032", "Gift Three", "Alternative", "65.00")

	service := mustRecommendationService(t, recommendationServiceDeps{
		candidateReader: &fakeCandidateReader{items: []catalogdomain.Gift{gift1, gift2, gift3}, total: 3},
		requestRepo:     &fakeRecommendationRequestRepo{},
		userReader:      fakeUserReader{user: mustRecommendationUser(t)},
		wishlistReader:  &fakeWishlistReader{},
		giftReader:      fakeGiftReader{gifts: map[string]catalogdomain.Gift{gift1.ID().String(): gift1, gift2.ID().String(): gift2, gift3.ID().String(): gift3}},
		ranker: &fakeRankingGateway{
			output: RankOutput{
				Items: []RankedItem{
					{
						GiftID: gift1.ID(),
						Score:  0.95,
					},
				},
			},
		},
	})

	output, err := service.Recommend(context.Background(), RecommendInput{
		UserID:             testRecommendationUserID,
		BudgetMax:          "100.00",
		TopN:               2,
		UseWishlistContext: &useWishlistContextDisabled,
	})
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}

	if output.Recommendation.Source != "ml" {
		t.Fatalf("Recommend() source = %q, want %q", output.Recommendation.Source, "ml")
	}
	if !output.Recommendation.FallbackUsed {
		t.Fatal("Recommend() fallback_used = false, want true")
	}
	if len(output.Recommendation.Recommendations) != 2 {
		t.Fatalf("Recommend() recommendations = %d, want 2", len(output.Recommendation.Recommendations))
	}
	if len(output.Recommendation.Recommendations[0].Explanations) == 0 {
		t.Fatal("Recommend() explanations should be synthesized for partial ML response")
	}
	if output.Recommendation.Recommendations[1].Source != "fallback" {
		t.Fatalf("Recommend() second recommendation source = %q, want %q", output.Recommendation.Recommendations[1].Source, "fallback")
	}
}

func TestServiceRecommendCapsOversizedCandidateSetBeforeML(t *testing.T) {
	t.Parallel()

	items := make([]catalogdomain.Gift, 0, 250)
	for index := range 250 {
		items = append(items, mustRecommendationGift(
			t,
			giftIDForIndex(index),
			"Gift Oversized",
			"Candidate",
			"10.00",
		))
	}

	ranker := &fakeRankingGateway{
		err: ErrRankingNotImplemented,
	}
	service := mustRecommendationService(t, recommendationServiceDeps{
		candidateReader: &fakeCandidateReader{items: items, total: len(items)},
		requestRepo:     &fakeRecommendationRequestRepo{},
		userReader:      fakeUserReader{user: mustRecommendationUser(t)},
		wishlistReader:  &fakeWishlistReader{},
		giftReader:      fakeGiftReader{},
		ranker:          ranker,
	})

	_, err := service.Recommend(context.Background(), RecommendInput{
		UserID:             testRecommendationUserID,
		BudgetMax:          "100.00",
		TopN:               1,
		UseWishlistContext: &useWishlistContextDisabled,
	})
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}

	if len(ranker.input.Candidates) != maxRankCandidates {
		t.Fatalf("Rank() candidates = %d, want %d", len(ranker.input.Candidates), maxRankCandidates)
	}
}

func TestServiceGetRecommendationHidesForeignRequest(t *testing.T) {
	t.Parallel()

	foreignUserID := "550e8400-e29b-41d4-a716-446655441040"
	foreignUser, err := userdomain.NewUserID(foreignUserID)
	if err != nil {
		t.Fatalf("NewUserID() error = %v", err)
	}

	questionnaire, err := recommendationdomain.NewQuestionnaire("", "", nil, "100.00", nil, nil, 1, false)
	if err != nil {
		t.Fatalf("NewQuestionnaire() error = %v", err)
	}
	requestID, err := recommendationdomain.NewRequestID(testRecommendationReqID)
	if err != nil {
		t.Fatalf("NewRequestID() error = %v", err)
	}
	request := recommendationdomain.NewRecommendationRequest(requestID, &foreignUser, questionnaire, time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC))

	service := mustRecommendationService(t, recommendationServiceDeps{
		candidateReader: &fakeCandidateReader{},
		requestRepo:     &fakeRecommendationRequestRepo{getRequest: &request},
		userReader:      fakeUserReader{user: mustRecommendationUser(t)},
		wishlistReader:  &fakeWishlistReader{},
		giftReader:      fakeGiftReader{},
		ranker:          &fakeRankingGateway{},
	})

	_, err = service.GetRecommendation(context.Background(), GetRecommendationInput{
		UserID:    testRecommendationUserID,
		RequestID: testRecommendationReqID,
	})
	if err == nil {
		t.Fatal("GetRecommendation() expected not found error")
	}

	appErr := apperrors.From(err)
	if appErr.Code() != "recommendation_request_not_found" {
		t.Fatalf("GetRecommendation() code = %q, want %q", appErr.Code(), "recommendation_request_not_found")
	}
}

type recommendationServiceDeps struct {
	candidateReader *fakeCandidateReader
	requestRepo     *fakeRecommendationRequestRepo
	userReader      fakeUserReader
	wishlistReader  *fakeWishlistReader
	giftReader      fakeGiftReader
	ranker          *fakeRankingGateway
}

func mustRecommendationService(t *testing.T, deps recommendationServiceDeps) *Service {
	t.Helper()

	service, err := NewService(
		deps.candidateReader,
		deps.requestRepo,
		deps.userReader,
		deps.wishlistReader,
		deps.giftReader,
		deps.ranker,
		fakeRecommendationRequestIDGenerator{id: testRecommendationReqID},
		&fakeRecommendationResultIDGenerator{},
		fakeRecommendationClock{now: time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)},
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	return service
}

type fakeCandidateReader struct {
	items []catalogdomain.Gift
	total int
	err   error
}

func (r *fakeCandidateReader) SelectCandidates(context.Context, CandidateSelection) ([]catalogdomain.Gift, int, error) {
	return r.items, r.total, r.err
}

type fakeRecommendationRequestRepo struct {
	created     *recommendationdomain.RecommendationRequest
	completed   *recommendationdomain.RecommendationRequest
	failed      *recommendationdomain.RecommendationRequest
	results     []recommendationdomain.RecommendationResult
	getRequest  *recommendationdomain.RecommendationRequest
	listResults []recommendationdomain.RecommendationResult
}

func (r *fakeRecommendationRequestRepo) CreateRequest(_ context.Context, request *recommendationdomain.RecommendationRequest) error {
	cloned := *request
	r.created = &cloned
	return nil
}

func (r *fakeRecommendationRequestRepo) CompleteRequest(_ context.Context, request *recommendationdomain.RecommendationRequest, results []recommendationdomain.RecommendationResult) error {
	cloned := *request
	r.completed = &cloned
	r.results = append([]recommendationdomain.RecommendationResult(nil), results...)
	return nil
}

func (r *fakeRecommendationRequestRepo) FailRequest(_ context.Context, request *recommendationdomain.RecommendationRequest) error {
	cloned := *request
	r.failed = &cloned
	return nil
}

func (r *fakeRecommendationRequestRepo) GetRequest(context.Context, recommendationdomain.RequestID) (*recommendationdomain.RecommendationRequest, error) {
	return r.getRequest, nil
}

func (r *fakeRecommendationRequestRepo) ListResults(context.Context, recommendationdomain.RequestID) ([]recommendationdomain.RecommendationResult, error) {
	return append([]recommendationdomain.RecommendationResult(nil), r.listResults...), nil
}

type fakeUserReader struct {
	user *userdomain.User
}

func (r fakeUserReader) GetByID(context.Context, userdomain.UserID) (*userdomain.User, error) {
	return r.user, nil
}

type fakeWishlistReader struct {
	wishlist               *wishlistdomain.Wishlist
	items                  map[string][]wishlistdomain.WishlistItem
	getWishlistCalls       int
	listWishlistItemsCalls int
}

func (r *fakeWishlistReader) GetWishlistByUserID(context.Context, userdomain.UserID) (*wishlistdomain.Wishlist, error) {
	r.getWishlistCalls++
	return r.wishlist, nil
}

func (r *fakeWishlistReader) ListWishlistItems(_ context.Context, wishlistID wishlistdomain.WishlistID) ([]wishlistdomain.WishlistItem, error) {
	r.listWishlistItemsCalls++
	if r.items == nil {
		return nil, nil
	}

	return r.items[wishlistID.String()], nil
}

type fakeGiftReader struct {
	gifts map[string]catalogdomain.Gift
}

func (r fakeGiftReader) GetGift(_ context.Context, id catalogdomain.GiftID) (*catalogdomain.Gift, error) {
	if r.gifts == nil {
		return nil, nil
	}

	gift, ok := r.gifts[id.String()]
	if !ok {
		return nil, nil
	}

	current := gift
	return &current, nil
}

type fakeRankingGateway struct {
	input  RankInput
	output RankOutput
	err    error
	called bool
}

func (g *fakeRankingGateway) Rank(_ context.Context, input RankInput) (RankOutput, error) {
	g.called = true
	g.input = input
	return g.output, g.err
}

type fakeRecommendationRequestIDGenerator struct {
	id string
}

func (g fakeRecommendationRequestIDGenerator) NewRecommendationRequestID() (recommendationdomain.RequestID, error) {
	return recommendationdomain.NewRequestID(g.id)
}

type fakeRecommendationResultIDGenerator struct {
	index int
}

func (g *fakeRecommendationResultIDGenerator) NewRecommendationResultID() (recommendationdomain.ResultID, error) {
	id := giftIDForIndex(100 + g.index)
	g.index++
	return recommendationdomain.NewResultID(id)
}

type fakeRecommendationClock struct {
	now time.Time
}

func (c fakeRecommendationClock) Now() time.Time {
	return c.now
}

func mustRecommendationUser(t *testing.T) *userdomain.User {
	t.Helper()

	user, err := userdomain.RestoreUser(
		testRecommendationUserID,
		"user@example.com",
		"$2a$10$4s1z0N9H8vK6L2S2Qz4B6OYX3j6Nn3kLBXjg5eQof8RP4eAbOeX6C",
		"user",
		"Tester",
		time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC),
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("RestoreUser() error = %v", err)
	}

	return &user
}

func mustRecommendationGift(t *testing.T, id, name, description, price string) catalogdomain.Gift {
	t.Helper()

	categoryID := "550e8400-e29b-41d4-a716-446655442000"
	categoryName := "Books"
	gift, err := catalogdomain.RestoreGift(
		id,
		&categoryID,
		&categoryName,
		name,
		description,
		price,
		"https://example.com/"+id,
		nil,
		nil,
		time.Date(2026, 4, 19, 9, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 19, 9, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("RestoreGift() error = %v", err)
	}

	return gift
}

func mustExplanationForTest(t *testing.T, code, text string) recommendationdomain.Explanation {
	t.Helper()

	explanation, err := recommendationdomain.NewExplanation(code, text)
	if err != nil {
		t.Fatalf("NewExplanation() error = %v", err)
	}

	return explanation
}

func giftIDForIndex(index int) string {
	return fmt.Sprintf("550e8400-e29b-41d4-a716-%012d", 1000+index)
}
