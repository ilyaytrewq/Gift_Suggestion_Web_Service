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

func TestHandlerCurrentWishlistRequiresAuthorization(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubWishlistService{}, authhttp.NewMiddleware(stubAuthorizer{}))
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/wishlist", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestHandlerGetCurrentWishlistSuccess(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	service := &stubWishlistService{
		getOutput: wishlistusecase.GetWishlistOutput{
			Wishlist: wishlistusecase.Wishlist{
				ID:        testHandlerWishlistID,
				Name:      "Список желаний",
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

	req := httptest.NewRequest(http.MethodGet, "/api/v1/wishlist", nil)
	req.Header.Set("Authorization", "Bearer access-token")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if service.getInput.UserID != testHandlerUserID {
		t.Fatalf("GetWishlist() user id = %q, want %q", service.getInput.UserID, testHandlerUserID)
	}
	if service.getInput.WishlistID != "" {
		t.Fatalf("GetWishlist() wishlist id = %q, want empty", service.getInput.WishlistID)
	}
}

func TestHandlerAddCurrentWishlistItemReturnsOKWhenGiftAlreadySaved(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubWishlistService{
		addOutput: wishlistusecase.AddWishlistItemOutput{
			AlreadyInWishlist: true,
			Item: wishlistusecase.WishlistItem{
				ID:        "550e8400-e29b-41d4-a716-446655440105",
				CreatedAt: time.Date(2026, 4, 19, 11, 0, 0, 0, time.UTC),
				Gift: wishlistusecase.GiftPreview{
					ID:          testHandlerGiftID,
					Name:        "LEGO Set",
					Description: "Creative building set",
					Price:       "129.99",
					StoreLink:   "https://example.com/gifts/lego",
					CreatedAt:   time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC),
					UpdatedAt:   time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC),
				},
			},
		},
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
		"/api/v1/wishlist/items",
		bytes.NewBufferString(`{"gift_id":"`+testHandlerGiftID+`"}`),
	)
	req.Header.Set("Authorization", "Bearer access-token")
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestHandlerCompatibilityListRejectsUnknownQueryParam(t *testing.T) {
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

func TestHandlerDeleteCurrentWishlistSuccess(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/wishlist", nil)
	req.Header.Set("Authorization", "Bearer access-token")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestHandlerCompatibilityAddWishlistItemReturnsValidationError(t *testing.T) {
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

type stubWishlistService struct {
	createOutput wishlistusecase.CreateWishlistOutput
	createErr    error
	createInput  wishlistusecase.CreateWishlistInput

	listOutput wishlistusecase.ListWishlistsOutput
	listErr    error
	listInput  wishlistusecase.ListWishlistsInput

	getOutput wishlistusecase.GetWishlistOutput
	getErr    error
	getInput  wishlistusecase.GetWishlistInput

	addOutput wishlistusecase.AddWishlistItemOutput
	addErr    error
	addInput  wishlistusecase.AddWishlistItemInput

	removeOutput wishlistusecase.RemoveWishlistItemOutput
	removeErr    error
	removeInput  wishlistusecase.RemoveWishlistItemInput

	deleteOutput wishlistusecase.DeleteWishlistOutput
	deleteErr    error
	deleteInput  wishlistusecase.DeleteWishlistInput
}

func (s *stubWishlistService) CreateWishlist(_ context.Context, input wishlistusecase.CreateWishlistInput) (wishlistusecase.CreateWishlistOutput, error) {
	s.createInput = input
	return s.createOutput, s.createErr
}

func (s *stubWishlistService) ListWishlists(_ context.Context, input wishlistusecase.ListWishlistsInput) (wishlistusecase.ListWishlistsOutput, error) {
	s.listInput = input
	return s.listOutput, s.listErr
}

func (s *stubWishlistService) GetWishlist(_ context.Context, input wishlistusecase.GetWishlistInput) (wishlistusecase.GetWishlistOutput, error) {
	s.getInput = input
	return s.getOutput, s.getErr
}

func (s *stubWishlistService) AddWishlistItem(_ context.Context, input wishlistusecase.AddWishlistItemInput) (wishlistusecase.AddWishlistItemOutput, error) {
	s.addInput = input
	return s.addOutput, s.addErr
}

func (s *stubWishlistService) RemoveWishlistItem(_ context.Context, input wishlistusecase.RemoveWishlistItemInput) (wishlistusecase.RemoveWishlistItemOutput, error) {
	s.removeInput = input
	return s.removeOutput, s.removeErr
}

func (s *stubWishlistService) DeleteWishlist(_ context.Context, input wishlistusecase.DeleteWishlistInput) (wishlistusecase.DeleteWishlistOutput, error) {
	s.deleteInput = input
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
