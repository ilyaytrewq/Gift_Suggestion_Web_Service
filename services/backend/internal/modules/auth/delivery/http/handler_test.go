package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	authusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/usecase"
	userusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/usecase"
)

const (
	testAuthHandlerUserID       = "550e8400-e29b-41d4-a716-446655440000"
	testAuthHandlerEmail        = "user@example.com"
	testAuthHandlerAccessToken  = "access-token"
	testAuthHandlerRefreshToken = "refresh-token"
)

func TestHandlerRegisterCreatesUser(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	service := &stubAuthService{
		registerOutput: authusecase.RegisterOutput{
			User: userusecase.Profile{ID: testAuthHandlerUserID, Email: testAuthHandlerEmail, DisplayName: "Alice"},
		},
	}
	handler, err := NewHandler(service, RefreshCookieConfig{Name: "refresh_token", Path: "/api/v1/auth"})
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	body := `{"email":"user@example.com","password":"ValidPass1!","display_name":"Alice"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}
	if service.registerInput.DisplayName != "Alice" {
		t.Fatalf("Register() display name = %q, want %q", service.registerInput.DisplayName, "Alice")
	}

	var response authHandlerResponse[userEnvelope]
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if response.Status != "ok" {
		t.Fatalf("expected ok status, got %q", response.Status)
	}
	if response.Data.User.Email != testAuthHandlerEmail {
		t.Fatalf("response user email = %q, want %q", response.Data.User.Email, testAuthHandlerEmail)
	}
}

func TestHandlerLoginSetsRefreshCookieAndStripsBodyToken(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	service := &stubAuthService{
		loginOutput: authusecase.LoginOutput{
			User: userusecase.Profile{ID: testAuthHandlerUserID, Email: testAuthHandlerEmail},
			Auth: authusecase.AuthPayload{
				AccessToken:  testAuthHandlerAccessToken,
				RefreshToken: testAuthHandlerRefreshToken,
				TokenType:    "Bearer",
				ExpiresIn:    int64((15 * time.Minute).Seconds()),
			},
		},
	}
	handler, err := NewHandler(service, RefreshCookieConfig{
		Name:     "refresh_token",
		Path:     "/api/v1/auth",
		MaxAge:   3600,
		SameSite: http.SameSiteLaxMode,
	})
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	body := `{"email":"user@example.com","password":"ValidPass1!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if cookie := recorder.Header().Get("Set-Cookie"); cookie == "" {
		t.Fatal("expected refresh cookie to be set")
	}

	var response authHandlerResponse[loginEnvelope]
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if response.Data.Auth.RefreshToken != "" {
		t.Fatalf("expected refresh token to be removed from response body, got %q", response.Data.Auth.RefreshToken)
	}
	if response.Data.Auth.AccessToken != testAuthHandlerAccessToken {
		t.Fatalf("response access token = %q, want %q", response.Data.Auth.AccessToken, testAuthHandlerAccessToken)
	}
}

func TestHandlerRefreshRequiresCookie(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubAuthService{}, RefreshCookieConfig{Name: "refresh_token", Path: "/api/v1/auth"})
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestHandlerLogoutClearsRefreshCookie(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	service := &stubAuthService{
		logoutOutput: authusecase.AcceptedOutput{Accepted: true},
	}
	handler, err := NewHandler(service, RefreshCookieConfig{
		Name:     "refresh_token",
		Path:     "/api/v1/auth",
		MaxAge:   3600,
		SameSite: http.SameSiteLaxMode,
	})
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewBufferString("{}"))
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: testAuthHandlerRefreshToken})
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if service.logoutInput.RefreshToken != testAuthHandlerRefreshToken {
		t.Fatalf("Logout() refresh token = %q, want %q", service.logoutInput.RefreshToken, testAuthHandlerRefreshToken)
	}
	cookie := recorder.Header().Get("Set-Cookie")
	if cookie == "" {
		t.Fatal("expected refresh cookie to be cleared")
	}
	if !strings.Contains(cookie, "refresh_token=") {
		t.Fatalf("expected refresh cookie in Set-Cookie header, got %q", cookie)
	}
	if !strings.Contains(cookie, "Max-Age=0") {
		t.Fatalf("expected cleared cookie Max-Age=0, got %q", cookie)
	}

	var response authHandlerResponse[authusecase.AcceptedOutput]
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !response.Data.Accepted {
		t.Fatal("expected accepted=true")
	}
}

type stubAuthService struct {
	registerInput  authusecase.RegisterInput
	registerOutput authusecase.RegisterOutput
	loginOutput    authusecase.LoginOutput
	refreshOutput  authusecase.RefreshOutput
	logoutInput    authusecase.LogoutInput
	logoutOutput   authusecase.AcceptedOutput
	resetOutput    authusecase.AcceptedOutput
}

func (s *stubAuthService) Register(_ context.Context, input authusecase.RegisterInput) (authusecase.RegisterOutput, error) {
	s.registerInput = input
	return s.registerOutput, nil
}

func (s *stubAuthService) Login(context.Context, authusecase.LoginInput) (authusecase.LoginOutput, error) {
	return s.loginOutput, nil
}

func (s *stubAuthService) Refresh(context.Context, authusecase.RefreshInput) (authusecase.RefreshOutput, error) {
	return s.refreshOutput, nil
}

func (s *stubAuthService) Logout(_ context.Context, input authusecase.LogoutInput) (authusecase.AcceptedOutput, error) {
	s.logoutInput = input
	return s.logoutOutput, nil
}

func (s *stubAuthService) RequestPasswordReset(context.Context, authusecase.RequestPasswordResetInput) (authusecase.AcceptedOutput, error) {
	return s.resetOutput, nil
}

func (s *stubAuthService) Authorize(context.Context, string) (authusecase.Actor, error) {
	return authusecase.Actor{}, nil
}

type authHandlerResponse[T any] struct {
	Status string `json:"status"`
	Data   T      `json:"data"`
}

type userEnvelope struct {
	User userusecase.Profile `json:"user"`
}

type loginEnvelope struct {
	User userusecase.Profile     `json:"user"`
	Auth authusecase.AuthPayload `json:"auth"`
}
