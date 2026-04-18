package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

func DecodeJSON(c *gin.Context, destination any) error {
	contentType := c.GetHeader("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		return apperrors.New(
			apperrors.KindValidation,
			"invalid_content_type",
			"content type must be application/json",
		)
	}

	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		if errors.Is(err, io.EOF) {
			return apperrors.New(
				apperrors.KindValidation,
				"invalid_request_body",
				"request body must not be empty",
			)
		}

		return apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_request_body",
			"request body is invalid",
			err,
		)
	}

	if decoder.More() {
		return apperrors.New(
			apperrors.KindValidation,
			"invalid_request_body",
			"request body must contain a single JSON object",
		)
	}

	return nil
}
