package http

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	authusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
	httpapi "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/transport/httpapi"
)

const actorContextKey = "auth_actor"

type authorizer interface {
	Authorize(ctx context.Context, rawAccessToken string) (authusecase.Actor, error)
}

func NewMiddleware(authorizer authorizer) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(header, "Bearer ") {
			httpapi.Fail(c, apperrors.New(
				apperrors.KindUnauthorized,
				"missing_access_token",
				"access token is required",
			))
			return
		}

		actor, err := authorizer.Authorize(c.Request.Context(), strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			httpapi.Fail(c, err)
			return
		}

		c.Set(actorContextKey, actor)
		c.Next()
	}
}

func ActorFromContext(c *gin.Context) (authusecase.Actor, bool) {
	value, ok := c.Get(actorContextKey)
	if !ok {
		return authusecase.Actor{}, false
	}

	actor, ok := value.(authusecase.Actor)
	return actor, ok
}

func UnauthorizedError() error {
	return apperrors.New(
		apperrors.KindUnauthorized,
		"missing_access_token",
		"access token is required",
	)
}

const adminRole = "admin"

// RequireAdmin requires a valid Bearer actor with role admin (JWT claim).
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		actor, ok := ActorFromContext(c)
		if !ok {
			httpapi.Fail(c, UnauthorizedError())
			return
		}

		if actor.Role != adminRole {
			httpapi.Fail(c, apperrors.New(
				apperrors.KindForbidden,
				"forbidden",
				"admin role is required",
			))
			return
		}

		c.Next()
	}
}
