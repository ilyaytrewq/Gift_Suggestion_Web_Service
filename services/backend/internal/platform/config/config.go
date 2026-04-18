package config

import (
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAppName           = "gift-suggestion-backend"
	defaultAppEnv            = "local"
	defaultLogLevel          = "info"
	defaultHTTPHost          = "0.0.0.0"
	defaultHTTPPort          = 8080
	defaultHTTPReadTimeout   = 5 * time.Second
	defaultHTTPWriteTimeout  = 10 * time.Second
	defaultHTTPIdleTimeout   = 60 * time.Second
	defaultHTTPShutdownGrace = 10 * time.Second
	defaultDBMaxOpenConns    = 10
	defaultDBMaxIdleConns    = 5
	defaultDBConnMaxLifetime = 30 * time.Minute
	defaultDBPingTimeout     = 2 * time.Second
	defaultImportMaxFileSize = 5 * 1024 * 1024
	defaultMLDialTimeout     = 3 * time.Second
	defaultJWTIssuer         = "gift-suggestion-backend"
	defaultJWTAudience       = "gift-suggestion-web-service"
	defaultAccessTokenTTL    = 15 * time.Minute
	defaultRefreshTokenTTL   = 7 * 24 * time.Hour
	defaultResetTokenTTL     = 30 * time.Minute
	defaultRefreshCookieName = "refresh_token"
	defaultRefreshCookiePath = "/api/v1/auth"
)

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Database DatabaseConfig
	Import   ImportConfig
	ML       MLConfig
	Auth     AuthConfig
}

type AppConfig struct {
	Name     string
	Env      string
	LogLevel string
}

type HTTPConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func (cfg HTTPConfig) Address() string {
	return net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
}

type DatabaseConfig struct {
	DSN               string
	MaxOpenConns      int
	MaxIdleConns      int
	ConnMaxLifetime   time.Duration
	PingTimeout       time.Duration
	MigrationsEnabled bool
}

type ImportConfig struct {
	MaxFileSizeBytes int64
}

type MLConfig struct {
	Enabled     bool
	Address     string
	DialTimeout time.Duration
}

type AuthConfig struct {
	JWTSecret             string
	JWTIssuer             string
	JWTAudience           string
	AccessTokenTTL        time.Duration
	RefreshTokenTTL       time.Duration
	PasswordResetTokenTTL time.Duration
	RefreshCookieName     string
	RefreshCookiePath     string
	RefreshCookieDomain   string
	RefreshCookieSecure   bool
}

func Load() (Config, error) {
	return LoadFromLookup(os.LookupEnv)
}

