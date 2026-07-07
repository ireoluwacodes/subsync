package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	NombaEnvSandbox    = "sandbox"
	NombaEnvProduction = "production"

	NombaSandboxBaseURL    = "https://sandbox.nomba.com"
	NombaProductionBaseURL = "https://api.nomba.com"
)

type Config struct {
	AppEnv     string
	HTTPPort   string
	LogLevel   string
	PostgresDSN string
	RedisURL   string

	NombaCredentialsEncryptionKey string
	NombaWebhookSigningKey        string

	JWTSecret      string
	JWTAccessTTL   time.Duration
	JWTRefreshTTL  time.Duration

	BillingMockResult    string
	WebhookSigningSecret string
	PublicBaseURL        string
	CORSAllowedOrigins   []string

	ResendAPIKey    string
	ResendFromEmail string

	CloudinaryCloudName   string
	CloudinaryAPIKey      string
	CloudinaryAPISecret   string
	CloudinaryFolder      string

	SentryDSN string
}

func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:                        getEnv("APP_ENV", "development"),
		HTTPPort:                      getEnv("HTTP_PORT", "8080"),
		LogLevel:                      getEnv("LOG_LEVEL", "info"),
		PostgresDSN:                   buildPostgresDSN(),
		RedisURL:                      getEnv("REDIS_URL", "redis://localhost:6379/0"),
		NombaCredentialsEncryptionKey: os.Getenv("NOMBA_CREDENTIALS_ENCRYPTION_KEY"),
		NombaWebhookSigningKey:        os.Getenv("NOMBA_WEBHOOK_SIGNING_KEY"),
		JWTSecret:                     os.Getenv("JWT_SECRET"),
		BillingMockResult:             os.Getenv("BILLING_MOCK_RESULT"),
		WebhookSigningSecret:          os.Getenv("WEBHOOK_SIGNING_SECRET"),
		PublicBaseURL:                 getEnv("PUBLIC_BASE_URL", "http://localhost:8080"),
		CORSAllowedOrigins:            parseCSVEnv(os.Getenv("CORS_ALLOWED_ORIGINS")),
		ResendAPIKey:                  os.Getenv("RESEND_API_KEY"),
		ResendFromEmail:               os.Getenv("RESEND_FROM_EMAIL"),
		CloudinaryCloudName:           os.Getenv("CLOUDINARY_CLOUD_NAME"),
		CloudinaryAPIKey:              os.Getenv("CLOUDINARY_API_KEY"),
		CloudinaryAPISecret:           os.Getenv("CLOUDINARY_API_SECRET"),
		CloudinaryFolder:              getEnv("CLOUDINARY_FOLDER", "subsync/invoices"),
		SentryDSN:                     os.Getenv("SENTRY_DSN"),
	}

	var err error
	cfg.JWTAccessTTL, err = parseDuration(getEnv("JWT_ACCESS_TTL", "24h"), 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("JWT_ACCESS_TTL: %w", err)
	}
	cfg.JWTRefreshTTL, err = parseDuration(getEnv("JWT_REFRESH_TTL", "168h"), 168*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("JWT_REFRESH_TTL: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func parseDuration(raw string, fallback time.Duration) (time.Duration, error) {
	if raw == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, err
	}
	return d, nil
}

func buildPostgresDSN() string {
	if dsn := os.Getenv("POSTGRES_DSN"); dsn != "" {
		return dsn
	}

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "cierge_user")
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "cierge_pass"
	}
	name := getEnv("DB_NAME", "subsync")

	userInfo := url.UserPassword(user, password)
	return fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=disable",
		userInfo.String(), host, port, name)
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}

func (c *Config) validate() error {
	if c.AppEnv == "production" {
		if c.PostgresDSN == "" {
			return fmt.Errorf("POSTGRES_DSN is required in production")
		}
		if c.RedisURL == "" {
			return fmt.Errorf("REDIS_URL is required in production")
		}
		if c.JWTSecret == "" {
			return fmt.Errorf("JWT_SECRET is required in production")
		}
		if c.NombaCredentialsEncryptionKey == "" {
			return fmt.Errorf("NOMBA_CREDENTIALS_ENCRYPTION_KEY is required in production")
		}
		if c.BillingMockResult != "" {
			return fmt.Errorf("BILLING_MOCK_RESULT must not be set in production")
		}
		if c.PublicBaseURL == "" || !strings.HasPrefix(c.PublicBaseURL, "https://") {
			return fmt.Errorf("PUBLIC_BASE_URL must be an https URL in production")
		}
	}
	return nil
}

func (c *Config) IsDevelopment() bool {
	return strings.EqualFold(c.AppEnv, "development")
}

// DefaultCheckoutSuccessURL is the hosted return page used when integrators omit success_url.
func (c *Config) DefaultCheckoutSuccessURL() string {
	base := strings.TrimRight(c.PublicBaseURL, "/")
	if base == "" {
		base = "http://localhost:8080"
	}
	return base + "/billing/success"
}

func (c *Config) DevEncryptionKey() string {
	if c.NombaCredentialsEncryptionKey != "" {
		return c.NombaCredentialsEncryptionKey
	}
	return "0123456789abcdef0123456789abcdef"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseCSVEnv(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
