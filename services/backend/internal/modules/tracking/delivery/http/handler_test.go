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
	trackingusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/tracking/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

const testTrackingHandlerUserID = "550e8400-e29b-41d4-a716-446655446000"

func TestHandlerTrackEventRequiresAuthorization(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubTrackingService{}, authhttp.NewMiddleware(stubTrackingAuthorizer{}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tracking/events", bytes.NewBufferString(`{"type":"card_view"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestHandlerTrackEventRejectsInvalidPayload(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubTrackingService{}, authhttp.NewMiddleware(stubTrackingAuthorizer{
		actor: authusecase.Actor{UserID: testTrackingHandlerUserID, Role: "user"},
	}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tracking/events", bytes.NewBufferString(`{"type":1}`))
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestHandlerTrackEventSuccess(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	service := &stubTrackingService{
		output: trackingusecase.TrackEventOutput{
			Event: trackingusecase.Event{
				ID:         "550e8400-e29b-41d4-a716-446655446001",
				Type:       "card_view",
				OccurredAt: time.Date(2026, 4, 19, 14, 0, 0, 0, time.UTC),
				RecordedAt: time.Date(2026, 4, 19, 14, 0, 0, 0, time.UTC),
			},
		},
	}
	handler, err := NewHandler(service, authhttp.NewMiddleware(stubTrackingAuthorizer{
		actor: authusecase.Actor{UserID: testTrackingHandlerUserID, Role: "user"},
	}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tracking/events", bytes.NewBufferString(`{"type":"card_view","gift_id":"550e8400-e29b-41d4-a716-446655446010","metadata":{"surface":"catalog","position":1}}`))
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}
	if service.input.UserID != testTrackingHandlerUserID {
		t.Fatalf("TrackEvent() user id = %q, want %q", service.input.UserID, testTrackingHandlerUserID)
	}
	if service.input.Type != "card_view" {
		t.Fatalf("TrackEvent() type = %q, want %q", service.input.Type, "card_view")
	}
}

func TestHandlerTrackEventReturnsOKOnDuplicate(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	service := &stubTrackingService{
		output: trackingusecase.TrackEventOutput{
			Event: trackingusecase.Event{
				ID:         "550e8400-e29b-41d4-a716-446655446002",
				Type:       "card_view",
				Duplicate:  true,
				OccurredAt: time.Date(2026, 4, 19, 14, 0, 0, 0, time.UTC),
				RecordedAt: time.Date(2026, 4, 19, 14, 0, 0, 0, time.UTC),
			},
		},
	}
	handler, err := NewHandler(service, authhttp.NewMiddleware(stubTrackingAuthorizer{
		actor: authusecase.Actor{UserID: testTrackingHandlerUserID, Role: "user"},
	}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tracking/events", bytes.NewBufferString(`{"type":"card_view","gift_id":"550e8400-e29b-41d4-a716-446655446010"}`))
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

type stubTrackingService struct {
	input  trackingusecase.TrackEventInput
	output trackingusecase.TrackEventOutput
	err    error
}

func (s *stubTrackingService) TrackEvent(_ context.Context, input trackingusecase.TrackEventInput) (trackingusecase.TrackEventOutput, error) {
	s.input = input
	return s.output, s.err
}

type stubTrackingAuthorizer struct {
	actor authusecase.Actor
	err   error
}

func (a stubTrackingAuthorizer) Authorize(context.Context, string) (authusecase.Actor, error) {
	if a.err != nil {
		return authusecase.Actor{}, a.err
	}

	return a.actor, nil
}

var (
	_ = json.RawMessage{}
	_ = apperrors.KindValidation
)
