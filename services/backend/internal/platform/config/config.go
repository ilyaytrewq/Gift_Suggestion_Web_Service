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
	defaultMLDialTimeout     = 3 * time.Second
)

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Database DatabaseConfig
	ML       MLConfig
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

type MLConfig struct {
	Enabled     bool
	Address     string
	DialTimeout time.Duration
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

	cfg.ML.Enabled, err = boolWithDefault(lookup, "ML_GRPC_ENABLED", false)
	if err != nil {
		return Config{}, err
	}
	cfg.ML.Address = stringWithDefault(lookup, "ML_GRPC_ADDR", "")
	cfg.ML.DialTimeout, err = durationWithDefault(lookup, "ML_GRPC_DIAL_TIMEOUT", defaultMLDialTimeout)
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
	if cfg.ML.Enabled && cfg.ML.Address == "" {
		return errors.New("ML_GRPC_ADDR is required when ML_GRPC_ENABLED=true")
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
