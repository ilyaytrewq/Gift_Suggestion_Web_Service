package http

import (
	"context"
	"errors"
	nethttp "net/http"

	"github.com/gin-gonic/gin"

	authhttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/delivery/http"
	vkintegrationusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/usecase"
	httpapi "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/transport/httpapi"
)

var ErrNilVKIntegrationService = errors.New("vk integration service is nil")

type service interface {
	Connect(ctx context.Context, input vkintegrationusecase.ConnectInput) (vkintegrationusecase.ConnectOutput, error)
	ExchangeOAuth(ctx context.Context, input vkintegrationusecase.ExchangeOAuthInput) (vkintegrationusecase.ExchangeOAuthOutput, error)
	GetCurrentConnection(ctx context.Context, input vkintegrationusecase.GetCurrentConnectionInput) (vkintegrationusecase.GetCurrentConnectionOutput, error)
	Disconnect(ctx context.Context, input vkintegrationusecase.DisconnectInput) (vkintegrationusecase.DisconnectOutput, error)
	SyncInterests(ctx context.Context, input vkintegrationusecase.SyncInterestsInput) (vkintegrationusecase.SyncInterestsOutput, error)
}

type Handler struct {
	service        service
	authMiddleware gin.HandlerFunc
}

func NewHandler(service service, authMiddleware gin.HandlerFunc) (*Handler, error) {
	if service == nil {
		return nil, ErrNilVKIntegrationService
	}

	return &Handler{
		service:        service,
		authMiddleware: authMiddleware,
	}, nil
}

func (h *Handler) Register(root gin.IRouter) {
	vk := root.Group("/integrations/vk")
	vk.Use(h.authMiddleware)
	vk.POST("/oauth/exchange", h.exchangeOAuth)

	connection := vk.Group("/connection")
	connection.GET("", h.getCurrentConnection)
	connection.PUT("", h.connect)
	connection.DELETE("", h.disconnect)
	connection.POST("/sync-interests", h.syncInterests)
}

func (h *Handler) exchangeOAuth(c *gin.Context) {
	actor, ok := authhttp.ActorFromContext(c)
	if !ok {
		httpapi.Fail(c, authhttp.UnauthorizedError())
		return
	}

	var request exchangeOAuthRequest
	if err := httpapi.DecodeJSON(c, &request); err != nil {
		httpapi.Fail(c, err)
		return
	}

	output, err := h.service.ExchangeOAuth(c.Request.Context(), vkintegrationusecase.ExchangeOAuthInput{
		UserID:       actor.UserID,
		Code:         request.Code,
		CodeVerifier: request.CodeVerifier,
		DeviceID:     request.DeviceID,
		State:        request.State,
		RedirectURI:  request.RedirectURI,
		Consent: vkintegrationusecase.ConsentInput{
			Granted:    request.Consent.Granted,
			Version:    request.Consent.Version,
			ObtainedAt: request.Consent.ObtainedAt,
		},
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, nethttp.StatusOK, output)
}

func (h *Handler) connect(c *gin.Context) {
	actor, ok := authhttp.ActorFromContext(c)
	if !ok {
		httpapi.Fail(c, authhttp.UnauthorizedError())
		return
	}

	var request connectRequest
	if err := httpapi.DecodeJSON(c, &request); err != nil {
		httpapi.Fail(c, err)
		return
	}

	output, err := h.service.Connect(c.Request.Context(), vkintegrationusecase.ConnectInput{
		UserID:         actor.UserID,
		ProviderUserID: request.ProviderUserID,
		Consent: vkintegrationusecase.ConsentInput{
			Granted:    request.Consent.Granted,
			Version:    request.Consent.Version,
			ObtainedAt: request.Consent.ObtainedAt,
		},
		Credential: mapCredentialInput(request.Credential),
		Profile:    mapProfileInput(request.Profile),
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, nethttp.StatusOK, output)
}

func (h *Handler) getCurrentConnection(c *gin.Context) {
	actor, ok := authhttp.ActorFromContext(c)
	if !ok {
		httpapi.Fail(c, authhttp.UnauthorizedError())
		return
	}

	output, err := h.service.GetCurrentConnection(c.Request.Context(), vkintegrationusecase.GetCurrentConnectionInput{
		UserID: actor.UserID,
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, nethttp.StatusOK, output)
}

func (h *Handler) disconnect(c *gin.Context) {
	actor, ok := authhttp.ActorFromContext(c)
	if !ok {
		httpapi.Fail(c, authhttp.UnauthorizedError())
		return
	}

	output, err := h.service.Disconnect(c.Request.Context(), vkintegrationusecase.DisconnectInput{
		UserID: actor.UserID,
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, nethttp.StatusOK, output)
}

func (h *Handler) syncInterests(c *gin.Context) {
	actor, ok := authhttp.ActorFromContext(c)
	if !ok {
		httpapi.Fail(c, authhttp.UnauthorizedError())
		return
	}

	output, err := h.service.SyncInterests(c.Request.Context(), vkintegrationusecase.SyncInterestsInput{
		UserID: actor.UserID,
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, nethttp.StatusOK, output)
}

func mapCredentialInput(request *credentialRequest) vkintegrationusecase.CredentialInput {
	if request == nil {
		return vkintegrationusecase.CredentialInput{}
	}

	return vkintegrationusecase.CredentialInput{
		AccessToken: request.AccessToken,
		ExpiresAt:   request.ExpiresAt,
		Scopes:      append([]string(nil), request.Scopes...),
	}
}

func mapProfileInput(request *profileRequest) vkintegrationusecase.ProfileInput {
	if request == nil {
		return vkintegrationusecase.ProfileInput{}
	}

	return vkintegrationusecase.ProfileInput{
		ScreenName: request.ScreenName,
		ProfileURL: request.ProfileURL,
	}
}
