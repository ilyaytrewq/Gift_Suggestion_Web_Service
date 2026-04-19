package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	authhttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/delivery/http"
	recommendationusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/recommendation/usecase"
	httpapi "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/transport/httpapi"
)

var ErrNilRecommendationService = errors.New("recommendation service is nil")

type service interface {
	Recommend(ctx context.Context, input recommendationusecase.RecommendInput) (recommendationusecase.RecommendOutput, error)
	GetRecommendation(ctx context.Context, input recommendationusecase.GetRecommendationInput) (recommendationusecase.GetRecommendationOutput, error)
}

type Handler struct {
	service        service
	authMiddleware gin.HandlerFunc
}

func NewHandler(service service, authMiddleware gin.HandlerFunc) (*Handler, error) {
	if service == nil {
		return nil, ErrNilRecommendationService
	}

	return &Handler{
		service:        service,
		authMiddleware: authMiddleware,
	}, nil
}

func (h *Handler) Register(root gin.IRouter) {
	recommendations := root.Group("/recommendations")
	recommendations.Use(h.authMiddleware)
	recommendations.POST("", h.recommend)
	recommendations.GET("/:request_id", h.getRecommendation)
}

func (h *Handler) recommend(c *gin.Context) {
	actor, ok := authhttp.ActorFromContext(c)
	if !ok {
		httpapi.Fail(c, authhttp.UnauthorizedError())
		return
	}

	var request recommendRequest
	if err := httpapi.DecodeJSON(c, &request); err != nil {
		httpapi.Fail(c, err)
		return
	}

	output, err := h.service.Recommend(c.Request.Context(), recommendationusecase.RecommendInput{
		UserID:               actor.UserID,
		Occasion:             request.Occasion,
		Relationship:         request.Relationship,
		RecipientAge:         request.RecipientAge,
		BudgetMax:            request.BudgetMax,
		PreferredCategoryIDs: request.PreferredCategoryIDs,
		Interests:            request.Interests,
		TopN:                 request.TopN,
		UseWishlistContext:   request.UseWishlistContext,
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, http.StatusOK, output)
}

func (h *Handler) getRecommendation(c *gin.Context) {
	actor, ok := authhttp.ActorFromContext(c)
	if !ok {
		httpapi.Fail(c, authhttp.UnauthorizedError())
		return
	}

	output, err := h.service.GetRecommendation(c.Request.Context(), recommendationusecase.GetRecommendationInput{
		UserID:    actor.UserID,
		RequestID: c.Param("request_id"),
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, http.StatusOK, output)
}
