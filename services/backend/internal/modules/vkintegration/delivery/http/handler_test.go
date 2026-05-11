package http

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	authhttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/delivery/http"
	authusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/usecase"
	vkintegrationusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

const testVKHandlerUserID = "550e8400-e29b-41d4-a716-446655447000"

func TestHandlerConnectRequiresAuthorization(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubVKIntegrationService{}, authhttp.NewMiddleware(stubVKAuthorizer{}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPut, "/api/v1/integrations/vk/connection", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestHandlerConnectRejectsInvalidPayload(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubVKIntegrationService{}, authhttp.NewMiddleware(stubVKAuthorizer{
		actor: authusecase.Actor{UserID: testVKHandlerUserID, Role: "user"},
	}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPut, "/api/v1/integrations/vk/connection", bytes.NewBufferString(`{"provider_user_id":1}`))
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestHandlerConnectSuccess(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	service := &stubVKIntegrationService{
		connectOutput: vkintegrationusecase.ConnectOutput{
			Connection: vkintegrationusecase.Connection{
				Provider:       "vk",
				State:          "sync_required",
				FeatureEnabled: true,
			},
		},
	}
	handler, err := NewHandler(service, authhttp.NewMiddleware(stubVKAuthorizer{
		actor: authusecase.Actor{UserID: testVKHandlerUserID, Role: "user"},
	}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPut, "/api/v1/integrations/vk/connection", bytes.NewBufferString(`{"provider_user_id":"vk_1","consent":{"granted":true,"version":"v1","obtained_at":"2026-04-19T12:00:00Z"}}`))
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if service.connectInput.UserID != testVKHandlerUserID {
		t.Fatalf("Connect() user id = %q, want %q", service.connectInput.UserID, testVKHandlerUserID)
	}
	if service.connectInput.ProviderUserID != "vk_1" {
		t.Fatalf("Connect() provider user id = %q, want %q", service.connectInput.ProviderUserID, "vk_1")
	}
}

func TestHandlerSyncInterestsRequiresAuthorization(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubVKIntegrationService{}, authhttp.NewMiddleware(stubVKAuthorizer{}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/vk/connection/sync-interests", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestHandlerSyncInterestsPropagatesServiceError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubVKIntegrationService{
		syncErr: apperrors.New(apperrors.KindConflict, "vk_consent_required", "vk consent is required before sync"),
	}, authhttp.NewMiddleware(stubVKAuthorizer{
		actor: authusecase.Actor{UserID: testVKHandlerUserID, Role: "user"},
	}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/vk/connection/sync-interests", nil)
	req.Header.Set("Authorization", "Bearer token")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", recorder.Code)
	}
}

type stubVKIntegrationService struct {
	connectInput      vkintegrationusecase.ConnectInput
	connectOutput     vkintegrationusecase.ConnectOutput
	connectErr        error
	exchangeInput     vkintegrationusecase.ExchangeOAuthInput
	exchangeOutput    vkintegrationusecase.ExchangeOAuthOutput
	exchangeErr       error
	getInput         vkintegrationusecase.GetCurrentConnectionInput
	getOutput        vkintegrationusecase.GetCurrentConnectionOutput
	getErr           error
	disconnectInput  vkintegrationusecase.DisconnectInput
	disconnectOutput vkintegrationusecase.DisconnectOutput
	disconnectErr    error
	syncInput        vkintegrationusecase.SyncInterestsInput
	syncOutput       vkintegrationusecase.SyncInterestsOutput
	syncErr          error
}

func (s *stubVKIntegrationService) Connect(_ context.Context, input vkintegrationusecase.ConnectInput) (vkintegrationusecase.ConnectOutput, error) {
	s.connectInput = input
	return s.connectOutput, s.connectErr
}

func (s *stubVKIntegrationService) ExchangeOAuth(_ context.Context, input vkintegrationusecase.ExchangeOAuthInput) (vkintegrationusecase.ExchangeOAuthOutput, error) {
	s.exchangeInput = input
	return s.exchangeOutput, s.exchangeErr
}

func (s *stubVKIntegrationService) GetCurrentConnection(_ context.Context, input vkintegrationusecase.GetCurrentConnectionInput) (vkintegrationusecase.GetCurrentConnectionOutput, error) {
	s.getInput = input
	return s.getOutput, s.getErr
}

func (s *stubVKIntegrationService) Disconnect(_ context.Context, input vkintegrationusecase.DisconnectInput) (vkintegrationusecase.DisconnectOutput, error) {
	s.disconnectInput = input
	return s.disconnectOutput, s.disconnectErr
}

func (s *stubVKIntegrationService) SyncInterests(_ context.Context, input vkintegrationusecase.SyncInterestsInput) (vkintegrationusecase.SyncInterestsOutput, error) {
	s.syncInput = input
	return s.syncOutput, s.syncErr
}

type stubVKAuthorizer struct {
	actor authusecase.Actor
	err   error
}

func (a stubVKAuthorizer) Authorize(context.Context, string) (authusecase.Actor, error) {
	if a.err != nil {
		return authusecase.Actor{}, a.err
	}

	return a.actor, nil
}
