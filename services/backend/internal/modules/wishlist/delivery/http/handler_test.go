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
	wishlistusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/wishlist/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

const (
	testHandlerUserID     = "550e8400-e29b-41d4-a716-446655440100"
	testHandlerSessionID  = "550e8400-e29b-41d4-a716-446655440200"
	testHandlerWishlistID = "550e8400-e29b-41d4-a716-446655440102"
	testHandlerGiftID     = "550e8400-e29b-41d4-a716-446655440104"
)

func TestHandlerCreateWishlistRequiresAuthorization(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubWishlistService{}, authhttp.NewMiddleware(stubAuthorizer{}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/wishlists", bytes.NewBufferString(`{"name":"Birthday Ideas"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestHandlerListWishlistsRejectsUnknownQueryParam(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubWishlistService{}, authhttp.NewMiddleware(stubAuthorizer{
		actor: authusecase.Actor{UserID: testHandlerUserID, SessionID: testHandlerSessionID, Role: "user"},
	}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/wishlists?bad=1", nil)
	req.Header.Set("Authorization", "Bearer access-token")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	var response handlerErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if response.Error.Code != "invalid_query_parameter" {
		t.Fatalf("error code = %q, want %q", response.Error.Code, "invalid_query_parameter")
	}
}

func TestHandlerCreateWishlistSuccess(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	service := &stubWishlistService{
		createOutput: wishlistusecase.CreateWishlistOutput{
			Wishlist: wishlistusecase.Wishlist{
				ID:        testHandlerWishlistID,
				Name:      "Birthday Ideas",
				ItemCount: 0,
				CreatedAt: time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC),
			},
		},
	}
	handler, err := NewHandler(service, authhttp.NewMiddleware(stubAuthorizer{
		actor: authusecase.Actor{UserID: testHandlerUserID, SessionID: testHandlerSessionID, Role: "user"},
	}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/wishlists", bytes.NewBufferString(`{"name":"Birthday Ideas"}`))
	req.Header.Set("Authorization", "Bearer access-token")
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}
	if service.createInput.UserID != testHandlerUserID {
		t.Fatalf("CreateWishlist() user id = %q, want %q", service.createInput.UserID, testHandlerUserID)
	}
	if service.createInput.Name != "Birthday Ideas" {
		t.Fatalf("CreateWishlist() name = %q, want %q", service.createInput.Name, "Birthday Ideas")
	}
}

func TestHandlerAddWishlistItemReturnsValidationError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubWishlistService{
		addErr: apperrors.New(apperrors.KindValidation, "invalid_gift_id", "gift id is invalid"),
	}, authhttp.NewMiddleware(stubAuthorizer{
		actor: authusecase.Actor{UserID: testHandlerUserID, SessionID: testHandlerSessionID, Role: "user"},
	}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/wishlists/"+testHandlerWishlistID+"/items",
		bytes.NewBufferString(`{"gift_id":""}`),
	)
	req.Header.Set("Authorization", "Bearer access-token")
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestHandlerDeleteWishlistSuccess(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubWishlistService{
		deleteOutput: wishlistusecase.DeleteWishlistOutput{Deleted: true},
	}, authhttp.NewMiddleware(stubAuthorizer{
		actor: authusecase.Actor{UserID: testHandlerUserID, SessionID: testHandlerSessionID, Role: "user"},
	}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/wishlists/"+testHandlerWishlistID, nil)
	req.Header.Set("Authorization", "Bearer access-token")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

type stubWishlistService struct {
	createOutput wishlistusecase.CreateWishlistOutput
	createErr    error
	createInput  wishlistusecase.CreateWishlistInput

	listOutput wishlistusecase.ListWishlistsOutput
	listErr    error

	getOutput wishlistusecase.GetWishlistOutput
	getErr    error

	addOutput wishlistusecase.AddWishlistItemOutput
	addErr    error

	removeOutput wishlistusecase.RemoveWishlistItemOutput
	removeErr    error

	deleteOutput wishlistusecase.DeleteWishlistOutput
	deleteErr    error
}

func (s *stubWishlistService) CreateWishlist(_ context.Context, input wishlistusecase.CreateWishlistInput) (wishlistusecase.CreateWishlistOutput, error) {
	s.createInput = input
	return s.createOutput, s.createErr
}

func (s *stubWishlistService) ListWishlists(context.Context, wishlistusecase.ListWishlistsInput) (wishlistusecase.ListWishlistsOutput, error) {
	return s.listOutput, s.listErr
}

func (s *stubWishlistService) GetWishlist(context.Context, wishlistusecase.GetWishlistInput) (wishlistusecase.GetWishlistOutput, error) {
	return s.getOutput, s.getErr
}

func (s *stubWishlistService) AddWishlistItem(context.Context, wishlistusecase.AddWishlistItemInput) (wishlistusecase.AddWishlistItemOutput, error) {
	return s.addOutput, s.addErr
}

func (s *stubWishlistService) RemoveWishlistItem(context.Context, wishlistusecase.RemoveWishlistItemInput) (wishlistusecase.RemoveWishlistItemOutput, error) {
	return s.removeOutput, s.removeErr
}

func (s *stubWishlistService) DeleteWishlist(context.Context, wishlistusecase.DeleteWishlistInput) (wishlistusecase.DeleteWishlistOutput, error) {
	return s.deleteOutput, s.deleteErr
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

type handlerErrorResponse struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}
