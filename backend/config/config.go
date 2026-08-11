package config

import (
	"log"
	"os"
	"strconv"

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

	// Kafka Config
	KafkaBrokers       string
	KafkaConsumerGroup string

	// OpenSearch Config
	OpenSearchURL      string
	OpenSearchUsername string
	OpenSearchPassword string

	// JWT Config
	JWTSecret       string
	JWTExpiredHours int

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
		AppBaseURL: getEnv("APP_BASE_URL", "http://localhost:8080"),

		DBDriver:   getEnv("DB_DRIVER", "postgres"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5434"),
		DBName:     getEnv("DB_NAME", "monitoring_audit"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "secret"),
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
		MinioAccessKey:  getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey:  getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:     getEnv("MINIO_BUCKET", "monitoring-audit-bucket"),
		MinioUseSSL:     getEnv("MINIO_USE_SSL", "false") == "true",
		MinioAllowedIPs: getEnv("MINIO_ALLOWED_IPS", "127.0.0.1/32,10.0.0.0/8,192.168.0.0/16"),

		SettingEncryptionKey: getEnv("SETTING_ENCRYPTION_KEY", "sCb2UdNCSu3RBEYLF6IG/18C6VAVuYftUhFB1lzRoyw="),

		KafkaBrokers:       getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "monitoring-audit-group"),

		OpenSearchURL:      getEnv("OPENSEARCH_URL", "http://localhost:9200"),
		OpenSearchUsername: getEnv("OPENSEARCH_USERNAME", "admin"),
		OpenSearchPassword: getEnv("OPENSEARCH_PASSWORD", "admin"),

		JWTSecret:       getEnv("JWT_SECRET", "super-secret-key-12345"),
		JWTExpiredHours: getEnvAsInt("JWT_EXPIRED_HOURS", 24),

		StorageDriver:    getEnv("STORAGE_DRIVER", "local"),
		StorageLocalPath: getEnv("STORAGE_LOCAL_PATH", "./uploads"),
		MaxUploadSizeMB:  maxUpload,

		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),

		LogLevel:    getEnv("LOG_LEVEL", "debug"),
		LogOutput:   getEnv("LOG_OUTPUT", "console"),
		LogFilePath: getEnv("LOG_FILE_PATH", "./logs/app.log"),
	}
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
