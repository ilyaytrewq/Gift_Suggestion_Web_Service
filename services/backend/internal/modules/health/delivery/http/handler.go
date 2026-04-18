package healthhttp

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/health/domain"
	transporthttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/transport/httpapi"
)

var ErrNilHealthService = errors.New("health service is nil")

type healthService interface {
	Live(ctx context.Context) domain.Report
	Ready(ctx context.Context) domain.Report
}

type Handler struct {
	service healthService
}

func NewHandler(service healthService) (*Handler, error) {
	if service == nil {
		return nil, ErrNilHealthService
	}

	return &Handler{service: service}, nil
}

func (h *Handler) Register(root gin.IRouter) {
	root.GET("/health/live", h.live)
	root.GET("/health/ready", h.ready)
}

func (h *Handler) live(c *gin.Context) {
	transporthttp.Success(c, http.StatusOK, h.service.Live(c.Request.Context()))
}

func (h *Handler) ready(c *gin.Context) {
	report := h.service.Ready(c.Request.Context())
	statusCode := http.StatusOK
	if report.Status == domain.StatusDown {
		statusCode = http.StatusServiceUnavailable
	}

	transporthttp.Success(c, statusCode, report)
}
