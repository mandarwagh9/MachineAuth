package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                int
	Env                 string
	DatabaseURL         string
	JWTSigningAlgorithm string
	JWTKeyID            string
	JWTExpirySeconds    int
	AllowedOrigins      string
	RequireHTTPS        bool
	AdminEmail          string
	AdminPassword       string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:                getEnvInt("PORT", 8080),
		Env:                 getEnv("ENV", "development"),
		DatabaseURL:         getEnv("DATABASE_URL", "json:agentauth.json"),
		JWTSigningAlgorithm: getEnv("JWT_SIGNING_ALGORITHM", "RS256"),
		JWTKeyID:            getEnv("JWT_KEY_ID", "key-1"),
		JWTExpirySeconds:    getEnvInt("JWT_ACCESS_TOKEN_EXPIRY", 3600),
		AllowedOrigins:      getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
		RequireHTTPS:        getEnvBool("REQUIRE_HTTPS", false),
		AdminEmail:          getEnv("ADMIN_EMAIL", "admin@example.com"),
		AdminPassword:       getEnv("ADMIN_PASSWORD", "changeme"),
	}

	if cfg.Env == "development" {
		cfg.JWTExpirySeconds = 86400
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}

func (c *Config) GetTokenExpiry() time.Duration {
	return time.Duration(c.JWTExpirySeconds) * time.Second
}
