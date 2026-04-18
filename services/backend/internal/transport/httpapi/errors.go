package httpapi

import (
	"github.com/gin-gonic/gin"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

func Fail(c *gin.Context, err error) {
	appErr := apperrors.From(err)
	c.AbortWithStatusJSON(appErr.HTTPStatus(), envelope{
		Status: "error",
		Error: &errorPayload{
			Code:    appErr.Code(),
			Message: appErr.Message(),
		},
		Meta: responseMeta{
			RequestID: RequestIDFromContext(c),
		},
	})
}
