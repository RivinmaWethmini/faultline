package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	CORS     CORSConfig
	LogLevel string
}

type ServerConfig struct {
	Port int
}

type PostgresConfig struct {
	URL      string
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

func (c PostgresConfig) DSN() string {
	if c.URL != "" {
		return c.URL
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.Database)
}

type RedisConfig struct {
	URL      string
	Host     string
	Port     int
	Optional bool
}

func (c RedisConfig) Addr() string {
	if c.URL != "" {
		return c.URL
	}
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type CORSConfig struct {
	AllowedOrigins []string
}

func Load() *Config {
	// Support standard cloud PORT, fallback to SERVER_PORT, default to 8080
	port := getEnvInt("PORT", getEnvInt("SERVER_PORT", 8080))

	corsRaw := getEnv("CORS_ALLOWED_ORIGINS", "*")
	var allowedOrigins []string
	for _, origin := range strings.Split(corsRaw, ",") {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			allowedOrigins = append(allowedOrigins, trimmed)
		}
	}
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}

	return &Config{
		Server: ServerConfig{
			Port: port,
		},
		Postgres: PostgresConfig{
			URL:      getEnv("DATABASE_URL", ""),
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvInt("POSTGRES_PORT", 5432),
			User:     getEnv("POSTGRES_USER", "faultline"),
			Password: getEnv("POSTGRES_PASSWORD", "faultline_secret"),
			Database: getEnv("POSTGRES_DB", "faultline"),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", ""),
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Optional: getEnvBool("REDIS_OPTIONAL", false),
		},
		CORS: CORSConfig{
			AllowedOrigins: allowedOrigins,
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if val := os.Getenv(key); val != "" {
		lower := strings.ToLower(val)
		return lower == "true" || lower == "1" || lower == "yes"
	}
	return fallback
}
