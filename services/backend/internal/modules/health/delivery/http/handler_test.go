package healthhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/health/domain"
	transporthttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/transport/httpapi"
)

type stubService struct {
	liveReport  domain.Report
	readyReport domain.Report
}

func (s stubService) Live(context.Context) domain.Report {
	return s.liveReport
}

func (s stubService) Ready(context.Context) domain.Report {
	return s.readyReport
}

func TestHandlerLiveRoute(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(stubService{
		liveReport: domain.NewLiveReport(time.Date(2026, time.April, 18, 9, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	router := gin.New()
	router.Use(transporthttp.RequestID())
	handler.Register(router)

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var envelope struct {
		Status string `json:"status"`
		Meta   struct {
			RequestID string `json:"request_id"`
		} `json:"meta"`
	}

	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unexpected json error: %v", err)
	}

	if envelope.Status != "ok" {
		t.Fatalf("expected ok status, got %s", envelope.Status)
	}
	if envelope.Meta.RequestID == "" {
		t.Fatal("expected request id to be present")
	}
}

func TestHandlerReadyRouteReturnsServiceUnavailable(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(stubService{
		readyReport: domain.Report{
			Status:    domain.StatusDown,
			CheckedAt: time.Date(2026, time.April, 18, 9, 0, 0, 0, time.UTC),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	router := gin.New()
	router.Use(transporthttp.RequestID())
	handler.Register(router)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", recorder.Code)
	}
}
