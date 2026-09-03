package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/faultline/faultline/internal/domain"
	"github.com/faultline/faultline/internal/service"
	"github.com/faultline/faultline/internal/simulator"
)

// Collector generates and persists synthetic metrics for all services.
type Collector struct {
	serviceSvc service.ServiceService
	metricSvc  service.MetricService
	simEngine  *simulator.Engine
	profiles   map[string]simulator.ServiceProfile
}

func NewCollector(
	serviceSvc service.ServiceService,
	metricSvc service.MetricService,
	simEngine *simulator.Engine,
) *Collector {
	profiles := make(map[string]simulator.ServiceProfile)
	for _, p := range simulator.DefaultProfiles() {
		profiles[p.Name] = p
	}
	return &Collector{
		serviceSvc: serviceSvc,
		metricSvc:  metricSvc,
		simEngine:  simEngine,
		profiles:   profiles,
	}
}

// Run starts the metric collection loop. It collects every 5 seconds.
func (c *Collector) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	slog.Info("metrics collector worker started", "interval", "5s")

	// Collect immediately on start
	c.collect(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("metrics collector worker stopped")
			return
		case <-ticker.C:
			c.collect(ctx)
		}
	}
}

func (c *Collector) collect(ctx context.Context) {
	services, err := c.serviceSvc.GetAll(ctx)
	if err != nil {
		slog.Error("failed to get services for collection", "error", err)
		return
	}

	for _, svc := range services {
		profile, ok := c.profiles[svc.Name]
		if !ok {
			continue
		}

		var modifiers map[string]float64
		if c.simEngine != nil {
			modifiers, _ = c.simEngine.GetModifiers(ctx, svc.Name)
		}

		// Generate synthetic metric with active modifiers
		metric := simulator.GenerateMetric(profile, modifiers)
		metric.ServiceID = svc.ID

		// Persist via MetricService (stores in Postgres and caches in Redis)
		if err := c.metricSvc.RecordMetric(ctx, &metric); err != nil {
			slog.Error("failed to record metric", "service", svc.Name, "error", err)
			continue
		}
	}
}

// SeedServices ensures all 5 simulated services and their dependencies exist in the database.
func SeedServices(ctx context.Context, serviceSvc service.ServiceService, depSvc service.DependencyService) error {
	profiles := simulator.DefaultProfiles()
	serviceMap := make(map[string]*domain.Service)

	for _, p := range profiles {
		// Check if service already exists
		existing, err := serviceSvc.GetByName(ctx, p.Name)
		if err == nil && existing != nil {
			serviceMap[p.Name] = existing
			continue
		}

		svc := &domain.Service{
			ID:          domain.NewUUID(),
			Name:        p.Name,
			DisplayName: p.DisplayName,
			Description: p.Description,
			Status:      domain.StatusHealthy,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := serviceSvc.Create(ctx, svc); err != nil {
			return err
		}
		serviceMap[p.Name] = svc
		slog.Info("seeded service", "name", p.Name, "display_name", p.DisplayName)
	}

	// Seed dependencies
	for _, dep := range simulator.DefaultDependencies() {
		source, ok := serviceMap[dep.Source]
		if !ok {
			continue
		}
		target, ok := serviceMap[dep.Target]
		if !ok {
			continue
		}

		d := &domain.Dependency{
			ID:             domain.NewUUID(),
			SourceID:       source.ID,
			TargetID:       target.ID,
			DependencyType: "sync",
			CreatedAt:      time.Now(),
		}
		if err := depSvc.Create(ctx, d); err != nil {
			// Duplicate or existing dependency is fine
			continue
		}
	}

	slog.Info("service seeding complete", "services", len(profiles))
	return nil
}
