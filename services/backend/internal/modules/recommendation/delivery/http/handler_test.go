package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	authhttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/delivery/http"
	authusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/usecase"
	recommendationusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/recommendation/usecase"
)

const (
	testRecommendationHandlerUserID    = "550e8400-e29b-41d4-a716-446655444000"
	testRecommendationHandlerRequestID = "550e8400-e29b-41d4-a716-446655444001"
)

func TestHandlerRecommendAllowsGuestSession(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	service := &stubRecommendationService{
		recommendOutput: recommendationusecase.RecommendOutput{
			Recommendation: recommendationusecase.RecommendationSet{
				RequestID: testRecommendationHandlerRequestID,
				Status:    "completed_empty",
				Source:    "empty",
			},
		},
	}
	handler, err := NewHandler(service, authhttp.NewMiddleware(stubRecommendationAuthorizer{}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/recommendations", bytes.NewBufferString(`{"budget_max":"100.00"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if service.recommendInput.UserID != "" {
		t.Fatalf("Recommend() user id = %q, want empty for guest flow", service.recommendInput.UserID)
	}
}

func TestHandlerRecommendRejectsInvalidPayload(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubRecommendationService{}, authhttp.NewMiddleware(stubRecommendationAuthorizer{
		actor: authusecase.Actor{UserID: testRecommendationHandlerUserID, Role: "user"},
	}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/recommendations", bytes.NewBufferString(`{"budget_max":100}`))
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestHandlerRecommendSuccess(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	service := &stubRecommendationService{
		recommendOutput: recommendationusecase.RecommendOutput{
			Recommendation: recommendationusecase.RecommendationSet{
				RequestID: testRecommendationHandlerRequestID,
				Status:    "completed",
				Source:    "ml",
				Recommendations: []recommendationusecase.RecommendationItem{
					{
						Rank:   1,
						Source: "ml",
						Gift: recommendationusecase.GiftPreview{
							ID:          "550e8400-e29b-41d4-a716-446655444010",
							Name:        "Gift",
							Description: "Description",
							Price:       "99.00",
							StoreLink:   "https://example.com/gift",
						},
					},
				},
				GeneratedAt: time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC),
			},
		},
	}
	handler, err := NewHandler(service, authhttp.NewMiddleware(stubRecommendationAuthorizer{
		actor: authusecase.Actor{UserID: testRecommendationHandlerUserID, Role: "user"},
	}), stubRecommendationAuthorizer{
		actor: authusecase.Actor{UserID: testRecommendationHandlerUserID, Role: "user"},
	})
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/recommendations", bytes.NewBufferString(`{"budget_max":"100.00","top_n":1}`))
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if service.recommendInput.UserID != testRecommendationHandlerUserID {
		t.Fatalf("Recommend() user id = %q, want %q", service.recommendInput.UserID, testRecommendationHandlerUserID)
	}
	if service.recommendInput.BudgetMax != "100.00" {
		t.Fatalf("Recommend() budget = %q, want %q", service.recommendInput.BudgetMax, "100.00")
	}
}

func TestHandlerGetRecommendationRequiresAuthorization(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubRecommendationService{}, authhttp.NewMiddleware(stubRecommendationAuthorizer{}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/"+testRecommendationHandlerRequestID, nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestHandlerGetRecommendationSuccess(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubRecommendationService{
		getOutput: recommendationusecase.GetRecommendationOutput{
			Recommendation: recommendationusecase.RecommendationSet{
				RequestID:   testRecommendationHandlerRequestID,
				Status:      "completed",
				Source:      "fallback",
				GeneratedAt: time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC),
			},
		},
	}, authhttp.NewMiddleware(stubRecommendationAuthorizer{
		actor: authusecase.Actor{UserID: testRecommendationHandlerUserID, Role: "user"},
	}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/"+testRecommendationHandlerRequestID, nil)
	req.Header.Set("Authorization", "Bearer token")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var response struct {
		Data struct {
			Recommendation recommendationusecase.RecommendationSet `json:"recommendation"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if response.Data.Recommendation.RequestID != testRecommendationHandlerRequestID {
		t.Fatalf("request id = %q, want %q", response.Data.Recommendation.RequestID, testRecommendationHandlerRequestID)
	}
}

type stubRecommendationService struct {
	recommendInput  recommendationusecase.RecommendInput
	recommendOutput recommendationusecase.RecommendOutput
	recommendErr    error
	getOutput       recommendationusecase.GetRecommendationOutput
	getErr          error
}

func (s *stubRecommendationService) Recommend(_ context.Context, input recommendationusecase.RecommendInput) (recommendationusecase.RecommendOutput, error) {
	s.recommendInput = input
	return s.recommendOutput, s.recommendErr
}

func (s *stubRecommendationService) GetRecommendation(_ context.Context, input recommendationusecase.GetRecommendationInput) (recommendationusecase.GetRecommendationOutput, error) {
	return s.getOutput, s.getErr
}

type stubRecommendationAuthorizer struct {
	actor authusecase.Actor
	err   error
}

func (a stubRecommendationAuthorizer) Authorize(context.Context, string) (authusecase.Actor, error) {
	if a.err != nil {
		return authusecase.Actor{}, a.err
	}

	return a.actor, nil
}
