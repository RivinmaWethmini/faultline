package postgres

import (
	"context"
	"fmt"

	"github.com/faultline/faultline/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IncidentRepo struct {
	pool *pgxpool.Pool
}

func NewIncidentRepo(pool *pgxpool.Pool) *IncidentRepo {
	return &IncidentRepo{pool: pool}
}

func (r *IncidentRepo) Create(ctx context.Context, inc *domain.Incident) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO incidents (id, service_id, title, severity, status, risk_score,
		 root_cause, propagation_path, impacted_services, anomalies, started_at, resolved_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		inc.ID, inc.ServiceID, inc.Title, inc.Severity, inc.Status, inc.RiskScore,
		inc.RootCause, inc.PropagationPath, inc.ImpactedServices, inc.Anomalies,
		inc.StartedAt, inc.ResolvedAt, inc.CreatedAt)
	if err != nil {
		return fmt.Errorf("inserting incident: %w", err)
	}
	return nil
}

func (r *IncidentRepo) Update(ctx context.Context, inc *domain.Incident) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE incidents SET severity = $1, status = $2, risk_score = $3,
		 root_cause = $4, propagation_path = $5, impacted_services = $6,
		 anomalies = $7, resolved_at = $8
		 WHERE id = $9`,
		inc.Severity, inc.Status, inc.RiskScore,
		inc.RootCause, inc.PropagationPath, inc.ImpactedServices,
		inc.Anomalies, inc.ResolvedAt, inc.ID)
	if err != nil {
		return fmt.Errorf("updating incident: %w", err)
	}
	return nil
}

func (r *IncidentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Incident, error) {
	var inc domain.Incident
	err := r.pool.QueryRow(ctx,
		`SELECT id, service_id, title, severity, status, risk_score, root_cause,
		 propagation_path, impacted_services, anomalies, started_at, resolved_at, created_at
		 FROM incidents WHERE id = $1`, id).
		Scan(&inc.ID, &inc.ServiceID, &inc.Title, &inc.Severity, &inc.Status,
			&inc.RiskScore, &inc.RootCause, &inc.PropagationPath, &inc.ImpactedServices,
			&inc.Anomalies, &inc.StartedAt, &inc.ResolvedAt, &inc.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying incident: %w", err)
	}
	return &inc, nil
}

func (r *IncidentRepo) GetAll(ctx context.Context, status string, limit int) ([]domain.Incident, error) {
	var query string
	var args []interface{}

	if status != "" {
		query = `SELECT id, service_id, title, severity, status, risk_score, root_cause,
		         propagation_path, impacted_services, anomalies, started_at, resolved_at, created_at
		         FROM incidents WHERE status = $1 ORDER BY started_at DESC LIMIT $2`
		args = []interface{}{status, limit}
	} else {
		query = `SELECT id, service_id, title, severity, status, risk_score, root_cause,
		         propagation_path, impacted_services, anomalies, started_at, resolved_at, created_at
		         FROM incidents ORDER BY started_at DESC LIMIT $1`
		args = []interface{}{limit}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying incidents: %w", err)
	}
	defer rows.Close()

	var incidents []domain.Incident
	for rows.Next() {
		var inc domain.Incident
		if err := rows.Scan(&inc.ID, &inc.ServiceID, &inc.Title, &inc.Severity, &inc.Status,
			&inc.RiskScore, &inc.RootCause, &inc.PropagationPath, &inc.ImpactedServices,
			&inc.Anomalies, &inc.StartedAt, &inc.ResolvedAt, &inc.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning incident: %w", err)
		}
		incidents = append(incidents, inc)
	}
	return incidents, rows.Err()
}

func (r *IncidentRepo) GetActiveByService(ctx context.Context, serviceID uuid.UUID) (*domain.Incident, error) {
	var inc domain.Incident
	err := r.pool.QueryRow(ctx,
		`SELECT id, service_id, title, severity, status, risk_score, root_cause,
		 propagation_path, impacted_services, anomalies, started_at, resolved_at, created_at
		 FROM incidents WHERE service_id = $1 AND status != 'resolved'
		 ORDER BY started_at DESC LIMIT 1`, serviceID).
		Scan(&inc.ID, &inc.ServiceID, &inc.Title, &inc.Severity, &inc.Status,
			&inc.RiskScore, &inc.RootCause, &inc.PropagationPath, &inc.ImpactedServices,
			&inc.Anomalies, &inc.StartedAt, &inc.ResolvedAt, &inc.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying active incident: %w", err)
	}
	return &inc, nil
}

func (r *IncidentRepo) AddEvent(ctx context.Context, evt *domain.IncidentEvent) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO incident_events (id, incident_id, event_type, message, metadata, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		evt.ID, evt.IncidentID, evt.EventType, evt.Message, evt.Metadata, evt.CreatedAt)
	if err != nil {
		return fmt.Errorf("inserting incident event: %w", err)
	}
	return nil
}

func (r *IncidentRepo) GetEvents(ctx context.Context, incidentID uuid.UUID) ([]domain.IncidentEvent, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, incident_id, event_type, message, metadata, created_at
		 FROM incident_events WHERE incident_id = $1 ORDER BY created_at`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("querying incident events: %w", err)
	}
	defer rows.Close()

	var events []domain.IncidentEvent
	for rows.Next() {
		var evt domain.IncidentEvent
		if err := rows.Scan(&evt.ID, &evt.IncidentID, &evt.EventType, &evt.Message,
			&evt.Metadata, &evt.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning incident event: %w", err)
		}
		events = append(events, evt)
	}
	return events, rows.Err()
}
