package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/faultline/faultline/internal/domain"
	"github.com/faultline/faultline/internal/engine/incident"
	"github.com/faultline/faultline/internal/engine/propagation"
	"github.com/faultline/faultline/internal/repository"
	"github.com/google/uuid"
)

// IncidentDetector monitors risk scores and creates/resolves incidents.
type IncidentDetector struct {
	serviceRepo    repository.ServiceRepository
	riskRepo       repository.RiskRepository
	depRepo        repository.DependencyRepository
	incidentRepo   repository.IncidentRepository
	lowStreakCount map[uuid.UUID]int // tracks consecutive low-risk evaluations per service
}

func NewIncidentDetector(
	serviceRepo repository.ServiceRepository,
	riskRepo repository.RiskRepository,
	depRepo repository.DependencyRepository,
	incidentRepo repository.IncidentRepository,
) *IncidentDetector {
	return &IncidentDetector{
		serviceRepo:    serviceRepo,
		riskRepo:       riskRepo,
		depRepo:        depRepo,
		incidentRepo:   incidentRepo,
		lowStreakCount: make(map[uuid.UUID]int),
	}
}

// Run starts the incident detection loop. It checks every 15 seconds.
func (d *IncidentDetector) Run(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	slog.Info("incident detector started", "interval", "15s")

	// Wait for risk scores to accumulate
	time.Sleep(30 * time.Second)

	d.detect(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("incident detector stopped")
			return
		case <-ticker.C:
			d.detect(ctx)
		}
	}
}

func (d *IncidentDetector) detect(ctx context.Context) {
	services, err := d.serviceRepo.GetAll(ctx)
	if err != nil {
		slog.Error("failed to get services for incident detection", "error", err)
		return
	}

	// Build risk scores map for propagation analysis
	riskScores := make(map[uuid.UUID]int)
	for _, svc := range services {
		latest, err := d.riskRepo.GetLatestByService(ctx, svc.ID)
		if err != nil || latest == nil {
			continue
		}
		riskScores[svc.ID] = latest.OverallScore
	}

	// Build propagation graph
	deps, err := d.depRepo.GetAll(ctx)
	if err != nil {
		slog.Error("failed to get dependencies", "error", err)
		return
	}
	graph := propagation.NewGraph(services, deps)

	for _, svc := range services {
		latest, err := d.riskRepo.GetLatestByService(ctx, svc.ID)
		if err != nil || latest == nil {
			continue
		}

		activeIncident, err := d.incidentRepo.GetActiveByService(ctx, svc.ID)
		if err != nil {
			slog.Error("failed to check active incident", "service", svc.Name, "error", err)
			continue
		}

		if latest.OverallScore >= 60 {
			// HIGH or CRITICAL risk — create or escalate incident
			d.lowStreakCount[svc.ID] = 0

			if activeIncident == nil {
				d.createIncident(ctx, svc, latest, graph, riskScores)
			} else {
				d.escalateIfNeeded(ctx, activeIncident, latest)
			}
		} else if latest.OverallScore < 30 {
			// LOW risk — track streak for auto-resolution
			d.lowStreakCount[svc.ID]++
			if activeIncident != nil && d.lowStreakCount[svc.ID] >= 5 {
				d.resolveIncident(ctx, activeIncident, svc.Name)
				d.lowStreakCount[svc.ID] = 0
			}
		} else {
			// MODERATE — reset streak
			d.lowStreakCount[svc.ID] = 0
		}
	}
}

