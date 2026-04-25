package app

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	authhttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/delivery/http"
	authpostgres "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/infra/postgres"
	authusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/usecase"
	cataloghttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/delivery/http"
	catalogpostgres "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/infra/postgres"
	catalogusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/usecase"
	catalogimporthttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalogimport/delivery/http"
	catalogimportparser "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalogimport/infra/parser"
	catalogimportpostgres "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalogimport/infra/postgres"
	catalogimportusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalogimport/usecase"
	healthhttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/health/delivery/http"
	healthgrpc "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/health/infra/grpc"
	healthpostgres "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/health/infra/postgres"
	healthusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/health/usecase"
	recommendationhttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/recommendation/delivery/http"
	recommendationgrpc "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/recommendation/infra/grpc"
	recommendationpostgres "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/recommendation/infra/postgres"
	recommendationusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/recommendation/usecase"
	trackinghttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/tracking/delivery/http"
	trackingpostgres "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/tracking/infra/postgres"
	trackingusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/tracking/usecase"
	userhttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/delivery/http"
	userpostgres "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/infra/postgres"
	userusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/usecase"
	vkintegrationhttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/delivery/http"
	vkintegrationcrypto "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/infra/crypto"
	vkintegrationpostgres "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/infra/postgres"
	vkintegrationvk "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/infra/vk"
	vkintegrationusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/usecase"
	wishlisthttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/wishlist/delivery/http"
	wishlistpostgres "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/wishlist/infra/postgres"
	wishlistusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/wishlist/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/authjwt"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/clock"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/config"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/httpserver"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/idgen"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/logger"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/mlgrpc"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/postgres"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/randomtoken"
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
			closeDatabaseWithLog(log, database, "failed to close database after migration error")

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
		closeDatabaseWithLog(log, database, "failed to close database after ml client error")

		return nil, apperrors.Wrap(
			apperrors.KindUnavailable,
			"ml_client_connect_failed",
			"failed to connect to ml service",
			err,
		)
	}

	healthHandler, err := newHealthHandler(database, mlClient)
	if err != nil {
		closeMLClientWithLog(log, mlClient, "failed to close ml client after health handler error")
		closeDatabaseWithLog(log, database, "failed to close database after health handler error")

		return nil, err
	}

	userRepository := userpostgres.NewRepository(database)
	catalogRepository := catalogpostgres.NewRepository(database)
	wishlistRepository := wishlistpostgres.NewRepository(database)
	sessionRepository := authpostgres.NewSessionRepository(database)
	passwordResetRepository := authpostgres.NewPasswordResetRepository(database)
	uuidGenerator := idgen.UUIDGenerator{}
	jwtManager := authjwt.NewManager(cfg.Auth)
	refreshTokenGenerator := randomtoken.NewGenerator(32)
	resetTokenGenerator := randomtoken.NewGenerator(32)

	authService, err := authusecase.NewService(
		userRepository,
		sessionRepository,
		passwordResetRepository,
		jwtManager,
		refreshTokenGenerator,
		resetTokenGenerator,
		uuidGenerator,
		uuidGenerator,
		uuidGenerator,
		cfg.Auth.RefreshTokenTTL,
		cfg.Auth.PasswordResetTokenTTL,
		clock.Real{},
	)
	if err != nil {
		return nil, err
	}

	userService, err := userusecase.NewService(userRepository, clock.Real{})
	if err != nil {
		return nil, err
	}
	catalogService, err := catalogusecase.NewService(catalogRepository)
	if err != nil {
		return nil, err
	}
	wishlistService, err := wishlistusecase.NewService(
		wishlistRepository,
		userRepository,
		catalogRepository,
		uuidGenerator,
		uuidGenerator,
		clock.Real{},
	)
	if err != nil {
		return nil, err
	}

	authHandler, err := authhttp.NewHandler(authService, authhttp.RefreshCookieConfig{
		Name:     cfg.Auth.RefreshCookieName,
		Path:     cfg.Auth.RefreshCookiePath,
		Domain:   cfg.Auth.RefreshCookieDomain,
		Secure:   cfg.Auth.RefreshCookieSecure,
		MaxAge:   int(cfg.Auth.RefreshTokenTTL.Seconds()),
		SameSite: http.SameSiteLaxMode,
	})
	if err != nil {
		return nil, err
	}

	authMiddleware := authhttp.NewMiddleware(authService) //nolint:contextcheck // request context is supplied inside the returned Gin handler.

	userHandler, err := userhttp.NewHandler(userService, authMiddleware)
	if err != nil {
		return nil, err
	}
	catalogHandler, err := cataloghttp.NewHandler(catalogService)
	if err != nil {
		return nil, err
	}
	wishlistHandler, err := wishlisthttp.NewHandler(wishlistService, authMiddleware)
	if err != nil {
		return nil, err
	}
	importHandler, err := newCatalogImportHandler(database, authMiddleware, cfg.Import, uuidGenerator)
	if err != nil {
		return nil, err
	}
	recommendationHandler, err := newRecommendationHandler(
		database,
		mlClient,
		authMiddleware,
		cfg.ML,
		userRepository,
		catalogRepository,
		wishlistRepository,
		uuidGenerator,
	)
	if err != nil {
		return nil, err
	}
	trackingHandler, err := newTrackingHandler(
		database,
		authMiddleware,
		userRepository,
		catalogRepository,
		wishlistRepository,
		uuidGenerator,
	)
	if err != nil {
		return nil, err
	}
	vkIntegrationHandler, err := newVKIntegrationHandler(
		database,
		authMiddleware,
		cfg.VK,
		userRepository,
		uuidGenerator,
	)
	if err != nil {
		return nil, err
	}

	router := transporthttp.NewRouter(
		log,
		healthHandler,
		authHandler,
		userHandler,
		catalogHandler,
		wishlistHandler,
		importHandler,
		recommendationHandler,
		trackingHandler,
		vkIntegrationHandler,
	)

	return &App{
		cfg:      cfg,
		logger:   log,
		server:   httpserver.New(cfg.HTTP, router),
		database: database,
		mlClient: mlClient,
	}, nil
}

