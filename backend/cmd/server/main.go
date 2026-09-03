package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/faultline/faultline/internal/api"
	"github.com/faultline/faultline/internal/config"
	"github.com/faultline/faultline/internal/repository/postgres"
	"github.com/faultline/faultline/internal/repository/redis"
	"github.com/faultline/faultline/internal/service"
	"github.com/faultline/faultline/internal/simulator"
	"github.com/faultline/faultline/internal/worker"
)

func main() {
	cfg := config.Load()

	// Structured logging with slog
	var logLevel slog.Level
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	slog.Info("starting Faultline server",
		"port", cfg.Server.Port,
		"postgres_configured", cfg.Postgres.DSN() != "",
		"redis_configured", cfg.Redis.Addr() != "",
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to PostgreSQL
	pool, err := postgres.NewPool(ctx, cfg.Postgres.DSN())
	if err != nil {
		slog.Error("failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("connected to PostgreSQL successfully")

	// Run migrations
	migrationsDir := "migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		migrationsDir = "backend/migrations"
	}
	if err := postgres.RunMigrations(ctx, pool, migrationsDir); err != nil {
		slog.Error("failed to run database migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("database migrations completed successfully")

	// Connect to Redis
	var cache *redis.Cache
	cache, err = redis.NewCache(cfg.Redis.Addr())
	if err != nil {
		if cfg.Redis.Optional {
			slog.Warn("failed to connect to Redis, but REDIS_OPTIONAL=true; running without Redis cache", "error", err)
			cache = &redis.Cache{}
		} else {
			slog.Error("failed to connect to Redis", "error", err)
			os.Exit(1)
		}
	} else {
		defer cache.Close()
		slog.Info("connected to Redis successfully")
	}

	// Repository Layer
	serviceRepo := postgres.NewServiceRepo(pool)
	metricRepo := postgres.NewMetricRepo(pool)
	riskRepo := postgres.NewRiskRepo(pool)
	depRepo := postgres.NewDependencyRepo(pool)
	incidentRepo := postgres.NewIncidentRepo(pool)
	simRepo := postgres.NewSimulationRepo(pool)

	// Simulation Engine
	simEngine := simulator.NewEngine(simRepo, cache)

	// Service Layer (Clean Business Logic Separation)
	serviceSvc := service.NewServiceService(serviceRepo)
	metricSvc := service.NewMetricService(metricRepo, cache)
	depSvc := service.NewDependencyService(depRepo, serviceRepo, riskRepo)
	incidentSvc := service.NewIncidentService(incidentRepo, serviceRepo)
	simSvc := service.NewSimulationService(simEngine, simRepo)
	riskSvc := service.NewRiskService(riskRepo, metricRepo, cache)

	// Seed simulated microservices and dependency graph
	if err := worker.SeedServices(ctx, serviceSvc, depSvc); err != nil {
		slog.Error("failed to seed initial services", "error", err)
		os.Exit(1)
	}

	// Background Workers
	// 1. Metric Collector: generates synthetic metrics every 5s and persists to PG + Redis
	collector := worker.NewCollector(serviceSvc, metricSvc, simEngine)
	go collector.Run(ctx)

	// 2. Risk Evaluator: continuously computes 0-100 explainable risk scores from rolling baselines
	riskEvaluator := worker.NewRiskEvaluator(serviceRepo, metricRepo, riskRepo, cache)
	go riskEvaluator.Run(ctx)

	// 3. Incident Detector: monitors threshold crossings and creates/escalates/resolves incidents
	incidentDetector := worker.NewIncidentDetector(serviceRepo, riskRepo, depRepo, incidentRepo)
	go incidentDetector.Run(ctx)

	// Setup API Router
	router := api.NewRouter(api.RouterDeps{
		ServiceSvc:     serviceSvc,
		MetricSvc:      metricSvc,
		DependencySvc:  depSvc,
		IncidentSvc:    incidentSvc,
		SimSvc:         simSvc,
		RiskSvc:        riskSvc,
		RiskRepo:       riskRepo,
		Pool:           pool,
		Cache:          cache,
		AllowedOrigins: cfg.CORS.AllowedOrigins,
	})

	// Start HTTP Server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful Shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		slog.Info("shutdown signal received, commencing graceful shutdown")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("error during server shutdown", "error", err)
		}
	}()

	slog.Info("faultline backend is running and listening", "port", cfg.Server.Port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("server fatal error", "error", err)
		os.Exit(1)
	}

	slog.Info("faultline server stopped cleanly")
}
