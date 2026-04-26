package config

import (
	"testing"
	"time"
)

func TestLoadFromLookupDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := LoadFromLookup(mapLookup(map[string]string{
		"DB_DSN":           "postgres://gift:gift@localhost:5432/gift_suggestion?sslmode=disable",
		"AUTH_JWT_SECRET":  "very-secret-key-123",
		"EMAIL_FROM_EMAIL": "noreply@example.com",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.App.Name != defaultAppName {
		t.Fatalf("expected default app name %q, got %q", defaultAppName, cfg.App.Name)
	}
	if cfg.HTTP.Port != defaultHTTPPort {
		t.Fatalf("expected default port %d, got %d", defaultHTTPPort, cfg.HTTP.Port)
	}
	if !cfg.Database.MigrationsEnabled {
		t.Fatal("expected migrations to be enabled by default")
	}
	if cfg.Import.MaxFileSizeBytes != defaultImportMaxFileSize {
		t.Fatalf("expected default import max file size %d, got %d", defaultImportMaxFileSize, cfg.Import.MaxFileSizeBytes)
	}
	if cfg.ML.Enabled {
		t.Fatal("expected ml grpc to be disabled by default")
	}
	if cfg.ML.RequestTimeout != defaultMLRequestTimeout {
		t.Fatalf("expected default ml request timeout %s, got %s", defaultMLRequestTimeout, cfg.ML.RequestTimeout)
	}
	if cfg.ML.MaxRetries != defaultMLMaxRetries {
		t.Fatalf("expected default ml max retries %d, got %d", defaultMLMaxRetries, cfg.ML.MaxRetries)
	}
	if cfg.VK.Enabled {
		t.Fatal("expected vk integration to be disabled by default")
	}
	if cfg.VK.RequestTimeout != defaultVKRequestTimeout {
		t.Fatalf("expected default vk request timeout %s, got %s", defaultVKRequestTimeout, cfg.VK.RequestTimeout)
	}
	if cfg.Email.Provider != defaultEmailProvider {
		t.Fatalf("expected default email provider %q, got %q", defaultEmailProvider, cfg.Email.Provider)
	}
	if cfg.Email.SendTimeout != defaultEmailSendTimeout {
		t.Fatalf("expected default email send timeout %s, got %s", defaultEmailSendTimeout, cfg.Email.SendTimeout)
	}
	if cfg.Auth.RefreshCookieName != defaultRefreshCookieName {
		t.Fatalf("expected default refresh cookie name %q, got %q", defaultRefreshCookieName, cfg.Auth.RefreshCookieName)
	}
	if cfg.Auth.EmailVerificationTTL != defaultVerificationTTL {
		t.Fatalf("expected default email verification ttl %s, got %s", defaultVerificationTTL, cfg.Auth.EmailVerificationTTL)
	}
}

func TestLoadFromLookupRequiresDatabaseDSN(t *testing.T) {
	t.Parallel()

	if _, err := LoadFromLookup(mapLookup(nil)); err == nil {
		t.Fatal("expected error when DB_DSN is missing")
	}
}

func TestLoadFromLookupValidatesValues(t *testing.T) {
	t.Parallel()

	_, err := LoadFromLookup(mapLookup(map[string]string{
		"DB_DSN":                     "postgres://gift:gift@localhost:5432/gift_suggestion?sslmode=disable",
		"AUTH_JWT_SECRET":            "very-secret-key-123",
		"HTTP_PORT":                  "invalid",
		"ML_GRPC_ENABLED":            "true",
		"ML_GRPC_DIAL_TIMEOUT":       "5s",
		"ML_GRPC_REQUEST_TIMEOUT":    "0s",
		"ML_GRPC_MAX_RETRIES":        "1",
		"DB_CONN_MAX_LIFETIME":       "30m",
		"DB_PING_TIMEOUT":            "2s",
		"HTTP_READ_TIMEOUT":          "5s",
		"HTTP_WRITE_TIMEOUT":         "10s",
		"HTTP_IDLE_TIMEOUT":          "60s",
		"HTTP_SHUTDOWN_TIMEOUT":      "10s",
		"DB_MAX_OPEN_CONNS":          "10",
		"DB_MAX_IDLE_CONNS":          "5",
		"IMPORT_MAX_FILE_SIZE_BYTES": "0",
		"ML_GRPC_ADDR":               "localhost:50051",
		"VK_ENABLED":                 "true",
		"VK_REQUEST_TIMEOUT":         "4s",
		"VK_TOKEN_ENCRYPTION_KEY":    "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=",
		"EMAIL_ENABLED":              "true",
		"EMAIL_PROVIDER":             "smtp",
		"EMAIL_FROM_EMAIL":           "noreply@example.com",
		"SMTP_HOST":                  "smtp.example.com",
		"SMTP_PORT":                  "587",
	}))
	if err == nil {
		t.Fatal("expected invalid HTTP_PORT to fail")
	}
}

func TestLoadFromLookupReadsExplicitValues(t *testing.T) {
	t.Parallel()

	cfg, err := LoadFromLookup(mapLookup(map[string]string{
		"APP_NAME":                    "gift-svc",
		"APP_ENV":                     "production",
		"LOG_LEVEL":                   "warn",
		"HTTP_HOST":                   "127.0.0.1",
		"HTTP_PORT":                   "9090",
		"HTTP_READ_TIMEOUT":           "6s",
		"HTTP_WRITE_TIMEOUT":          "11s",
		"HTTP_IDLE_TIMEOUT":           "61s",
		"HTTP_SHUTDOWN_TIMEOUT":       "12s",
		"DB_DSN":                      "postgres://gift:gift@localhost:5432/gift_suggestion?sslmode=disable",
		"AUTH_JWT_SECRET":             "very-secret-key-123",
		"DB_MAX_OPEN_CONNS":           "12",
		"DB_MAX_IDLE_CONNS":           "6",
		"DB_CONN_MAX_LIFETIME":        "45m",
		"DB_PING_TIMEOUT":             "3s",
		"DB_MIGRATIONS_ENABLED":       "false",
		"IMPORT_MAX_FILE_SIZE_BYTES":  "2097152",
		"ML_GRPC_ENABLED":             "true",
		"ML_GRPC_ADDR":                "localhost:50051",
		"ML_GRPC_DIAL_TIMEOUT":        "7s",
		"ML_GRPC_REQUEST_TIMEOUT":     "2500ms",
		"ML_GRPC_MAX_RETRIES":         "1",
		"VK_ENABLED":                  "true",
		"VK_REQUEST_TIMEOUT":          "4s",
		"VK_TOKEN_ENCRYPTION_KEY":     "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=",
		"EMAIL_ENABLED":               "true",
		"EMAIL_PROVIDER":              "smtp",
		"EMAIL_FROM_EMAIL":            "noreply@example.com",
		"EMAIL_FROM_NAME":             "Gift Suggestion",
		"EMAIL_SEND_TIMEOUT":          "8s",
		"FRONTEND_BASE_URL":           "http://localhost:5173",
		"SMTP_HOST":                   "smtp.example.com",
		"SMTP_PORT":                   "2525",
		"SMTP_USERNAME":               "mailer",
		"SMTP_PASSWORD":               "secret",
		"SMTP_USE_TLS":                "true",
		"AUTH_JWT_ISSUER":             "gift-api",
		"AUTH_JWT_AUDIENCE":           "gift-web",
		"AUTH_ACCESS_TTL":             "20m",
		"AUTH_REFRESH_TTL":            "168h",
		"AUTH_PASSWORD_RESET_TTL":     "45m",
		"AUTH_EMAIL_VERIFICATION_TTL": "72h",
		"AUTH_REFRESH_COOKIE_NAME":    "gift_refresh",
		"AUTH_REFRESH_COOKIE_PATH":    "/api/v1/auth",
		"AUTH_REFRESH_COOKIE_SECURE":  "true",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.App.Name != "gift-svc" {
		t.Fatalf("expected app name gift-svc, got %s", cfg.App.Name)
	}
	if cfg.HTTP.Address() != "127.0.0.1:9090" {
		t.Fatalf("unexpected http address: %s", cfg.HTTP.Address())
	}
	if cfg.Database.MigrationsEnabled {
		t.Fatal("expected migrations to be disabled")
	}
	if cfg.Import.MaxFileSizeBytes != 2097152 {
		t.Fatalf("expected import max file size 2097152, got %d", cfg.Import.MaxFileSizeBytes)
	}
	if !cfg.ML.Enabled {
		t.Fatal("expected ml grpc to be enabled")
	}
	if cfg.ML.DialTimeout != 7*time.Second {
		t.Fatalf("expected ml dial timeout 7s, got %s", cfg.ML.DialTimeout)
	}
	if cfg.ML.RequestTimeout != 2500*time.Millisecond {
		t.Fatalf("expected ml request timeout 2500ms, got %s", cfg.ML.RequestTimeout)
	}
	if cfg.ML.MaxRetries != 1 {
		t.Fatalf("expected ml max retries 1, got %d", cfg.ML.MaxRetries)
	}
	if !cfg.VK.Enabled {
		t.Fatal("expected vk integration to be enabled")
	}
	if cfg.VK.RequestTimeout != 4*time.Second {
		t.Fatalf("expected vk request timeout 4s, got %s", cfg.VK.RequestTimeout)
	}
	if cfg.VK.TokenEncryptionKey == "" {
		t.Fatal("expected vk token encryption key to be loaded")
	}
	if !cfg.Email.Enabled {
		t.Fatal("expected email to be enabled")
	}
	if cfg.Email.Provider != "smtp" {
		t.Fatalf("expected email provider smtp, got %s", cfg.Email.Provider)
	}
	if cfg.Email.SMTP.Port != 2525 {
		t.Fatalf("expected smtp port 2525, got %d", cfg.Email.SMTP.Port)
	}
	if cfg.Email.SendTimeout != 8*time.Second {
		t.Fatalf("expected email send timeout 8s, got %s", cfg.Email.SendTimeout)
	}
	if cfg.Email.FrontendBaseURL != "http://localhost:5173" {
		t.Fatalf("unexpected frontend base url: %s", cfg.Email.FrontendBaseURL)
	}
	if cfg.Auth.JWTIssuer != "gift-api" {
		t.Fatalf("expected auth issuer gift-api, got %s", cfg.Auth.JWTIssuer)
	}
	if cfg.Auth.EmailVerificationTTL != 72*time.Hour {
		t.Fatalf("expected auth email verification ttl 72h, got %s", cfg.Auth.EmailVerificationTTL)
	}
	if !cfg.Auth.RefreshCookieSecure {
		t.Fatal("expected refresh cookie secure to be true")
	}
}

func TestLoadFromLookupAllowsVKWithoutEncryptionKey(t *testing.T) {
	t.Parallel()

	cfg, err := LoadFromLookup(mapLookup(map[string]string{
		"DB_DSN":           "postgres://gift:gift@localhost:5432/gift_suggestion?sslmode=disable",
		"AUTH_JWT_SECRET":  "very-secret-key-123",
		"EMAIL_FROM_EMAIL": "noreply@example.com",
		"VK_ENABLED":       "true",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.VK.Enabled {
		t.Fatal("expected vk integration to be enabled")
	}
	if cfg.VK.TokenEncryptionKey != "" {
		t.Fatalf("expected empty vk token encryption key, got %q", cfg.VK.TokenEncryptionKey)
	}
}

func TestLoadFromLookupRejectsInvalidVKEncryptionKey(t *testing.T) {
	t.Parallel()

	_, err := LoadFromLookup(mapLookup(map[string]string{
		"DB_DSN":                  "postgres://gift:gift@localhost:5432/gift_suggestion?sslmode=disable",
		"AUTH_JWT_SECRET":         "very-secret-key-123",
		"EMAIL_FROM_EMAIL":        "noreply@example.com",
		"VK_TOKEN_ENCRYPTION_KEY": "invalid",
	}))
	if err == nil {
		t.Fatal("expected invalid vk encryption key to fail")
	}
}

func mapLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
