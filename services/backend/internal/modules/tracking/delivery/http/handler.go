package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	authhttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/delivery/http"
	trackingusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/tracking/usecase"
	httpapi "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/transport/httpapi"
)

var ErrNilTrackingService = errors.New("tracking service is nil")

type service interface {
	TrackEvent(ctx context.Context, input trackingusecase.TrackEventInput) (trackingusecase.TrackEventOutput, error)
}

type Handler struct {
	service        service
	authMiddleware gin.HandlerFunc
}

func NewHandler(service service, authMiddleware gin.HandlerFunc) (*Handler, error) {
	if service == nil {
		return nil, ErrNilTrackingService
	}

	return &Handler{
		service:        service,
		authMiddleware: authMiddleware,
	}, nil
}

func (h *Handler) Register(root gin.IRouter) {
	events := root.Group("/tracking/events")
	events.Use(h.authMiddleware)
	events.POST("", h.trackEvent)
}

func (h *Handler) trackEvent(c *gin.Context) {
	actor, ok := authhttp.ActorFromContext(c)
	if !ok {
		httpapi.Fail(c, authhttp.UnauthorizedError())
		return
	}

	var request trackEventRequest
	if err := httpapi.DecodeJSON(c, &request); err != nil {
		httpapi.Fail(c, err)
		return
	}

	output, err := h.service.TrackEvent(c.Request.Context(), trackingusecase.TrackEventInput{
		UserID:                  actor.UserID,
		Type:                    request.Type,
		ClientEventID:           request.ClientEventID,
		RecommendationRequestID: request.RecommendationRequestID,
		WishlistID:              request.WishlistID,
		GiftID:                  request.GiftID,
		OccurredAt:              request.OccurredAt,
		Metadata:                mapMetadataInput(request.Metadata),
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	statusCode := http.StatusCreated
	if output.Event.Duplicate {
		statusCode = http.StatusOK
	}

	httpapi.Success(c, statusCode, output)
}

func mapMetadataInput(metadata *trackEventMetadata) trackingusecase.EventMetadataInput {
	if metadata == nil {
		return trackingusecase.EventMetadataInput{}
	}

	return trackingusecase.EventMetadataInput{
		Surface:  metadata.Surface,
		Position: metadata.Position,
	}
}
