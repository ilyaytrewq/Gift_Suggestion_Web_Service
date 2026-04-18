package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	catalogusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/usecase"
	httpapi "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/transport/httpapi"
)

var ErrNilCatalogService = errors.New("catalog service is nil")

type service interface {
	ListGifts(ctx context.Context, input catalogusecase.ListGiftsInput) (catalogusecase.ListGiftsOutput, error)
	GetGift(ctx context.Context, input catalogusecase.GetGiftInput) (catalogusecase.GetGiftOutput, error)
	ListCategories(ctx context.Context, input catalogusecase.ListCategoriesInput) (catalogusecase.ListCategoriesOutput, error)
}

type Handler struct {
	service service
}

func NewHandler(service service) (*Handler, error) {
	if service == nil {
		return nil, ErrNilCatalogService
	}

	return &Handler{service: service}, nil
}

func (h *Handler) Register(root gin.IRouter) {
	catalog := root.Group("/catalog")
	catalog.GET("/gifts", h.listGifts)
	catalog.GET("/gifts/:gift_id", h.getGift)
	catalog.GET("/categories", h.listCategories)
}

func (h *Handler) listGifts(c *gin.Context) {
	input, err := parseListGiftsInput(c.Request.URL.Query())
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	output, err := h.service.ListGifts(c.Request.Context(), input)
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, http.StatusOK, output)
}

func (h *Handler) getGift(c *gin.Context) {
	output, err := h.service.GetGift(c.Request.Context(), catalogusecase.GetGiftInput{
		GiftID: c.Param("gift_id"),
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, http.StatusOK, output)
}

func (h *Handler) listCategories(c *gin.Context) {
	input, err := parseListCategoriesInput(c.Request.URL.Query())
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	output, err := h.service.ListCategories(c.Request.Context(), input)
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, http.StatusOK, output)
}
