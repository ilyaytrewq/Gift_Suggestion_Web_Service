package httpapi

import "github.com/gin-gonic/gin"

type envelope struct {
	Status string        `json:"status"`
	Data   any           `json:"data,omitempty"`
	Error  *errorPayload `json:"error,omitempty"`
	Meta   responseMeta  `json:"meta"`
}

type errorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type responseMeta struct {
	RequestID string `json:"request_id,omitempty"`
}

func Success(c *gin.Context, statusCode int, data any) {
	c.JSON(statusCode, envelope{
		Status: "ok",
		Data:   data,
		Meta: responseMeta{
			RequestID: RequestIDFromContext(c),
		},
	})
}
