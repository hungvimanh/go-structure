package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	I18nDir string

	JWTPublicKey  string
	JWTPrivateKey string
}

func Load() (*Config, error) {
	// Load .env nếu file tồn tại.
	// Khi deploy production, có thể truyền environment variables
	// trực tiếp mà không cần file .env.
	_ = godotenv.Load()

	config := &Config{
		AppPort:    getEnv("APP_PORT", "5000"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     getEnv("DB_NAME", "postgres"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		I18nDir: getEnv("I18N_DIR", "internal"),

		JWTPublicKey:  getEnv("JWT_PUBLIC_KEY", ""),
		JWTPrivateKey: getEnv("JWT_PRIVATE_KEY", ""),
	}

	if config.DBPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}

	return config, nil
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	return value
}
