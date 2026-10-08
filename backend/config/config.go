package config

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// App
	AppName    string
	AppEnv     string
	AppPort    string
	AppBaseURL string

	// Database
	DBDriver   string
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBTimezone string
	DBSSLMode  string // disable | require | verify-ca | verify-full

	// Redis Config
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// SMTP Config
	SMTPHost        string
	SMTPPort        int
	SMTPUser        string
	SMTPPassword    string
	SMTPSenderEmail string

	// MinIO Config
	MinioEndpoint   string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioBucket     string
	MinioUseSSL     bool
	MinioAllowedIPs string // Comma separated IP CIDRs

	// Encryption
	SettingEncryptionKey string // Base64-encoded 32-byte AES-256 key

	// Web Push (VAPID) — generate a pair once with webpush.GenerateVAPIDKeys
	// and keep VAPIDPrivateKey secret. Empty keys disable push sending.
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string // e.g. "mailto:ops@example.com"

	// Kafka Config
	KafkaBrokers       string
	KafkaConsumerGroup string

	// OpenSearch Config
	OpenSearchURL      string
	OpenSearchUsername string
	OpenSearchPassword string

	// JWT Config — AccessTokenTTL is intentionally short (minutes); session
	// longevity comes from the separate, revocable refresh token below.
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration

	// Auth cookies. CookieSecure MUST be true in any deployment served over
	// HTTPS — browsers silently drop `Secure` cookies over plain HTTP, so
	// leaving this true against a plain-HTTP origin breaks login entirely
	// (the Set-Cookie is sent but never stored). Set it to false ONLY for a
	// deployment that genuinely has no TLS yet (see docker-compose.proxy.yml
	// to add TLS termination), and flip it back to true the moment it does.
	CookieSecure   bool
	CookieDomain   string // empty = host-only cookie (recommended default)
	CookieSameSite string // "Lax" (default) | "Strict" | "None"

	// Storage
	StorageDriver    string
	StorageLocalPath string
	MaxUploadSizeMB  int64

	// CORS
	CORSAllowedOrigins string

	// Logger
	LogLevel    string
	LogOutput   string
	LogFilePath string
}

// Load reads .env file (if present) then loads values from environment variables.
func Load() *Config {
	// Load .env — no error if file is missing (production uses real env vars)
	if err := godotenv.Load(); err != nil {
		log.Println("[config] .env not found, reading from OS environment")
	}

	maxUpload, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_SIZE_MB", "10"), 10, 64)

	return &Config{
		AppName:    getEnv("APP_NAME", "MonitoringAudit"),
		AppEnv:     getEnv("APP_ENV", "development"),
		AppPort:    getEnv("APP_PORT", "8080"),
		AppBaseURL: getEnv("APP_BASE_URL", "http://localhost:3000"),

		DBDriver:   getEnv("DB_DRIVER", "postgres"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5434"),
		DBName:     getEnv("DB_NAME", "monitoring_audit"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBTimezone: getEnv("DB_TIMEZONE", "Asia/Jakarta"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		SMTPHost:        getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:        getEnvAsInt("SMTP_PORT", 587),
		SMTPUser:        getEnv("SMTP_USER", ""),
		SMTPPassword:    getEnv("SMTP_PASSWORD", ""),
		SMTPSenderEmail: getEnv("SMTP_SENDER_EMAIL", "noreply@monitoring-audit.local"),

		MinioEndpoint:   getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey:  getEnv("MINIO_ACCESS_KEY", ""),
		MinioSecretKey:  getEnv("MINIO_SECRET_KEY", ""),
		MinioBucket:     getEnv("MINIO_BUCKET", "monitoring-audit-bucket"),
		MinioUseSSL:     getEnv("MINIO_USE_SSL", "false") == "true",
		MinioAllowedIPs: getEnv("MINIO_ALLOWED_IPS", "127.0.0.1/32,10.0.0.0/8,192.168.0.0/16"),

		SettingEncryptionKey: getEnv("SETTING_ENCRYPTION_KEY", ""),

		VAPIDPublicKey:  getEnv("VAPID_PUBLIC_KEY", ""),
		VAPIDPrivateKey: getEnv("VAPID_PRIVATE_KEY", ""),
		VAPIDSubject:    getEnv("VAPID_SUBJECT", ""),

		KafkaBrokers:       getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "monitoring-audit-group"),

		OpenSearchURL:      getEnv("OPENSEARCH_URL", "http://localhost:9200"),
		OpenSearchUsername: getEnv("OPENSEARCH_USERNAME", "admin"),
		OpenSearchPassword: getEnv("OPENSEARCH_PASSWORD", ""),

		JWTSecret:       getEnv("JWT_SECRET", ""),
		AccessTokenTTL:  time.Duration(getEnvAsInt("ACCESS_TOKEN_TTL_MINUTES", 15)) * time.Minute,
		RefreshTokenTTL: time.Duration(getEnvAsInt("REFRESH_TOKEN_TTL_DAYS", 7)) * 24 * time.Hour,

		CookieSecure:   getEnv("COOKIE_SECURE", "true") == "true",
		CookieDomain:   getEnv("COOKIE_DOMAIN", ""),
		CookieSameSite: getEnv("COOKIE_SAMESITE", "Lax"),

		StorageDriver:    getEnv("STORAGE_DRIVER", "local"),
		StorageLocalPath: getEnv("STORAGE_LOCAL_PATH", "./uploads"),
		MaxUploadSizeMB:  maxUpload,

		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),

		LogLevel:    getEnv("LOG_LEVEL", "debug"),
		LogOutput:   getEnv("LOG_OUTPUT", "console"),
		LogFilePath: getEnv("LOG_FILE_PATH", "./logs/app.log"),
	}
}

// ValidateServer rejects missing or known-development credentials before the
// API starts. Keeping this separate from Load allows maintenance utilities to
// load only the credentials they actually need.
func (c *Config) ValidateServer() error {
	required := map[string]string{
		"DB_PASSWORD":            c.DBPassword,
		"MINIO_ACCESS_KEY":       c.MinioAccessKey,
		"MINIO_SECRET_KEY":       c.MinioSecretKey,
		"SETTING_ENCRYPTION_KEY": c.SettingEncryptionKey,
		"JWT_SECRET":             c.JWTSecret,
	}
	for name, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}

	weak := map[string]bool{
		"secret": true, "minioadmin": true, "admin": true,
	}
	for name, value := range required {
		if weak[strings.ToLower(strings.TrimSpace(value))] {
			return fmt.Errorf("%s uses a known insecure default", name)
		}
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	key, err := base64.StdEncoding.DecodeString(c.SettingEncryptionKey)
	if err != nil || len(key) != 32 {
		return fmt.Errorf("SETTING_ENCRYPTION_KEY must be a base64-encoded 32-byte key")
	}
	return nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			return val
		}
	}
	return defaultVal
}
