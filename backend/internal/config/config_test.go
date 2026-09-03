package config

import (
	"os"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear relevant env vars
	os.Unsetenv("PORT")
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("POSTGRES_HOST")
	os.Unsetenv("REDIS_URL")
	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("CORS_ALLOWED_ORIGINS")

	cfg := Load()

	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Postgres.Host != "localhost" {
		t.Errorf("expected default postgres host 'localhost', got %s", cfg.Postgres.Host)
	}
	if cfg.Postgres.Port != 5432 {
		t.Errorf("expected default postgres port 5432, got %d", cfg.Postgres.Port)
	}
	if cfg.Redis.Host != "localhost" {
		t.Errorf("expected default redis host 'localhost', got %s", cfg.Redis.Host)
	}
	if cfg.Postgres.DSN() != "postgres://faultline:faultline_secret@localhost:5432/faultline?sslmode=disable" {
		t.Errorf("unexpected default DSN: %s", cfg.Postgres.DSN())
	}
}

func TestLoadConfig_CustomEnv(t *testing.T) {
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("POSTGRES_HOST", "db.example.com")
	os.Setenv("POSTGRES_PORT", "5433")
	os.Setenv("LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("POSTGRES_HOST")
		os.Unsetenv("POSTGRES_PORT")
		os.Unsetenv("LOG_LEVEL")
	}()

	cfg := Load()

	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Postgres.Host != "db.example.com" {
		t.Errorf("expected postgres host 'db.example.com', got %s", cfg.Postgres.Host)
	}
	if cfg.Postgres.Port != 5433 {
		t.Errorf("expected postgres port 5433, got %d", cfg.Postgres.Port)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected log level 'debug', got %s", cfg.LogLevel)
	}
}

func TestLoadConfig_ProductionEnv(t *testing.T) {
	os.Setenv("PORT", "10000")
	os.Setenv("DATABASE_URL", "postgres://user:pass@render-pg.com:5432/prod_db?sslmode=require")
	os.Setenv("REDIS_URL", "rediss://default:token@render-redis.com:6379")
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://faultline.vercel.app, https://preview.vercel.app")
	os.Setenv("REDIS_OPTIONAL", "true")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("REDIS_URL")
		os.Unsetenv("CORS_ALLOWED_ORIGINS")
		os.Unsetenv("REDIS_OPTIONAL")
	}()

	cfg := Load()

	if cfg.Server.Port != 10000 {
		t.Errorf("expected port 10000 from PORT, got %d", cfg.Server.Port)
	}
	if cfg.Postgres.DSN() != "postgres://user:pass@render-pg.com:5432/prod_db?sslmode=require" {
		t.Errorf("expected DATABASE_URL to override DSN, got %s", cfg.Postgres.DSN())
	}
	if cfg.Redis.Addr() != "rediss://default:token@render-redis.com:6379" {
		t.Errorf("expected REDIS_URL to override Addr, got %s", cfg.Redis.Addr())
	}
	if len(cfg.CORS.AllowedOrigins) != 2 || cfg.CORS.AllowedOrigins[0] != "https://faultline.vercel.app" {
		t.Errorf("expected 2 CORS origins parsed, got %v", cfg.CORS.AllowedOrigins)
	}
	if !cfg.Redis.Optional {
		t.Errorf("expected Redis.Optional to be true")
	}
}