func (d *IncidentDetector) createIncident(
	ctx context.Context,
	svc domain.Service,
	rs *domain.RiskScore,
	graph *propagation.Graph,
	riskScores map[uuid.UUID]int,
) {
	breakdown := domain.AnomalyBreakdown{
		LatencyAnomaly:    rs.LatencyAnomaly,
		ErrorAnomaly:      rs.ErrorAnomaly,
		TimeoutAnomaly:    rs.TimeoutAnomaly,
		TrafficAnomaly:    rs.TrafficAnomaly,
		DependencyAnomaly: rs.DependencyAnomaly,
	}

	// Run propagation analysis
	prop := graph.Analyze(svc.ID, riskScores)

	// Generate explanation
	title := incident.GenerateTitle(svc.DisplayName, rs.Level, breakdown)
	rootCause := incident.GenerateExplanation(svc.DisplayName, breakdown, prop)

	anomaliesJSON, _ := json.Marshal(breakdown)
	pathJSON, _ := json.Marshal(prop.PropagationPath)
	impactedJSON, _ := json.Marshal(prop.AffectedServices)

	now := time.Now()
	inc := &domain.Incident{
		ID:               uuid.New(),
		ServiceID:        svc.ID,
		Title:            title,
		Severity:         rs.Level,
		Status:           domain.IncidentActive,
		RiskScore:        rs.OverallScore,
		RootCause:        rootCause,
		PropagationPath:  pathJSON,
		ImpactedServices: impactedJSON,
		Anomalies:        anomaliesJSON,
		StartedAt:        now,
		CreatedAt:        now,
	}

	if err := d.incidentRepo.Create(ctx, inc); err != nil {
		slog.Error("failed to create incident", "service", svc.Name, "error", err)
		return
	}

	// Add creation event
	evt := &domain.IncidentEvent{
		ID:         uuid.New(),
		IncidentID: inc.ID,
		EventType:  "created",
		Message:    title,
		CreatedAt:  now,
	}
	if err := d.incidentRepo.AddEvent(ctx, evt); err != nil {
		slog.Error("failed to add incident event", "error", err)
	}

	slog.Warn("incident created", "service", svc.Name, "severity", rs.Level, "score", rs.OverallScore)
}

func (d *IncidentDetector) escalateIfNeeded(ctx context.Context, inc *domain.Incident, rs *domain.RiskScore) {
	newLevel := rs.Level
	if newLevel == inc.Severity {
		return
	}

	// Only escalate, don't deescalate during active incident
	levelOrder := map[domain.RiskLevel]int{
		domain.RiskLow: 0, domain.RiskModerate: 1, domain.RiskHigh: 2, domain.RiskCritical: 3,
	}
	if levelOrder[newLevel] <= levelOrder[inc.Severity] {
		return
	}

	inc.Severity = newLevel
	inc.RiskScore = rs.OverallScore

	anomaliesJSON, _ := json.Marshal(domain.AnomalyBreakdown{
		LatencyAnomaly:    rs.LatencyAnomaly,
		ErrorAnomaly:      rs.ErrorAnomaly,
		TimeoutAnomaly:    rs.TimeoutAnomaly,
		TrafficAnomaly:    rs.TrafficAnomaly,
		DependencyAnomaly: rs.DependencyAnomaly,
	})
	inc.Anomalies = anomaliesJSON

	if err := d.incidentRepo.Update(ctx, inc); err != nil {
		slog.Error("failed to escalate incident", "error", err)
		return
	}

	now := time.Now()
	evt := &domain.IncidentEvent{
		ID:         uuid.New(),
		IncidentID: inc.ID,
		EventType:  "escalated",
		Message:    "Incident escalated to " + string(newLevel),
		CreatedAt:  now,
	}
	d.incidentRepo.AddEvent(ctx, evt)

	slog.Warn("incident escalated", "incident", inc.ID, "new_severity", newLevel)
}

func (d *IncidentDetector) resolveIncident(ctx context.Context, inc *domain.Incident, serviceName string) {
	now := time.Now()
	inc.Status = domain.IncidentResolved
	inc.ResolvedAt = &now

	if err := d.incidentRepo.Update(ctx, inc); err != nil {
		slog.Error("failed to resolve incident", "error", err)
		return
	}

	evt := &domain.IncidentEvent{
		ID:         uuid.New(),
		IncidentID: inc.ID,
		EventType:  "resolved",
		Message:    "Risk level returned to LOW for 5 consecutive evaluations. Incident auto-resolved.",
		CreatedAt:  now,
	}
	d.incidentRepo.AddEvent(ctx, evt)

	slog.Info("incident resolved", "service", serviceName, "incident", inc.ID)
}
