package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/faultline/faultline/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RiskRepo struct {
	pool *pgxpool.Pool
}

func NewRiskRepo(pool *pgxpool.Pool) *RiskRepo {
	return &RiskRepo{pool: pool}
}

func (r *RiskRepo) Create(ctx context.Context, rs *domain.RiskScore) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO risk_scores (id, service_id, timestamp, overall_score, risk_level,
		 latency_anomaly, error_anomaly, timeout_anomaly, traffic_anomaly, dependency_anomaly)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		rs.ID, rs.ServiceID, rs.Timestamp, rs.OverallScore, rs.Level,
		rs.LatencyAnomaly, rs.ErrorAnomaly, rs.TimeoutAnomaly, rs.TrafficAnomaly, rs.DependencyAnomaly)
	if err != nil {
		return fmt.Errorf("inserting risk score: %w", err)
	}
	return nil
}

func (r *RiskRepo) GetByService(ctx context.Context, serviceID uuid.UUID, since time.Time, limit int) ([]domain.RiskScore, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, service_id, timestamp, overall_score, risk_level,
		 latency_anomaly, error_anomaly, timeout_anomaly, traffic_anomaly, dependency_anomaly
		 FROM risk_scores WHERE service_id = $1 AND timestamp >= $2
		 ORDER BY timestamp DESC LIMIT $3`,
		serviceID, since, limit)
	if err != nil {
		return nil, fmt.Errorf("querying risk scores: %w", err)
	}
	defer rows.Close()

	var scores []domain.RiskScore
	for rows.Next() {
		var rs domain.RiskScore
		if err := rows.Scan(&rs.ID, &rs.ServiceID, &rs.Timestamp, &rs.OverallScore,
			&rs.Level, &rs.LatencyAnomaly, &rs.ErrorAnomaly, &rs.TimeoutAnomaly,
			&rs.TrafficAnomaly, &rs.DependencyAnomaly); err != nil {
			return nil, fmt.Errorf("scanning risk score: %w", err)
		}
		scores = append(scores, rs)
	}
	return scores, rows.Err()
}

func (r *RiskRepo) GetLatestByService(ctx context.Context, serviceID uuid.UUID) (*domain.RiskScore, error) {
	var rs domain.RiskScore
	err := r.pool.QueryRow(ctx,
		`SELECT id, service_id, timestamp, overall_score, risk_level,
		 latency_anomaly, error_anomaly, timeout_anomaly, traffic_anomaly, dependency_anomaly
		 FROM risk_scores WHERE service_id = $1 ORDER BY timestamp DESC LIMIT 1`,
		serviceID).
		Scan(&rs.ID, &rs.ServiceID, &rs.Timestamp, &rs.OverallScore,
			&rs.Level, &rs.LatencyAnomaly, &rs.ErrorAnomaly, &rs.TimeoutAnomaly,
			&rs.TrafficAnomaly, &rs.DependencyAnomaly)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying latest risk score: %w", err)
	}
	return &rs, nil
}

func (r *RiskRepo) GetAllLatest(ctx context.Context) ([]domain.RiskScore, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT ON (service_id)
		 id, service_id, timestamp, overall_score, risk_level,
		 latency_anomaly, error_anomaly, timeout_anomaly, traffic_anomaly, dependency_anomaly
		 FROM risk_scores ORDER BY service_id, timestamp DESC`)
	if err != nil {
		return nil, fmt.Errorf("querying all latest risk scores: %w", err)
	}
	defer rows.Close()

	var scores []domain.RiskScore
	for rows.Next() {
		var rs domain.RiskScore
		if err := rows.Scan(&rs.ID, &rs.ServiceID, &rs.Timestamp, &rs.OverallScore,
			&rs.Level, &rs.LatencyAnomaly, &rs.ErrorAnomaly, &rs.TimeoutAnomaly,
			&rs.TrafficAnomaly, &rs.DependencyAnomaly); err != nil {
			return nil, fmt.Errorf("scanning risk score: %w", err)
		}
		scores = append(scores, rs)
	}
	return scores, rows.Err()
}
