package api

import (
	"net/http"
	"time"

	"github.com/faultline/faultline/internal/api/handlers"
	"github.com/faultline/faultline/internal/api/middleware"
	"github.com/faultline/faultline/internal/repository"
	"github.com/faultline/faultline/internal/repository/redis"
	"github.com/faultline/faultline/internal/service"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RouterDeps struct {
	ServiceSvc     service.ServiceService
	MetricSvc      service.MetricService
	DependencySvc  service.DependencyService
	IncidentSvc    service.IncidentService
	SimSvc         service.SimulationService
	RiskSvc        service.RiskService
	RiskRepo       repository.RiskRepository
	Pool           *pgxpool.Pool
	Cache          *redis.Cache
	AllowedOrigins []string
}

func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()

	// Global Middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.StructuredLogger)
	r.Use(middleware.Recoverer)
	r.Use(chiMiddleware.Timeout(30 * time.Second))
	r.Use(middleware.CORS(deps.AllowedOrigins...))

	// Initialize Handlers
	serviceH := handlers.NewServiceHandler(deps.ServiceSvc, deps.RiskRepo)
	metricH := handlers.NewMetricHandler(deps.MetricSvc)
	riskH := handlers.NewRiskHandler(deps.RiskSvc)
	depH := handlers.NewDependencyHandler(deps.DependencySvc)
	incidentH := handlers.NewIncidentHandler(deps.IncidentSvc)
	simH := handlers.NewSimulationHandler(deps.SimSvc)
	healthH := handlers.NewHealthHandler(deps.Pool, deps.Cache)

	// Direct root health endpoint for orchestrators/probes
	r.Get("/healthz", healthH.Check)

	// API Routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/system/health", healthH.Check)

		r.Route("/services", func(r chi.Router) {
			r.Get("/", serviceH.List)
			r.Get("/{id}", serviceH.Get)
			r.Get("/{id}/metrics", metricH.GetByService)
			r.Get("/{id}/risk", riskH.GetByService)
			r.Get("/{id}/dependency-impact", depH.Impact)
		})

		r.Get("/dependencies", depH.List)

		r.Route("/incidents", func(r chi.Router) {
			r.Get("/", incidentH.List)
			r.Get("/{id}", incidentH.Get)
		})

		r.Route("/simulations", func(r chi.Router) {
			r.Get("/", simH.List)
			r.Post("/", simH.Create)
			r.Post("/{id}/stop", simH.Stop)
			r.Post("/degrade", simH.Degrade)
			r.Get("/scenarios", simH.Scenarios)
		})
	})

	return r
}
