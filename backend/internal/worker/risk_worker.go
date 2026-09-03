package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/faultline/faultline/internal/domain"
	"github.com/faultline/faultline/internal/engine/risk"
	"github.com/faultline/faultline/internal/repository"
	"github.com/faultline/faultline/internal/repository/redis"
	"github.com/google/uuid"
)

// RiskEvaluator periodically evaluates risk scores for all services.
type RiskEvaluator struct {
	serviceRepo repository.ServiceRepository
	metricRepo  repository.MetricRepository
	riskRepo    repository.RiskRepository
	cache       *redis.Cache
}

func NewRiskEvaluator(
	serviceRepo repository.ServiceRepository,
	metricRepo repository.MetricRepository,
	riskRepo repository.RiskRepository,
	cache *redis.Cache,
) *RiskEvaluator {
	return &RiskEvaluator{
		serviceRepo: serviceRepo,
		metricRepo:  metricRepo,
		riskRepo:    riskRepo,
		cache:       cache,
	}
}

// Run starts the risk evaluation loop. It evaluates every 10 seconds.
func (e *RiskEvaluator) Run(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	slog.Info("risk evaluator started", "interval", "10s")

	// Wait a bit for initial metrics to accumulate
	time.Sleep(15 * time.Second)

	e.evaluate(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("risk evaluator stopped")
			return
		case <-ticker.C:
			e.evaluate(ctx)
		}
	}
}

func (e *RiskEvaluator) evaluate(ctx context.Context) {
	services, err := e.serviceRepo.GetAll(ctx)
	if err != nil {
		slog.Error("failed to get services for risk evaluation", "error", err)
		return
	}

	for _, svc := range services {
		if err := e.evaluateService(ctx, svc); err != nil {
			slog.Error("failed to evaluate risk for service", "service", svc.Name, "error", err)
		}
	}
}

func (e *RiskEvaluator) evaluateService(ctx context.Context, svc domain.Service) error {
	// Get last 30 minutes of metrics for baseline
	since := time.Now().Add(-30 * time.Minute)
	metrics, err := e.metricRepo.GetByService(ctx, svc.ID, since, 360)
	if err != nil {
		return err
	}

	if len(metrics) < 3 {
		return nil // Not enough data
	}

	// Build history arrays
	history := risk.MetricHistory{}
	for _, m := range metrics {
		history.Latencies = append(history.Latencies, m.ResponseLatency)
		history.ErrorRates = append(history.ErrorRates, m.ErrorRate)
		history.TimeoutRates = append(history.TimeoutRates, m.TimeoutRate)
		history.RequestRates = append(history.RequestRates, m.RequestRate)
		history.DepLatencies = append(history.DepLatencies, m.DepLatency)
		history.DepErrorRates = append(history.DepErrorRates, m.DepErrorRate)
	}

	// Current metric is the most recent
	current := &metrics[0]

	// Calculate risk
	breakdown := risk.CalculateRiskScore(current, history)
	overall := risk.ComputeOverallScore(breakdown)
	level := domain.ClassifyRisk(overall)

	// Create risk score record
	rs := &domain.RiskScore{
		ID:                uuid.New(),
		ServiceID:         svc.ID,
		Timestamp:         time.Now(),
		OverallScore:      overall,
		Level:             level,
		LatencyAnomaly:    breakdown.LatencyAnomaly,
		ErrorAnomaly:      breakdown.ErrorAnomaly,
		TimeoutAnomaly:    breakdown.TimeoutAnomaly,
		TrafficAnomaly:    breakdown.TrafficAnomaly,
		DependencyAnomaly: breakdown.DependencyAnomaly,
	}

	if err := e.riskRepo.Create(ctx, rs); err != nil {
		return err
	}

	// Cache in Redis
	if err := e.cache.SetRiskScore(ctx, svc.ID, rs); err != nil {
		slog.Error("failed to cache risk score", "service", svc.Name, "error", err)
	}

	// Update service status based on risk level
	var status domain.ServiceStatus
	switch level {
	case domain.RiskCritical:
		status = domain.StatusCritical
	case domain.RiskHigh:
		status = domain.StatusUnhealthy
	case domain.RiskModerate:
		status = domain.StatusDegraded
	default:
		status = domain.StatusHealthy
	}

	if err := e.serviceRepo.UpdateStatus(ctx, svc.ID, status); err != nil {
		slog.Error("failed to update service status", "service", svc.Name, "error", err)
	}

	return nil
}
