package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	authhttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/delivery/http"
	userusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
	httpapi "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/transport/httpapi"
)

var ErrNilUserService = errors.New("user service is nil")

type service interface {
	GetCurrentUser(ctx context.Context, userID string) (userusecase.Profile, error)
	UpdateProfile(ctx context.Context, input userusecase.UpdateProfileInput) (userusecase.Profile, error)
	PromoteUserToAdmin(ctx context.Context, email string) (userusecase.Profile, error)
}

type Handler struct {
	service        service
	authMiddleware gin.HandlerFunc
}

func NewHandler(service service, authMiddleware gin.HandlerFunc) (*Handler, error) {
	if service == nil {
		return nil, ErrNilUserService
	}

	return &Handler{
		service:        service,
		authMiddleware: authMiddleware,
	}, nil
}

func (h *Handler) Register(root gin.IRouter) {
	users := root.Group("/users")
	users.Use(h.authMiddleware)
	users.GET("/me", h.getCurrentUser)
	users.PATCH("/me", h.updateProfile)

	admin := root.Group("/admin/users")
	admin.Use(h.authMiddleware, authhttp.RequireAdmin())
	admin.POST("/promote", h.promoteToAdmin)
}

func (h *Handler) getCurrentUser(c *gin.Context) {
	actor, ok := authhttp.ActorFromContext(c)
	if !ok {
		httpapi.Fail(c, authhttp.UnauthorizedError())
		return
	}

	profile, err := h.service.GetCurrentUser(c.Request.Context(), actor.UserID)
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, http.StatusOK, gin.H{"user": profile})
}

func (h *Handler) updateProfile(c *gin.Context) {
	actor, ok := authhttp.ActorFromContext(c)
	if !ok {
		httpapi.Fail(c, authhttp.UnauthorizedError())
		return
	}

	var request updateProfileRequest
	if err := httpapi.DecodeJSON(c, &request); err != nil {
		httpapi.Fail(c, err)
		return
	}
	if !request.Profile.DisplayName.Set {
		httpapi.Fail(c, apperrors.New(
			apperrors.KindValidation,
			"invalid_profile_update",
			"at least one profile field is required",
		))
		return
	}

	displayName := ""
	if request.Profile.DisplayName.Value != nil {
		displayName = *request.Profile.DisplayName.Value
	}

	profile, err := h.service.UpdateProfile(c.Request.Context(), userusecase.UpdateProfileInput{
		UserID:      actor.UserID,
		DisplayName: displayName,
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, http.StatusOK, gin.H{"user": profile})
}

func (h *Handler) promoteToAdmin(c *gin.Context) {
	var req promoteToAdminRequest
	if err := httpapi.DecodeJSON(c, &req); err != nil {
		httpapi.Fail(c, err)
		return
	}

	email := strings.TrimSpace(req.Email)
	if email == "" {
		httpapi.Fail(c, apperrors.New(
			apperrors.KindValidation,
			"missing_email",
			"email is required",
		))
		return
	}

	profile, err := h.service.PromoteUserToAdmin(c.Request.Context(), email)
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, http.StatusOK, gin.H{"user": profile})
}
