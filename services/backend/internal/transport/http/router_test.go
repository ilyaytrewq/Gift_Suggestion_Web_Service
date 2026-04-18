package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubHealthHandler struct{}

func (stubHealthHandler) Register(root gin.IRouter) {
	root.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func TestNewRouterReturnsFormattedNotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	router := NewRouter(nil, stubHealthHandler{})
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}

	var body struct {
		Status string `json:"status"`
		Error  struct {
			Code string `json:"code"`
		} `json:"error"`
	}

	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected json error: %v", err)
	}

	if body.Status != "error" {
		t.Fatalf("expected error envelope, got %s", body.Status)
	}
	if body.Error.Code != "route_not_found" {
		t.Fatalf("expected route_not_found code, got %s", body.Error.Code)
	}
}
