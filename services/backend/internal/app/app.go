package app

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	healthhttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/health/delivery/http"
	healthgrpc "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/health/infra/grpc"
	healthpostgres "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/health/infra/postgres"
	healthusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/health/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/clock"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/config"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/httpserver"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/logger"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/mlgrpc"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/postgres"
	transporthttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/transport/http"
)

type App struct {
	cfg      config.Config
	logger   *slog.Logger
	server   *httpserver.Server
	database *sql.DB
	mlClient *mlgrpc.Client
}

func Run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(cfg.App)

	application, err := New(ctx, cfg, log)
	if err != nil {
		return err
	}
	defer application.Close()

	return application.Start(ctx)
}

func New(ctx context.Context, cfg config.Config, log *slog.Logger) (*App, error) {
	if log == nil {
		return nil, apperrors.New(
			apperrors.KindInternal,
			"logger_missing",
			"logger is required",
		)
	}

	setGinMode(cfg.App.Env)

	database, err := postgres.Connect(ctx, cfg.Database)
	if err != nil {
		return nil, apperrors.Wrap(
			apperrors.KindUnavailable,
			"database_connect_failed",
			"failed to connect to postgres",
			err,
		)
	}

	if cfg.Database.MigrationsEnabled {
		if err := postgres.RunMigrations(database); err != nil {
			if closeErr := database.Close(); closeErr != nil {
				log.Error("failed to close database after migration error", "error", closeErr)
			}

			return nil, apperrors.Wrap(
				apperrors.KindInternal,
				"database_migration_failed",
				"failed to apply database migrations",
				err,
			)
		}
	}

	mlClient, err := mlgrpc.NewClient(ctx, cfg.ML)
	if err != nil {
		if closeErr := database.Close(); closeErr != nil {
			log.Error("failed to close database after ml client error", "error", closeErr)
		}

		return nil, apperrors.Wrap(
			apperrors.KindUnavailable,
			"ml_client_connect_failed",
			"failed to connect to ml service",
			err,
		)
	}

	healthService, err := healthusecase.NewService(
		clock.Real{},
		[]healthusecase.Dependency{
			{
				Name:     "postgres",
				Required: true,
				Enabled:  true,
				Probe:    healthpostgres.NewProbe(database),
			},
			{
				Name:     "ml_service",
				Required: false,
				Enabled:  mlClient.Enabled(),
				Probe:    healthgrpc.NewProbe(mlClient),
			},
		},
	)
	if err != nil {
		if closeErr := mlClient.Close(); closeErr != nil {
			log.Error("failed to close ml client after health service error", "error", closeErr)
		}
		if closeErr := database.Close(); closeErr != nil {
			log.Error("failed to close database after health service error", "error", closeErr)
		}

		return nil, err
	}

	healthHandler, err := healthhttp.NewHandler(healthService)
	if err != nil {
		if closeErr := mlClient.Close(); closeErr != nil {
			log.Error("failed to close ml client after handler error", "error", closeErr)
		}
		if closeErr := database.Close(); closeErr != nil {
			log.Error("failed to close database after handler error", "error", closeErr)
		}

		return nil, err
	}

	router := transporthttp.NewRouter(log, healthHandler)

	return &App{
		cfg:      cfg,
		logger:   log,
		server:   httpserver.New(cfg.HTTP, router),
		database: database,
		mlClient: mlClient,
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	a.server.Start(errCh)

	a.logger.Info(
		"backend started",
		"address",
		a.server.Address(),
		"env",
		a.cfg.App.Env,
	)

	select {
	case <-ctx.Done():
		a.logger.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			return apperrors.Wrap(
				apperrors.KindInternal,
				"http_server_failed",
				"http server failed",
				err,
			)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), a.cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return apperrors.Wrap(
			apperrors.KindInternal,
			"http_server_shutdown_failed",
			"failed to shutdown http server",
			err,
		)
	}

	return nil
}

func (a *App) Close() {
	if a.mlClient != nil {
		if err := a.mlClient.Close(); err != nil {
			a.logger.Error("failed to close ml client", "error", err)
		}
	}

	if a.database != nil {
		if err := a.database.Close(); err != nil {
			a.logger.Error("failed to close database", "error", err)
		}
	}
}

func setGinMode(env string) {
	switch env {
	case "production":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
}
