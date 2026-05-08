package http

import (
	"context"
	"errors"
	"io"
	"log/slog"
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
	logger           *slog.Logger
}

func NewHandler(
	service service,
	authMiddleware gin.HandlerFunc,
	maxFileSizeBytes int64,
	logger *slog.Logger,
) (*Handler, error) {
	if service == nil {
		return nil, ErrNilImportService
	}

	return &Handler{
		service:          service,
		authMiddleware:   authMiddleware,
		maxFileSizeBytes: maxFileSizeBytes,
		logger:           logger,
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
	actor, _ := authhttp.ActorFromContext(c)

	if !strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		err := apperrors.New(
			apperrors.KindValidation,
			"invalid_content_type",
			"content type must be multipart/form-data",
		)
		h.logCatalogImportRejected(c, actor.UserID, "", err)
		httpapi.Fail(c, err)
		return
	}

	if h.maxFileSizeBytes > 0 {
		c.Request.Body = nethttp.MaxBytesReader(c.Writer, c.Request.Body, h.maxFileSizeBytes)
	}

	if err := c.Request.ParseMultipartForm(h.maxFileSizeBytes); err != nil {
		var maxBytesErr *nethttp.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			e := apperrors.New(
				apperrors.KindValidation,
				"file_too_large",
				"import file exceeds size limit",
			)
			h.logCatalogImportRejected(c, actor.UserID, "", e)
			httpapi.Fail(c, e)
			return
		}

		e := apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_request_body",
			"multipart request is invalid",
			err,
		)
		h.logCatalogImportRejected(c, actor.UserID, "", e)
		httpapi.Fail(c, e)
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		e := apperrors.New(
			apperrors.KindValidation,
			"file_required",
			"file is required",
		)
		h.logCatalogImportRejected(c, actor.UserID, "", e)
		httpapi.Fail(c, e)
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		e := apperrors.Wrap(
			apperrors.KindValidation,
			"import_file_read_failed",
			"failed to read uploaded import file",
			err,
		)
		h.logCatalogImportRejected(c, actor.UserID, fileHeader.Filename, e)
		httpapi.Fail(c, e)
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			_ = c.Error(closeErr)
		}
	}()

	payload, err := io.ReadAll(file)
	if err != nil {
		e := apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_request_body",
			"failed to read import file",
			err,
		)
		h.logCatalogImportRejected(c, actor.UserID, fileHeader.Filename, e)
		httpapi.Fail(c, e)
		return
	}

	output, err := h.service.RunImport(c.Request.Context(), catalogimportusecase.RunImportInput{
		RequestedByUserID: actor.UserID,
		Filename:          fileHeader.Filename,
		SourceLabel:       strings.TrimSpace(c.PostForm("source")),
		FileSizeBytes:     int64(len(payload)),
		File:              payload,
	})
	if err != nil {
		h.logCatalogImportRejected(c, actor.UserID, fileHeader.Filename, err)
		httpapi.Fail(c, err)
		return
	}

	if output.Job.Status == "failed" {
		h.logCatalogImportJobFailed(
			c,
			actor.UserID,
			fileHeader.Filename,
			output.Job.ID,
			output.Job.FailureCode,
			output.Job.FailureMessage,
		)
	}

	httpapi.Success(c, nethttp.StatusCreated, output)
}

func (h *Handler) logCatalogImportRejected(c *gin.Context, userID, filename string, err error) {
	if h.logger == nil || err == nil {
		return
	}

	appErr := apperrors.From(err)
	h.logger.Warn("catalog import rejected",
		"request_id", httpapi.RequestIDFromContext(c),
		"user_id", strings.TrimSpace(userID),
		"filename", strings.TrimSpace(filename),
		"error_code", appErr.Code(),
		"error_kind", appErr.Kind(),
		"err", err,
	)
}

func (h *Handler) logCatalogImportJobFailed(
	c *gin.Context,
	userID string,
	filename string,
	jobID string,
	failureCode *string,
	failureMsg *string,
) {
	if h.logger == nil {
		return
	}

	fc := ""
	if failureCode != nil {
		fc = *failureCode
	}
	fm := ""
	if failureMsg != nil {
		fm = *failureMsg
	}

	h.logger.Warn("catalog import job failed",
		"request_id", httpapi.RequestIDFromContext(c),
		"user_id", strings.TrimSpace(userID),
		"filename", strings.TrimSpace(filename),
		"job_id", jobID,
		"failure_code", fc,
		"failure_detail", fm,
	)
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
