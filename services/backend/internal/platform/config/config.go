package config

import (
	"encoding/base64"
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
	defaultHTTPReadTimeout   = 30 * time.Second
	defaultHTTPWriteTimeout  = 120 * time.Second
	defaultHTTPIdleTimeout   = 60 * time.Second
	defaultHTTPShutdownGrace = 10 * time.Second
	defaultDBMaxOpenConns    = 10
	defaultDBMaxIdleConns    = 5
	defaultDBConnMaxLifetime = 30 * time.Minute
	defaultDBPingTimeout     = 2 * time.Second
	defaultImportMaxFileSize = 5 * 1024 * 1024
	defaultMLDialTimeout     = 3 * time.Second
	defaultMLRequestTimeout  = 2500 * time.Millisecond
	defaultMLMaxRetries      = 1
	defaultVKRequestTimeout  = 3 * time.Second
	defaultEmailProvider     = "noop"
	defaultSMTPPort          = 587
	defaultEmailSendTimeout  = 10 * time.Second
	defaultJWTIssuer         = "gift-suggestion-backend"
	defaultJWTAudience       = "gift-suggestion-web-service"
	defaultAccessTokenTTL    = 15 * time.Minute
	defaultRefreshTokenTTL   = 7 * 24 * time.Hour
	defaultResetTokenTTL     = 30 * time.Minute
	defaultVerificationTTL   = 24 * time.Hour
	defaultRefreshCookieName = "refresh_token"
	defaultRefreshCookiePath = "/api/v1/auth"
)

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Database DatabaseConfig
	Import   ImportConfig
	ML       MLConfig
	VK       VKConfig
	Email    EmailConfig
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
	Enabled        bool
	Address        string
	DialTimeout    time.Duration
	RequestTimeout time.Duration
	MaxRetries     int
}

type VKConfig struct {
	Enabled            bool
	RequestTimeout     time.Duration
	TokenEncryptionKey string
	AppID              string
	OAuthRedirectURI   string
	OAuthTokenURL      string
	OAuthUserInfoURL   string
}

type EmailConfig struct {
	Enabled         bool
	Provider        string
	FromEmail       string
	FromName        string
	FrontendBaseURL string
	SendTimeout     time.Duration
	SMTP            SMTPConfig
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	UseTLS   bool
}

type AuthConfig struct {
	JWTSecret             string
	JWTIssuer             string
	JWTAudience           string
	AccessTokenTTL        time.Duration
	RefreshTokenTTL       time.Duration
	PasswordResetTokenTTL time.Duration
	EmailVerificationTTL  time.Duration
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

	httpCfg, err := loadHTTPConfig(lookup)
	if err != nil {
		return Config{}, err
	}

	databaseCfg, err := loadDatabaseConfig(lookup)
	if err != nil {
		return Config{}, err
	}

	importCfg, err := loadImportConfig(lookup)
	if err != nil {
		return Config{}, err
	}

	mlCfg, err := loadMLConfig(lookup)
	if err != nil {
		return Config{}, err
	}

	vkCfg, err := loadVKConfig(lookup)
	if err != nil {
		return Config{}, err
	}

	emailCfg, err := loadEmailConfig(lookup)
	if err != nil {
		return Config{}, err
	}

	authCfg, err := loadAuthConfig(lookup)
	if err != nil {
		return Config{}, err
	}

	cfg.HTTP = httpCfg
	cfg.Database = databaseCfg
	cfg.Import = importCfg
	cfg.ML = mlCfg
	cfg.VK = vkCfg
	cfg.Email = emailCfg
	cfg.Auth = authCfg

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) validate() error {
	if err := cfg.validateAppAndHTTP(); err != nil {
		return err
	}
	if err := cfg.validateStorageAndIntegrations(); err != nil {
		return err
	}
	if err := cfg.validateAuth(); err != nil {
		return err
	}

	return nil
}

