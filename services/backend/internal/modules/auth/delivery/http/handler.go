package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	authusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
	httpapi "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/transport/httpapi"
)

var ErrNilAuthService = errors.New("auth service is nil")

type service interface {
	Register(ctx context.Context, input authusecase.RegisterInput) (authusecase.RegisterOutput, error)
	Login(ctx context.Context, input authusecase.LoginInput) (authusecase.LoginOutput, error)
	Refresh(ctx context.Context, input authusecase.RefreshInput) (authusecase.RefreshOutput, error)
	Logout(ctx context.Context, input authusecase.LogoutInput) (authusecase.AcceptedOutput, error)
	RequestPasswordReset(ctx context.Context, input authusecase.RequestPasswordResetInput) (authusecase.AcceptedOutput, error)
	ConfirmPasswordReset(ctx context.Context, input authusecase.ConfirmPasswordResetInput) (authusecase.AcceptedOutput, error)
	ConfirmEmailVerification(ctx context.Context, input authusecase.ConfirmEmailVerificationInput) (authusecase.AcceptedOutput, error)
	Authorize(ctx context.Context, rawAccessToken string) (authusecase.Actor, error)
}

type RefreshCookieConfig struct {
	Name     string
	Path     string
	Domain   string
	Secure   bool
	MaxAge   int
	SameSite http.SameSite
}

type Handler struct {
	service       service
	refreshCookie RefreshCookieConfig
}

func NewHandler(service service, refreshCookie RefreshCookieConfig) (*Handler, error) {
	if service == nil {
		return nil, ErrNilAuthService
	}

	return &Handler{
		service:       service,
		refreshCookie: refreshCookie,
	}, nil
}

func (h *Handler) Register(root gin.IRouter) {
	root.POST("/users", h.register)

	authGroup := root.Group("/auth")
	authGroup.POST("/login", h.login)
	authGroup.POST("/refresh", h.refresh)
	authGroup.POST("/logout", h.logout)
	authGroup.POST("/password-reset/request", h.requestPasswordReset)
	authGroup.POST("/password-reset/confirm", h.confirmPasswordReset)
	authGroup.POST("/email-verification/confirm", h.confirmEmailVerification)
}

func (h *Handler) register(c *gin.Context) {
	var request registerRequest
	if err := httpapi.DecodeJSON(c, &request); err != nil {
		httpapi.Fail(c, err)
		return
	}

	response, err := h.service.Register(c.Request.Context(), authusecase.RegisterInput{
		Email:       request.Email,
		Password:    request.Password,
		DisplayName: request.DisplayName,
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, http.StatusCreated, response)
}

func (h *Handler) login(c *gin.Context) {
	var request loginRequest
	if err := httpapi.DecodeJSON(c, &request); err != nil {
		httpapi.Fail(c, err)
		return
	}

	response, err := h.service.Login(c.Request.Context(), authusecase.LoginInput{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	h.writeRefreshCookie(c, response.Auth.RefreshToken)
	response.Auth.RefreshToken = ""

	httpapi.Success(c, http.StatusOK, response)
}

func (h *Handler) refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(h.refreshCookie.Name)
	if err != nil || strings.TrimSpace(refreshToken) == "" {
		httpapi.Fail(c, apperrors.New(
			apperrors.KindUnauthorized,
			"invalid_refresh_token",
			"refresh token is invalid",
		))
		return
	}

	response, err := h.service.Refresh(c.Request.Context(), authusecase.RefreshInput{
		RefreshToken: refreshToken,
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	h.writeRefreshCookie(c, response.Auth.RefreshToken)
	response.Auth.RefreshToken = ""

	httpapi.Success(c, http.StatusOK, response)
}

func (h *Handler) requestPasswordReset(c *gin.Context) {
	var request passwordResetRequest
	if err := httpapi.DecodeJSON(c, &request); err != nil {
		httpapi.Fail(c, err)
		return
	}

	response, err := h.service.RequestPasswordReset(c.Request.Context(), authusecase.RequestPasswordResetInput{
		Email: request.Email,
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, http.StatusAccepted, response)
}

func (h *Handler) confirmPasswordReset(c *gin.Context) {
	var request passwordResetConfirmRequest
	if err := httpapi.DecodeJSON(c, &request); err != nil {
		httpapi.Fail(c, err)
		return
	}

	response, err := h.service.ConfirmPasswordReset(c.Request.Context(), authusecase.ConfirmPasswordResetInput{
		Token:       request.Token,
		NewPassword: request.NewPassword,
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, http.StatusOK, response)
}

func (h *Handler) confirmEmailVerification(c *gin.Context) {
	var request emailVerificationConfirmRequest
	if err := httpapi.DecodeJSON(c, &request); err != nil {
		httpapi.Fail(c, err)
		return
	}

	response, err := h.service.ConfirmEmailVerification(c.Request.Context(), authusecase.ConfirmEmailVerificationInput{
		Token: request.Token,
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, http.StatusOK, response)
}

func (h *Handler) logout(c *gin.Context) {
	refreshToken := ""
	if cookie, err := c.Cookie(h.refreshCookie.Name); err == nil {
		refreshToken = cookie
	}

	response, err := h.service.Logout(c.Request.Context(), authusecase.LogoutInput{
		RefreshToken: refreshToken,
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	h.clearRefreshCookie(c)

	httpapi.Success(c, http.StatusOK, response)
}

func (h *Handler) writeRefreshCookie(c *gin.Context, refreshToken string) {
	c.SetSameSite(h.refreshCookie.SameSite)
	c.SetCookie(
		h.refreshCookie.Name,
		refreshToken,
		h.refreshCookie.MaxAge,
		h.refreshCookie.Path,
		h.refreshCookie.Domain,
		h.refreshCookie.Secure,
		true,
	)
}

func (h *Handler) clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(h.refreshCookie.SameSite)
	c.SetCookie(
		h.refreshCookie.Name,
		"",
		-1,
		h.refreshCookie.Path,
		h.refreshCookie.Domain,
		h.refreshCookie.Secure,
		true,
	)
}