func LoadFromLookup(lookup func(string) (string, bool)) (Config, error) {
	cfg := Config{
		App: AppConfig{
			Name:     stringWithDefault(lookup, "APP_NAME", defaultAppName),
			Env:      stringWithDefault(lookup, "APP_ENV", defaultAppEnv),
			LogLevel: stringWithDefault(lookup, "LOG_LEVEL", defaultLogLevel),
		},
	}

	httpPort, err := intWithDefault(lookup, "HTTP_PORT", defaultHTTPPort)
	if err != nil {
		return Config{}, err
	}
	httpReadTimeout, err := durationWithDefault(lookup, "HTTP_READ_TIMEOUT", defaultHTTPReadTimeout)
	if err != nil {
		return Config{}, err
	}
	httpWriteTimeout, err := durationWithDefault(lookup, "HTTP_WRITE_TIMEOUT", defaultHTTPWriteTimeout)
	if err != nil {
		return Config{}, err
	}
	httpIdleTimeout, err := durationWithDefault(lookup, "HTTP_IDLE_TIMEOUT", defaultHTTPIdleTimeout)
	if err != nil {
		return Config{}, err
	}
	httpShutdownTimeout, err := durationWithDefault(lookup, "HTTP_SHUTDOWN_TIMEOUT", defaultHTTPShutdownGrace)
	if err != nil {
		return Config{}, err
	}

	cfg.HTTP = HTTPConfig{
		Host:            stringWithDefault(lookup, "HTTP_HOST", defaultHTTPHost),
		Port:            httpPort,
		ReadTimeout:     httpReadTimeout,
		WriteTimeout:    httpWriteTimeout,
		IdleTimeout:     httpIdleTimeout,
		ShutdownTimeout: httpShutdownTimeout,
	}

	if cfg.Database.DSN, err = requiredString(lookup, "DB_DSN"); err != nil {
		return Config{}, err
	}
	cfg.Database.MaxOpenConns, err = intWithDefault(lookup, "DB_MAX_OPEN_CONNS", defaultDBMaxOpenConns)
	if err != nil {
		return Config{}, err
	}
	cfg.Database.MaxIdleConns, err = intWithDefault(lookup, "DB_MAX_IDLE_CONNS", defaultDBMaxIdleConns)
	if err != nil {
		return Config{}, err
	}
	cfg.Database.ConnMaxLifetime, err = durationWithDefault(lookup, "DB_CONN_MAX_LIFETIME", defaultDBConnMaxLifetime)
	if err != nil {
		return Config{}, err
	}
	cfg.Database.PingTimeout, err = durationWithDefault(lookup, "DB_PING_TIMEOUT", defaultDBPingTimeout)
	if err != nil {
		return Config{}, err
	}
	cfg.Database.MigrationsEnabled, err = boolWithDefault(lookup, "DB_MIGRATIONS_ENABLED", true)
	if err != nil {
		return Config{}, err
	}

	cfg.Import, err = loadImportConfig(lookup)
	if err != nil {
		return Config{}, err
	}

	cfg.ML, err = loadMLConfig(lookup)
	if err != nil {
		return Config{}, err
	}

	if cfg.Auth.JWTSecret, err = requiredString(lookup, "AUTH_JWT_SECRET"); err != nil {
		return Config{}, err
	}
	cfg.Auth.JWTIssuer = stringWithDefault(lookup, "AUTH_JWT_ISSUER", defaultJWTIssuer)
	cfg.Auth.JWTAudience = stringWithDefault(lookup, "AUTH_JWT_AUDIENCE", defaultJWTAudience)
	cfg.Auth.AccessTokenTTL, err = durationWithDefault(lookup, "AUTH_ACCESS_TTL", defaultAccessTokenTTL)
	if err != nil {
		return Config{}, err
	}
	cfg.Auth.RefreshTokenTTL, err = durationWithDefault(lookup, "AUTH_REFRESH_TTL", defaultRefreshTokenTTL)
	if err != nil {
		return Config{}, err
	}
	cfg.Auth.PasswordResetTokenTTL, err = durationWithDefault(lookup, "AUTH_PASSWORD_RESET_TTL", defaultResetTokenTTL)
	if err != nil {
		return Config{}, err
	}
	cfg.Auth.RefreshCookieName = stringWithDefault(lookup, "AUTH_REFRESH_COOKIE_NAME", defaultRefreshCookieName)
	cfg.Auth.RefreshCookiePath = stringWithDefault(lookup, "AUTH_REFRESH_COOKIE_PATH", defaultRefreshCookiePath)
	cfg.Auth.RefreshCookieDomain = stringWithDefault(lookup, "AUTH_REFRESH_COOKIE_DOMAIN", "")
	cfg.Auth.RefreshCookieSecure, err = boolWithDefault(lookup, "AUTH_REFRESH_COOKIE_SECURE", false)
	if err != nil {
		return Config{}, err
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) validate() error {
	if cfg.HTTP.Port < 1 || cfg.HTTP.Port > 65535 {
		return errors.New("HTTP_PORT must be between 1 and 65535")
	}

	switch strings.ToLower(cfg.App.LogLevel) {
	case "debug", "info", "warn", "error":
	default:
		return errors.New("LOG_LEVEL must be one of debug, info, warn, error")
	}

	if cfg.Database.MaxOpenConns < 1 {
		return errors.New("DB_MAX_OPEN_CONNS must be greater than zero")
	}
	if cfg.Database.MaxIdleConns < 0 {
		return errors.New("DB_MAX_IDLE_CONNS must be greater than or equal to zero")
	}
	if cfg.Database.MaxIdleConns > cfg.Database.MaxOpenConns {
		return errors.New("DB_MAX_IDLE_CONNS must be less than or equal to DB_MAX_OPEN_CONNS")
	}
	if cfg.Import.MaxFileSizeBytes < 1 {
		return errors.New("IMPORT_MAX_FILE_SIZE_BYTES must be greater than zero")
	}
	if cfg.ML.Enabled && cfg.ML.Address == "" {
		return errors.New("ML_GRPC_ADDR is required when ML_GRPC_ENABLED=true")
	}
	if len(cfg.Auth.JWTSecret) < 16 {
		return errors.New("AUTH_JWT_SECRET must be at least 16 characters long")
	}
	if cfg.Auth.AccessTokenTTL <= 0 {
		return errors.New("AUTH_ACCESS_TTL must be greater than zero")
	}
	if cfg.Auth.RefreshTokenTTL <= 0 {
		return errors.New("AUTH_REFRESH_TTL must be greater than zero")
	}
	if cfg.Auth.PasswordResetTokenTTL <= 0 {
		return errors.New("AUTH_PASSWORD_RESET_TTL must be greater than zero")
	}

	return nil
}

func requiredString(lookup func(string) (string, bool), key string) (string, error) {
	value, ok := lookup(key)
	if !ok {
		return "", errors.New(key + " is required")
	}

	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", errors.New(key + " must not be empty")
	}

	return trimmed, nil
}