func (cfg Config) validateAppAndHTTP() error {
	if cfg.HTTP.Port < 1 || cfg.HTTP.Port > 65535 {
		return errors.New("HTTP_PORT must be between 1 and 65535")
	}

	switch strings.ToLower(cfg.App.LogLevel) {
	case "debug", "info", "warn", "error":
	default:
		return errors.New("LOG_LEVEL must be one of debug, info, warn, error")
	}

	return nil
}

func (cfg Config) validateStorageAndIntegrations() error {
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
	if cfg.ML.RequestTimeout <= 0 {
		return errors.New("ML_GRPC_REQUEST_TIMEOUT must be greater than zero")
	}
	if cfg.ML.MaxRetries < 0 || cfg.ML.MaxRetries > 3 {
		return errors.New("ML_GRPC_MAX_RETRIES must be between 0 and 3")
	}
	if cfg.VK.RequestTimeout <= 0 {
		return errors.New("VK_REQUEST_TIMEOUT must be greater than zero")
	}
	if strings.TrimSpace(cfg.VK.TokenEncryptionKey) != "" {
		if err := validateBase64AES256Key(cfg.VK.TokenEncryptionKey); err != nil {
			return err
		}
	}
	if err := cfg.validateEmail(); err != nil {
		return err
	}

	return nil
}

func (cfg Config) validateEmail() error {
	if cfg.Email.SendTimeout <= 0 {
		return errors.New("EMAIL_SEND_TIMEOUT must be greater than zero")
	}
	if strings.TrimSpace(cfg.Email.FromEmail) == "" {
		return errors.New("EMAIL_FROM_EMAIL must not be empty")
	}
	switch strings.ToLower(cfg.Email.Provider) {
	case "noop", "smtp":
	default:
		return errors.New("EMAIL_PROVIDER must be one of noop, smtp")
	}
	if cfg.Email.Enabled && cfg.Email.Provider == "smtp" {
		if strings.TrimSpace(cfg.Email.SMTP.Host) == "" {
			return errors.New("SMTP_HOST is required when EMAIL_PROVIDER=smtp and EMAIL_ENABLED=true")
		}
		if cfg.Email.SMTP.Port < 1 || cfg.Email.SMTP.Port > 65535 {
			return errors.New("SMTP_PORT must be between 1 and 65535 when EMAIL_PROVIDER=smtp and EMAIL_ENABLED=true")
		}
	}

	return nil
}

