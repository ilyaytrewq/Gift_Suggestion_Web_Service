package http

import (
	"context"
	"errors"
	"io"
	nethttp "net/http"
	"strings"

	"github.com/gin-gonic/gin"

	authhttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/delivery/http"
	catalogimportusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalogimport/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
	httpapi "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/transport/httpapi"
)

var ErrNilImportService = errors.New("import service is nil")

type service interface {
	RunImport(ctx context.Context, input catalogimportusecase.RunImportInput) (catalogimportusecase.RunImportOutput, error)
	GetImportJob(ctx context.Context, input catalogimportusecase.GetImportJobInput) (catalogimportusecase.GetImportJobOutput, error)
	ListImportErrors(ctx context.Context, input catalogimportusecase.ListImportErrorsInput) (catalogimportusecase.ListImportErrorsOutput, error)
}

type Handler struct {
	service          service
	authMiddleware   gin.HandlerFunc
	maxFileSizeBytes int64
}

func NewHandler(service service, authMiddleware gin.HandlerFunc, maxFileSizeBytes int64) (*Handler, error) {
	if service == nil {
		return nil, ErrNilImportService
	}

	return &Handler{
		service:          service,
		authMiddleware:   authMiddleware,
		maxFileSizeBytes: maxFileSizeBytes,
	}, nil
}

func (h *Handler) Register(root gin.IRouter) {
	admin := root.Group("/admin/import-jobs")
	admin.Use(h.authMiddleware, authhttp.RequireAdmin())
	admin.POST("", h.runImport)
	admin.GET("/:job_id", h.getImportJob)
	admin.GET("/:job_id/errors", h.listImportErrors)
}

func (h *Handler) runImport(c *gin.Context) {
	if !strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		httpapi.Fail(c, apperrors.New(
			apperrors.KindValidation,
			"invalid_content_type",
			"content type must be multipart/form-data",
		))
		return
	}

	if h.maxFileSizeBytes > 0 {
		c.Request.Body = nethttp.MaxBytesReader(c.Writer, c.Request.Body, h.maxFileSizeBytes)
	}

	if err := c.Request.ParseMultipartForm(h.maxFileSizeBytes); err != nil {
		var maxBytesErr *nethttp.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpapi.Fail(c, apperrors.New(
				apperrors.KindValidation,
				"file_too_large",
				"import file exceeds size limit",
			))
			return
		}

		httpapi.Fail(c, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_request_body",
			"multipart request is invalid",
			err,
		))
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		httpapi.Fail(c, apperrors.New(
			apperrors.KindValidation,
			"file_required",
			"file is required",
		))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		httpapi.Fail(c, err)
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			_ = c.Error(closeErr)
		}
	}()

	payload, err := io.ReadAll(file)
	if err != nil {
		httpapi.Fail(c, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_request_body",
			"failed to read import file",
			err,
		))
		return
	}

	actor, _ := authhttp.ActorFromContext(c)
	output, err := h.service.RunImport(c.Request.Context(), catalogimportusecase.RunImportInput{
		RequestedByUserID: actor.UserID,
		Filename:          fileHeader.Filename,
		SourceLabel:       strings.TrimSpace(c.PostForm("source")),
		FileSizeBytes:     int64(len(payload)),
		File:              payload,
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, nethttp.StatusCreated, output)
}

func (h *Handler) getImportJob(c *gin.Context) {
	output, err := h.service.GetImportJob(c.Request.Context(), catalogimportusecase.GetImportJobInput{
		JobID: c.Param("job_id"),
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, nethttp.StatusOK, output)
}

func (h *Handler) listImportErrors(c *gin.Context) {
	input, err := parseListImportErrorsInput(c.Request.URL.Query())
	if err != nil {
		httpapi.Fail(c, err)
		return
	}
	input.JobID = c.Param("job_id")

	output, err := h.service.ListImportErrors(c.Request.Context(), input)
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, nethttp.StatusOK, output)
}
