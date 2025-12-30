package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	MinIO    MinIOConfig
	JWT      JWTConfig
	Server   ServerConfig
	Upload   UploadConfig
	Env      string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	PublicURL string
}

type JWTConfig struct {
	Secret     string
	Expiration string
}

type ServerConfig struct {
	Port           string
	AllowedOrigins []string
}

type UploadConfig struct {
	MaxSize int64
}

func Load() (*Config, error) {
	// Load .env file if it exists (ignore error in production)
	_ = godotenv.Load()

	env := getEnv("APP_ENV", "development")
	useSSL, _ := strconv.ParseBool(getEnv("MINIO_USE_SSL", "false"))
	maxSize, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_SIZE", "5242880"), 10, 64)

	config := &Config{
		Env: env,
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "food_sharing"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		MinIO: MinIOConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			Bucket:    getEnv("MINIO_BUCKET", "food-images"),
			UseSSL:    useSSL,
			PublicURL: getEnv("MINIO_PUBLIC_URL", "http://localhost:9000"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key"),
			Expiration: getEnv("JWT_EXPIRATION", "24h"),
		},
		Server: ServerConfig{
			Port:           getEnv("PORT", "8000"),
			AllowedOrigins: parseOrigins(getEnv("ALLOWED_ORIGINS", "http://localhost:3000")),
		},
		Upload: UploadConfig{
			MaxSize: maxSize,
		},
	}

	// Validate production configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate checks if required configuration is present
func (c *Config) Validate() error {
	// Production-specific validations
	if c.Env == "production" {
		if c.Database.Password == "" {
			return fmt.Errorf("DB_PASSWORD is required in production")
		}
		if c.JWT.Secret == "" || c.JWT.Secret == "your-secret-key" {
			return fmt.Errorf("JWT_SECRET must be set to a strong secret in production")
		}
		if c.Database.SSLMode == "disable" {
			return fmt.Errorf("DB_SSLMODE should not be 'disable' in production")
		}
		if !c.MinIO.UseSSL {
			// Warning but not fatal
			fmt.Println("WARNING: MinIO SSL is disabled in production")
		}
		if len(c.Server.AllowedOrigins) == 0 {
			return fmt.Errorf("ALLOWED_ORIGINS must be configured in production")
		}
	}

	// General validations
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.MinIO.Endpoint == "" {
		return fmt.Errorf("MINIO_ENDPOINT is required")
	}

	return nil
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseOrigins(origins string) []string {
	if origins == "" {
		return []string{}
	}

	var result []string
	parts := strings.Split(origins, ",")
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}
