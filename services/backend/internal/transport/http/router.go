package http

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
	httpapi "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/transport/httpapi"
)

type HealthHandler interface {
	Register(root gin.IRoutes)
}

func NewRouter(logger *slog.Logger, healthHandler HealthHandler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(httpapi.RequestID())
	router.Use(httpapi.RequestLogger(logger))

	router.NoRoute(func(c *gin.Context) {
		httpapi.Fail(c, apperrors.New(
			apperrors.KindNotFound,
			"route_not_found",
			"route not found",
		))
	})

	router.NoMethod(func(c *gin.Context) {
		httpapi.Fail(c, apperrors.New(
			apperrors.KindMethod,
			"method_not_allowed",
			"method not allowed",
		))
	})

	if healthHandler != nil {
		healthHandler.Register(router)
	}

	api := router.Group("/api/v1")
	api.GET("/health/live", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/health/live")
	})
	api.GET("/health/ready", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/health/ready")
	})

	return router
}
