package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	catalogusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/usecase"
)

const (
	testCatalogHandlerGiftID = "550e8400-e29b-41d4-a716-446655440010"
)

func TestHandlerListGiftsRejectsUnknownQueryParam(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(stubCatalogService{})
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/gifts?unexpected=1", nil)
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

func TestHandlerGetGiftSuccess(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	service := stubCatalogService{
		getGiftOutput: catalogusecase.GetGiftOutput{
			Gift: catalogusecase.Gift{
				ID:          testCatalogHandlerGiftID,
				Name:        "LEGO Set",
				Description: "Creative building set",
				Price:       "129.99",
				StoreLink:   "https://example.com/gifts/lego",
				CreatedAt:   time.Date(2026, 4, 18, 12, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2026, 4, 18, 12, 30, 0, 0, time.UTC),
			},
		},
	}
	handler, err := NewHandler(service)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/gifts/"+testCatalogHandlerGiftID, nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var response getGiftResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if response.Data.Gift.ID != testCatalogHandlerGiftID {
		t.Fatalf("gift id = %q, want %q", response.Data.Gift.ID, testCatalogHandlerGiftID)
	}
}

func TestHandlerListCategoriesSuccess(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	service := stubCatalogService{
		listCategoriesOutput: catalogusecase.ListCategoriesOutput{
			Items: []catalogusecase.Category{
				{
					ID:        "550e8400-e29b-41d4-a716-446655440011",
					Name:      "Books",
					CreatedAt: time.Date(2026, 4, 18, 11, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2026, 4, 18, 11, 0, 0, 0, time.UTC),
				},
			},
			Page: catalogusecase.Page{Limit: 20, Offset: 0, Total: 1},
		},
	}
	handler, err := NewHandler(service)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalog/categories", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var response listCategoriesResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(response.Data.Items) != 1 {
		t.Fatalf("expected one category, got %d", len(response.Data.Items))
	}
}

type stubCatalogService struct {
	listGiftsOutput      catalogusecase.ListGiftsOutput
	getGiftOutput        catalogusecase.GetGiftOutput
	listCategoriesOutput catalogusecase.ListCategoriesOutput
}

func (s stubCatalogService) ListGifts(context.Context, catalogusecase.ListGiftsInput) (catalogusecase.ListGiftsOutput, error) {
	return s.listGiftsOutput, nil
}

func (s stubCatalogService) GetGift(context.Context, catalogusecase.GetGiftInput) (catalogusecase.GetGiftOutput, error) {
	return s.getGiftOutput, nil
}

func (s stubCatalogService) ListCategories(context.Context, catalogusecase.ListCategoriesInput) (catalogusecase.ListCategoriesOutput, error) {
	return s.listCategoriesOutput, nil
}

type handlerErrorResponse struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

type getGiftResponse struct {
	Data struct {
		Gift catalogusecase.Gift `json:"gift"`
	} `json:"data"`
}

type listCategoriesResponse struct {
	Data struct {
		Items []catalogusecase.Category `json:"items"`
	} `json:"data"`
}
