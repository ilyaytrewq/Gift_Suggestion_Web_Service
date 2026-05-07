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
	userusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

const (
	testUserHandlerUserID    = "550e8400-e29b-41d4-a716-446655440000"
	testUserHandlerSessionID = "550e8400-e29b-41d4-a716-446655440001"
	testUserHandlerEmail     = "user@example.com"
)

func TestHandlerGetCurrentUserRequiresAuthorization(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubUserService{}, authhttp.NewMiddleware(stubAuthorizer{}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestHandlerUpdateProfileRejectsEmptyPatch(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	service := &stubUserService{}
	handler, err := NewHandler(service, authhttp.NewMiddleware(stubAuthorizer{
		actor: authusecase.Actor{UserID: testUserHandlerUserID, SessionID: testUserHandlerSessionID, Role: "user"},
	}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me", bytes.NewBufferString(`{"profile":{}}`))
	req.Header.Set("Authorization", "Bearer access-token")
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
	if service.updateCalls != 0 {
		t.Fatalf("expected no UpdateProfile() calls, got %d", service.updateCalls)
	}

	var response userHandlerErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if response.Error.Code != "invalid_profile_update" {
		t.Fatalf("error code = %q, want %q", response.Error.Code, "invalid_profile_update")
	}
}

func TestHandlerUpdateProfileSuccess(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	service := &stubUserService{
		updateOutput: userusecase.Profile{
			ID:          testUserHandlerUserID,
			Email:       testUserHandlerEmail,
			DisplayName: "Alice",
			UpdatedAt:   time.Date(2026, 4, 18, 17, 0, 0, 0, time.UTC),
		},
	}
	handler, err := NewHandler(service, authhttp.NewMiddleware(stubAuthorizer{
		actor: authusecase.Actor{UserID: testUserHandlerUserID, SessionID: testUserHandlerSessionID, Role: "user"},
	}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me", bytes.NewBufferString(`{"profile":{"display_name":"Alice"}}`))
	req.Header.Set("Authorization", "Bearer access-token")
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if service.lastUpdateInput.UserID != testUserHandlerUserID {
		t.Fatalf("UpdateProfile() user id = %q, want %q", service.lastUpdateInput.UserID, testUserHandlerUserID)
	}
	if service.lastUpdateInput.DisplayName != "Alice" {
		t.Fatalf("UpdateProfile() display name = %q, want %q", service.lastUpdateInput.DisplayName, "Alice")
	}
}

type stubUserService struct {
	getOutput        userusecase.Profile
	updateOutput     userusecase.Profile
	updateErr        error
	lastUpdateInput  userusecase.UpdateProfileInput
	updateCalls      int
	promoteOutput    userusecase.Profile
	promoteErr       error
	lastPromoteEmail string
	promoteCalls     int
}

func (s *stubUserService) GetCurrentUser(context.Context, string) (userusecase.Profile, error) {
	return s.getOutput, nil
}

func (s *stubUserService) UpdateProfile(_ context.Context, input userusecase.UpdateProfileInput) (userusecase.Profile, error) {
	s.lastUpdateInput = input
	s.updateCalls++
	if s.updateErr != nil {
		return userusecase.Profile{}, s.updateErr
	}
	return s.updateOutput, nil
}

func (s *stubUserService) PromoteUserToAdmin(_ context.Context, email string) (userusecase.Profile, error) {
	s.lastPromoteEmail = email
	s.promoteCalls++
	if s.promoteErr != nil {
		return userusecase.Profile{}, s.promoteErr
	}
	return s.promoteOutput, nil
}

type stubAuthorizer struct {
	actor authusecase.Actor
	err   error
}

func (a stubAuthorizer) Authorize(context.Context, string) (authusecase.Actor, error) {
	if a.err != nil {
		return authusecase.Actor{}, a.err
	}
	return a.actor, nil
}

type userHandlerErrorResponse struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

func TestHandlerGetCurrentUserPropagatesServiceErrors(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	service := &stubUserService{
		updateErr: apperrors.New(apperrors.KindForbidden, "forbidden", "forbidden"),
	}
	handler, err := NewHandler(service, authhttp.NewMiddleware(stubAuthorizer{
		actor: authusecase.Actor{UserID: testUserHandlerUserID, SessionID: testUserHandlerSessionID, Role: "user"},
	}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me", bytes.NewBufferString(`{"profile":{"display_name":"Alice"}}`))
	req.Header.Set("Authorization", "Bearer access-token")
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
}

func TestHandlerPromoteToAdminForbiddenForNonAdminActor(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubUserService{}, authhttp.NewMiddleware(stubAuthorizer{
		actor: authusecase.Actor{UserID: testUserHandlerUserID, SessionID: testUserHandlerSessionID, Role: "user"},
	}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/promote", bytes.NewBufferString(`{"email":"alice@example.com"}`))
	req.Header.Set("Authorization", "Bearer access-token")
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
}

func TestHandlerPromoteToAdminSuccess(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	service := &stubUserService{
		promoteOutput: userusecase.Profile{
			ID:        "661e8400-e29b-41d4-a716-446655440001",
			Email:     "alice@example.com",
			Role:      "admin",
			UpdatedAt: time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC),
		},
	}
	handler, err := NewHandler(service, authhttp.NewMiddleware(stubAuthorizer{
		actor: authusecase.Actor{UserID: testUserHandlerUserID, SessionID: testUserHandlerSessionID, Role: "admin"},
	}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/promote", bytes.NewBufferString(`{"email":"alice@example.com"}`))
	req.Header.Set("Authorization", "Bearer access-token")
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if service.lastPromoteEmail != "alice@example.com" {
		t.Fatalf("PromoteUserToAdmin email = %q", service.lastPromoteEmail)
	}
	if service.promoteCalls != 1 {
		t.Fatalf("promote calls = %d, want 1", service.promoteCalls)
	}
}