func (cfg Config) validateAuth() error {
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
	if cfg.Auth.EmailVerificationTTL <= 0 {
		return errors.New("AUTH_EMAIL_VERIFICATION_TTL must be greater than zero")
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

func loadHTTPConfig(lookup func(string) (string, bool)) (HTTPConfig, error) {
	httpPort, err := intWithDefault(lookup, "HTTP_PORT", defaultHTTPPort)
	if err != nil {
		return HTTPConfig{}, err
	}
	httpReadTimeout, err := durationWithDefault(lookup, "HTTP_READ_TIMEOUT", defaultHTTPReadTimeout)
	if err != nil {
		return HTTPConfig{}, err
	}
	httpWriteTimeout, err := durationWithDefault(lookup, "HTTP_WRITE_TIMEOUT", defaultHTTPWriteTimeout)
	if err != nil {
		return HTTPConfig{}, err
	}
	httpIdleTimeout, err := durationWithDefault(lookup, "HTTP_IDLE_TIMEOUT", defaultHTTPIdleTimeout)
	if err != nil {
		return HTTPConfig{}, err
	}
	httpShutdownTimeout, err := durationWithDefault(lookup, "HTTP_SHUTDOWN_TIMEOUT", defaultHTTPShutdownGrace)
	if err != nil {
		return HTTPConfig{}, err
	}

	return HTTPConfig{
		Host:            stringWithDefault(lookup, "HTTP_HOST", defaultHTTPHost),
		Port:            httpPort,
		ReadTimeout:     httpReadTimeout,
		WriteTimeout:    httpWriteTimeout,
		IdleTimeout:     httpIdleTimeout,
		ShutdownTimeout: httpShutdownTimeout,
	}, nil
}

func loadDatabaseConfig(lookup func(string) (string, bool)) (DatabaseConfig, error) {
	dsn, err := requiredString(lookup, "DB_DSN")
	if err != nil {
		return DatabaseConfig{}, err
	}
	maxOpenConns, err := intWithDefault(lookup, "DB_MAX_OPEN_CONNS", defaultDBMaxOpenConns)
	if err != nil {
		return DatabaseConfig{}, err
	}
	maxIdleConns, err := intWithDefault(lookup, "DB_MAX_IDLE_CONNS", defaultDBMaxIdleConns)
	if err != nil {
		return DatabaseConfig{}, err
	}
	connMaxLifetime, err := durationWithDefault(lookup, "DB_CONN_MAX_LIFETIME", defaultDBConnMaxLifetime)
	if err != nil {
		return DatabaseConfig{}, err
	}
	pingTimeout, err := durationWithDefault(lookup, "DB_PING_TIMEOUT", defaultDBPingTimeout)
	if err != nil {
		return DatabaseConfig{}, err
	}
	migrationsEnabled, err := boolWithDefault(lookup, "DB_MIGRATIONS_ENABLED", true)
	if err != nil {
		return DatabaseConfig{}, err
	}

	return DatabaseConfig{
		DSN:               dsn,
		MaxOpenConns:      maxOpenConns,
		MaxIdleConns:      maxIdleConns,
		ConnMaxLifetime:   connMaxLifetime,
		PingTimeout:       pingTimeout,
		MigrationsEnabled: migrationsEnabled,
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
	requestTimeout, err := durationWithDefault(lookup, "ML_GRPC_REQUEST_TIMEOUT", defaultMLRequestTimeout)
	if err != nil {
		return MLConfig{}, err
	}
	maxRetries, err := intWithDefault(lookup, "ML_GRPC_MAX_RETRIES", defaultMLMaxRetries)
	if err != nil {
		return MLConfig{}, err
	}

	return MLConfig{
		Enabled:        enabled,
		Address:        stringWithDefault(lookup, "ML_GRPC_ADDR", ""),
		DialTimeout:    dialTimeout,
		RequestTimeout: requestTimeout,
		MaxRetries:     maxRetries,
	}, nil
}

func loadVKConfig(lookup func(string) (string, bool)) (VKConfig, error) {
	enabled, err := boolWithDefault(lookup, "VK_ENABLED", false)
	if err != nil {
		return VKConfig{}, err
	}

	requestTimeout, err := durationWithDefault(lookup, "VK_REQUEST_TIMEOUT", defaultVKRequestTimeout)
	if err != nil {
		return VKConfig{}, err
	}

	return VKConfig{
		Enabled:            enabled,
		RequestTimeout:     requestTimeout,
		TokenEncryptionKey: stringWithDefault(lookup, "VK_TOKEN_ENCRYPTION_KEY", ""),
		AppID:              stringWithDefault(lookup, "VK_APP_ID", ""),
		OAuthRedirectURI:   stringWithDefault(lookup, "VK_OAUTH_REDIRECT_URI", ""),
		OAuthTokenURL:      stringWithDefault(lookup, "VK_OAUTH_TOKEN_URL", ""),
		OAuthUserInfoURL:   stringWithDefault(lookup, "VK_OAUTH_USER_INFO_URL", ""),
	}, nil
}

func loadEmailConfig(lookup func(string) (string, bool)) (EmailConfig, error) {
	enabled, err := boolWithDefault(lookup, "EMAIL_ENABLED", false)
	if err != nil {
		return EmailConfig{}, err
	}
	sendTimeout, err := durationWithDefault(lookup, "EMAIL_SEND_TIMEOUT", defaultEmailSendTimeout)
	if err != nil {
		return EmailConfig{}, err
	}
	smtpPort, err := intWithDefault(lookup, "SMTP_PORT", defaultSMTPPort)
	if err != nil {
		return EmailConfig{}, err
	}
	useTLS, err := boolWithDefault(lookup, "SMTP_USE_TLS", true)
	if err != nil {
		return EmailConfig{}, err
	}

	return EmailConfig{
		Enabled:         enabled,
		Provider:        strings.ToLower(stringWithDefault(lookup, "EMAIL_PROVIDER", defaultEmailProvider)),
		FromEmail:       stringWithDefault(lookup, "EMAIL_FROM_EMAIL", ""),
		FromName:        stringWithDefault(lookup, "EMAIL_FROM_NAME", ""),
		FrontendBaseURL: strings.TrimRight(stringWithDefault(lookup, "FRONTEND_BASE_URL", ""), "/"),
		SendTimeout:     sendTimeout,
		SMTP: SMTPConfig{
			Host:     stringWithDefault(lookup, "SMTP_HOST", ""),
			Port:     smtpPort,
			Username: stringWithDefault(lookup, "SMTP_USERNAME", ""),
			Password: stringWithDefault(lookup, "SMTP_PASSWORD", ""),
			UseTLS:   useTLS,
		},
	}, nil
}

func loadAuthConfig(lookup func(string) (string, bool)) (AuthConfig, error) {
	jwtSecret, err := requiredString(lookup, "AUTH_JWT_SECRET")
	if err != nil {
		return AuthConfig{}, err
	}
	accessTokenTTL, err := durationWithDefault(lookup, "AUTH_ACCESS_TTL", defaultAccessTokenTTL)
	if err != nil {
		return AuthConfig{}, err
	}
	refreshTokenTTL, err := durationWithDefault(lookup, "AUTH_REFRESH_TTL", defaultRefreshTokenTTL)
	if err != nil {
		return AuthConfig{}, err
	}
	passwordResetTokenTTL, err := durationWithDefault(lookup, "AUTH_PASSWORD_RESET_TTL", defaultResetTokenTTL)
	if err != nil {
		return AuthConfig{}, err
	}
	emailVerificationTTL, err := durationWithDefault(lookup, "AUTH_EMAIL_VERIFICATION_TTL", defaultVerificationTTL)
	if err != nil {
		return AuthConfig{}, err
	}
	refreshCookieSecure, err := boolWithDefault(lookup, "AUTH_REFRESH_COOKIE_SECURE", false)
	if err != nil {
		return AuthConfig{}, err
	}

	return AuthConfig{
		JWTSecret:             jwtSecret,
		JWTIssuer:             stringWithDefault(lookup, "AUTH_JWT_ISSUER", defaultJWTIssuer),
		JWTAudience:           stringWithDefault(lookup, "AUTH_JWT_AUDIENCE", defaultJWTAudience),
		AccessTokenTTL:        accessTokenTTL,
		RefreshTokenTTL:       refreshTokenTTL,
		PasswordResetTokenTTL: passwordResetTokenTTL,
		EmailVerificationTTL:  emailVerificationTTL,
		RefreshCookieName:     stringWithDefault(lookup, "AUTH_REFRESH_COOKIE_NAME", defaultRefreshCookieName),
		RefreshCookiePath:     stringWithDefault(lookup, "AUTH_REFRESH_COOKIE_PATH", defaultRefreshCookiePath),
		RefreshCookieDomain:   stringWithDefault(lookup, "AUTH_REFRESH_COOKIE_DOMAIN", ""),
		RefreshCookieSecure:   refreshCookieSecure,
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

func validateBase64AES256Key(raw string) error {
	decoded, err := decodeBase64(raw)
	if err != nil {
		return errors.New("VK_TOKEN_ENCRYPTION_KEY must be a valid base64 string")
	}
	if len(decoded) != 32 {
		return errors.New("VK_TOKEN_ENCRYPTION_KEY must decode to 32 bytes")
	}

	return nil
}

func decodeBase64(raw string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(strings.TrimSpace(raw))
}