func newHealthHandler(database *sql.DB, mlClient *mlgrpc.Client) (*healthhttp.Handler, error) {
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
		return nil, err
	}

	return healthhttp.NewHandler(healthService)
}

func newCatalogImportHandler(
	database *sql.DB,
	authMiddleware gin.HandlerFunc,
	cfg config.ImportConfig,
	uuidGenerator idgen.UUIDGenerator,
) (*catalogimporthttp.Handler, error) {
	importRepository := catalogimportpostgres.NewRepository(database)
	importParsers := catalogimportparser.NewRegistry()

	importService, err := catalogimportusecase.NewService(
		importRepository,
		importParsers,
		uuidGenerator,
		uuidGenerator,
		uuidGenerator,
		cfg.MaxFileSizeBytes,
		clock.Real{},
	)
	if err != nil {
		return nil, err
	}

	return catalogimporthttp.NewHandler(importService, authMiddleware, cfg.MaxFileSizeBytes)
}

func newRecommendationHandler(
	database *sql.DB,
	mlClient *mlgrpc.Client,
	authMiddleware gin.HandlerFunc,
	cfg config.MLConfig,
	userRepository *userpostgres.Repository,
	catalogRepository *catalogpostgres.Repository,
	wishlistRepository *wishlistpostgres.Repository,
	uuidGenerator idgen.UUIDGenerator,
) (*recommendationhttp.Handler, error) {
	recommendationRepository := recommendationpostgres.NewRepository(database)
	rankingGateway := recommendationgrpc.NewGateway(mlClient, cfg)

	recommendationService, err := recommendationusecase.NewService(
		recommendationRepository,
		recommendationRepository,
		userRepository,
		wishlistRepository,
		catalogRepository,
		rankingGateway,
		uuidGenerator,
		uuidGenerator,
		clock.Real{},
	)
	if err != nil {
		return nil, err
	}

	return recommendationhttp.NewHandler(recommendationService, authMiddleware)
}

func newTrackingHandler(
	database *sql.DB,
	authMiddleware gin.HandlerFunc,
	userRepository *userpostgres.Repository,
	catalogRepository *catalogpostgres.Repository,
	wishlistRepository *wishlistpostgres.Repository,
	uuidGenerator idgen.UUIDGenerator,
) (*trackinghttp.Handler, error) {
	trackingRepository := trackingpostgres.NewRepository(database)
	recommendationRepository := recommendationpostgres.NewRepository(database)

	trackingService, err := trackingusecase.NewService(
		trackingRepository,
		userRepository,
		catalogRepository,
		wishlistRepository,
		recommendationRepository,
		uuidGenerator,
		clock.Real{},
	)
	if err != nil {
		return nil, err
	}

	return trackinghttp.NewHandler(trackingService, authMiddleware)
}

func newVKIntegrationHandler(
	database *sql.DB,
	authMiddleware gin.HandlerFunc,
	cfg config.VKConfig,
	userRepository *userpostgres.Repository,
	uuidGenerator idgen.UUIDGenerator,
) (*vkintegrationhttp.Handler, error) {
	connectionRepository := vkintegrationpostgres.NewRepository(database)
	importer := vkintegrationvk.NewClient(cfg)

	var tokenProtector vkintegrationusecase.TokenProtector = vkintegrationcrypto.NewDisabledProtector()
	if strings.TrimSpace(cfg.TokenEncryptionKey) != "" {
		protector, err := vkintegrationcrypto.NewAESGCMProtector(cfg.TokenEncryptionKey)
		if err != nil {
			return nil, apperrors.Wrap(
				apperrors.KindInternal,
				"vk_token_protector_init_failed",
				"failed to initialize vk token protector",
				err,
			)
		}
		tokenProtector = protector
	}

	vkIntegrationService, err := vkintegrationusecase.NewService(
		connectionRepository,
		userRepository,
		tokenProtector,
		importer,
		uuidGenerator,
		cfg.Enabled,
		cfg.RequestTimeout,
		clock.Real{},
	)
	if err != nil {
		return nil, err
	}

	return vkintegrationhttp.NewHandler(vkIntegrationService, authMiddleware)
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

func closeDatabaseWithLog(log *slog.Logger, database *sql.DB, message string) {
	if database == nil {
		return
	}

	if err := database.Close(); err != nil {
		log.Error(message, "error", err)
	}
}

func closeMLClientWithLog(log *slog.Logger, client *mlgrpc.Client, message string) {
	if client == nil {
		return
	}

	if err := client.Close(); err != nil {
		log.Error(message, "error", err)
	}
}
