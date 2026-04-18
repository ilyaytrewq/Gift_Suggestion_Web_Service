package http

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	authhttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/delivery/http"
	authusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/usecase"
	catalogimportusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalogimport/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

const (
	testAdminUserID = "550e8400-e29b-41d4-a716-446655440910"
	testAdminJobID  = "550e8400-e29b-41d4-a716-446655440911"
)

func TestHandlerRunImportRequiresAuthorization(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubImportService{}, authhttp.NewMiddleware(stubImportAuthorizer{}), 1024)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	body, contentType := multipartBody(t, "catalog.csv", "data")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/import-jobs", body)
	req.Header.Set("Content-Type", contentType)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestHandlerRunImportRequiresAdminRole(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubImportService{}, authhttp.NewMiddleware(stubImportAuthorizer{
		actor: authusecase.Actor{UserID: testAdminUserID, Role: "user"},
	}), 1024)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	body, contentType := multipartBody(t, "catalog.csv", "data")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/import-jobs", body)
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("Content-Type", contentType)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
}

func TestHandlerRunImportSuccess(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	service := &stubImportService{
		runOutput: catalogimportusecase.RunImportOutput{
			Job: catalogimportusecase.Job{
				ID:              testAdminJobID,
				Status:          "completed",
				SourceFormat:    "csv",
				SourceFilename:  "catalog.csv",
				SourceSizeBytes: 4,
				Summary: catalogimportusecase.Summary{
					ImportedRows: 1,
				},
				CreatedAt: time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC),
			},
		},
	}
	handler, err := NewHandler(service, authhttp.NewMiddleware(stubImportAuthorizer{
		actor: authusecase.Actor{UserID: testAdminUserID, Role: "admin"},
	}), 1024)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	body, contentType := multipartBody(t, "catalog.csv", "data")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/import-jobs", body)
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("Content-Type", contentType)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}
	if service.runInput.RequestedByUserID != testAdminUserID {
		t.Fatalf("RunImport() user id = %q, want %q", service.runInput.RequestedByUserID, testAdminUserID)
	}
}

func TestHandlerRunImportRejectsInvalidMultipart(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubImportService{}, authhttp.NewMiddleware(stubImportAuthorizer{
		actor: authusecase.Actor{UserID: testAdminUserID, Role: "admin"},
	}), 1024)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/import-jobs", bytes.NewBufferString("broken"))
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("Content-Type", "multipart/form-data; boundary=broken")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestHandlerGetImportJobSuccess(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubImportService{
		getOutput: catalogimportusecase.GetImportJobOutput{
			Job: catalogimportusecase.Job{
				ID:             testAdminJobID,
				Status:         "completed",
				SourceFormat:   "csv",
				SourceFilename: "catalog.csv",
				CreatedAt:      time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC),
				UpdatedAt:      time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC),
			},
		},
	}, authhttp.NewMiddleware(stubImportAuthorizer{
		actor: authusecase.Actor{UserID: testAdminUserID, Role: "admin"},
	}), 1024)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/import-jobs/"+testAdminJobID, nil)
	req.Header.Set("Authorization", "Bearer token")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestHandlerGetImportJobNotFound(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubImportService{
		getErr: apperrors.New(apperrors.KindNotFound, "import_job_not_found", "import job not found"),
	}, authhttp.NewMiddleware(stubImportAuthorizer{
		actor: authusecase.Actor{UserID: testAdminUserID, Role: "admin"},
	}), 1024)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/import-jobs/"+testAdminJobID, nil)
	req.Header.Set("Authorization", "Bearer token")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestHandlerListImportErrorsRejectsUnknownQueryParam(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	handler, err := NewHandler(&stubImportService{}, authhttp.NewMiddleware(stubImportAuthorizer{
		actor: authusecase.Actor{UserID: testAdminUserID, Role: "admin"},
	}), 1024)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/import-jobs/"+testAdminJobID+"/errors?bad=1", nil)
	req.Header.Set("Authorization", "Bearer token")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	var response struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if response.Error.Code != "invalid_query_parameter" {
		t.Fatalf("error code = %q, want %q", response.Error.Code, "invalid_query_parameter")
	}
}

type stubImportService struct {
	runInput   catalogimportusecase.RunImportInput
	runOutput  catalogimportusecase.RunImportOutput
	runErr     error
	getOutput  catalogimportusecase.GetImportJobOutput
	getErr     error
	listOutput catalogimportusecase.ListImportErrorsOutput
	listErr    error
}

func (s *stubImportService) RunImport(_ context.Context, input catalogimportusecase.RunImportInput) (catalogimportusecase.RunImportOutput, error) {
	s.runInput = input
	return s.runOutput, s.runErr
}

func (s *stubImportService) GetImportJob(context.Context, catalogimportusecase.GetImportJobInput) (catalogimportusecase.GetImportJobOutput, error) {
	return s.getOutput, s.getErr
}

func (s *stubImportService) ListImportErrors(context.Context, catalogimportusecase.ListImportErrorsInput) (catalogimportusecase.ListImportErrorsOutput, error) {
	return s.listOutput, s.listErr
}

type stubImportAuthorizer struct {
	actor authusecase.Actor
	err   error
}

func (a stubImportAuthorizer) Authorize(context.Context, string) (authusecase.Actor, error) {
	if a.err != nil {
		return authusecase.Actor{}, a.err
	}

	return a.actor, nil
}

func multipartBody(t *testing.T, filename, content string) (*bytes.Buffer, string) {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	fileWriter, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := fileWriter.Write([]byte(content)); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	return &body, writer.FormDataContentType()
}
