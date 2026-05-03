package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	authhttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/delivery/http"
	authusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/usecase"
	recommendationusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/recommendation/usecase"
	httpapi "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/transport/httpapi"
)

var ErrNilRecommendationService = errors.New("recommendation service is nil")

type service interface {
	Recommend(ctx context.Context, input recommendationusecase.RecommendInput) (recommendationusecase.RecommendOutput, error)
	GetRecommendation(ctx context.Context, input recommendationusecase.GetRecommendationInput) (recommendationusecase.GetRecommendationOutput, error)
}

type authorizer interface {
	Authorize(ctx context.Context, rawAccessToken string) (authusecase.Actor, error)
}

type Handler struct {
	service        service
	authMiddleware gin.HandlerFunc
	authorizer     authorizer
}

func NewHandler(service service, authMiddleware gin.HandlerFunc, optionalAuthorizer ...authorizer) (*Handler, error) {
	if service == nil {
		return nil, ErrNilRecommendationService
	}

	var currentAuthorizer authorizer
	if len(optionalAuthorizer) > 0 {
		currentAuthorizer = optionalAuthorizer[0]
	}

	return &Handler{
		service:        service,
		authMiddleware: authMiddleware,
		authorizer:     currentAuthorizer,
	}, nil
}

func (h *Handler) Register(root gin.IRouter) {
	root.POST("/recommendations", h.recommend)

	recommendations := root.Group("/recommendations")
	recommendations.Use(h.authMiddleware)
	recommendations.GET("/:request_id", h.getRecommendation)
}

func (h *Handler) recommend(c *gin.Context) {
	var request recommendRequest
	if err := httpapi.DecodeJSON(c, &request); err != nil {
		httpapi.Fail(c, err)
		return
	}

	userID := ""
	if actor := h.authorizeIfPresent(c); actor != nil {
		userID = actor.UserID
	}

	output, err := h.service.Recommend(c.Request.Context(), recommendationusecase.RecommendInput{
		UserID:               userID,
		Occasion:             request.Occasion,
		Relationship:         request.Relationship,
		RecipientAge:         request.RecipientAge,
		RecipientGender:      request.RecipientGender,
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

func (h *Handler) authorizeIfPresent(c *gin.Context) *authusecase.Actor {
	if h.authorizer == nil {
		return nil
	}

	header := strings.TrimSpace(c.GetHeader("Authorization"))
	if header == "" || !strings.HasPrefix(header, "Bearer ") {
		return nil
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	if token == "" {
		return nil
	}

	actor, err := h.authorizer.Authorize(c.Request.Context(), token)
	if err != nil {
		return nil
	}

	return &actor
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