func stringWithDefault(lookup func(string) (string, bool), key, defaultValue string) string {
	value, ok := lookup(key)
	if !ok {
		return defaultValue
	}

	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return defaultValue
	}

	return trimmed
}

func intWithDefault(lookup func(string) (string, bool), key string, defaultValue int) (int, error) {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, errors.New(key + " must be a valid integer")
	}

	return parsed, nil
}

func int64WithDefault(lookup func(string) (string, bool), key string, defaultValue int64) (int64, error) {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0, errors.New(key + " must be a valid integer")
	}

	return parsed, nil
}

func loadImportConfig(lookup func(string) (string, bool)) (ImportConfig, error) {
	maxFileSizeBytes, err := int64WithDefault(lookup, "IMPORT_MAX_FILE_SIZE_BYTES", defaultImportMaxFileSize)
	if err != nil {
		return ImportConfig{}, err
	}

	return ImportConfig{
		MaxFileSizeBytes: maxFileSizeBytes,
	}, nil
}

func loadMLConfig(lookup func(string) (string, bool)) (MLConfig, error) {
	enabled, err := boolWithDefault(lookup, "ML_GRPC_ENABLED", false)
	if err != nil {
		return MLConfig{}, err
	}

	dialTimeout, err := durationWithDefault(lookup, "ML_GRPC_DIAL_TIMEOUT", defaultMLDialTimeout)
	if err != nil {
		return MLConfig{}, err
	}

	return MLConfig{
		Enabled:     enabled,
		Address:     stringWithDefault(lookup, "ML_GRPC_ADDR", ""),
		DialTimeout: dialTimeout,
	}, nil
}

func durationWithDefault(
	lookup func(string) (string, bool),
	key string,
	defaultValue time.Duration,
) (time.Duration, error) {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return defaultValue, nil
	}

	parsed, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, errors.New(key + " must be a valid duration")
	}

	return parsed, nil
}

func boolWithDefault(lookup func(string) (string, bool), key string, defaultValue bool) (bool, error) {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, errors.New(key + " must be a valid boolean")
	}

	return parsed, nil
}
