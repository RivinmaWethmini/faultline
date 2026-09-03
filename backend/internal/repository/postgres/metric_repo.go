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

type MetricRepo struct {
	pool *pgxpool.Pool
}

func NewMetricRepo(pool *pgxpool.Pool) *MetricRepo {
	return &MetricRepo{pool: pool}
}

func (r *MetricRepo) Create(ctx context.Context, m *domain.Metric) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO metrics (id, service_id, timestamp, response_latency_ms, request_rate,
		 error_rate, timeout_rate, cpu_usage, memory_usage, dep_latency_ms, dep_error_rate)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		m.ID, m.ServiceID, m.Timestamp, m.ResponseLatency, m.RequestRate,
		m.ErrorRate, m.TimeoutRate, m.CPUUsage, m.MemoryUsage, m.DepLatency, m.DepErrorRate)
	if err != nil {
		return fmt.Errorf("inserting metric: %w", err)
	}
	return nil
}

func (r *MetricRepo) GetByService(ctx context.Context, serviceID uuid.UUID, since time.Time, limit int) ([]domain.Metric, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, service_id, timestamp, response_latency_ms, request_rate,
		 error_rate, timeout_rate, cpu_usage, memory_usage, dep_latency_ms, dep_error_rate
		 FROM metrics WHERE service_id = $1 AND timestamp >= $2
		 ORDER BY timestamp DESC LIMIT $3`,
		serviceID, since, limit)
	if err != nil {
		return nil, fmt.Errorf("querying metrics: %w", err)
	}
	defer rows.Close()

	var metrics []domain.Metric
	for rows.Next() {
		var m domain.Metric
		if err := rows.Scan(&m.ID, &m.ServiceID, &m.Timestamp, &m.ResponseLatency,
			&m.RequestRate, &m.ErrorRate, &m.TimeoutRate, &m.CPUUsage, &m.MemoryUsage,
			&m.DepLatency, &m.DepErrorRate); err != nil {
			return nil, fmt.Errorf("scanning metric: %w", err)
		}
		metrics = append(metrics, m)
	}
	return metrics, rows.Err()
}

func (r *MetricRepo) GetLatestByService(ctx context.Context, serviceID uuid.UUID) (*domain.Metric, error) {
	var m domain.Metric
	err := r.pool.QueryRow(ctx,
		`SELECT id, service_id, timestamp, response_latency_ms, request_rate,
		 error_rate, timeout_rate, cpu_usage, memory_usage, dep_latency_ms, dep_error_rate
		 FROM metrics WHERE service_id = $1 ORDER BY timestamp DESC LIMIT 1`,
		serviceID).
		Scan(&m.ID, &m.ServiceID, &m.Timestamp, &m.ResponseLatency,
			&m.RequestRate, &m.ErrorRate, &m.TimeoutRate, &m.CPUUsage, &m.MemoryUsage,
			&m.DepLatency, &m.DepErrorRate)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying latest metric: %w", err)
	}
	return &m, nil
}
