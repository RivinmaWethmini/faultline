package postgres

import (
	"context"
	"fmt"

	"github.com/faultline/faultline/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ServiceRepo struct {
	pool *pgxpool.Pool
}

func NewServiceRepo(pool *pgxpool.Pool) *ServiceRepo {
	return &ServiceRepo{pool: pool}
}

func (r *ServiceRepo) GetAll(ctx context.Context) ([]domain.Service, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, display_name, description, status, created_at, updated_at
		 FROM services ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("querying services: %w", err)
	}
	defer rows.Close()

	var services []domain.Service
	for rows.Next() {
		var s domain.Service
		if err := rows.Scan(&s.ID, &s.Name, &s.DisplayName, &s.Description,
			&s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning service: %w", err)
		}
		services = append(services, s)
	}
	return services, rows.Err()
}

func (r *ServiceRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Service, error) {
	var s domain.Service
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, display_name, description, status, created_at, updated_at
		 FROM services WHERE id = $1`, id).
		Scan(&s.ID, &s.Name, &s.DisplayName, &s.Description,
			&s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying service by id: %w", err)
	}
	return &s, nil
}

func (r *ServiceRepo) GetByName(ctx context.Context, name string) (*domain.Service, error) {
	var s domain.Service
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, display_name, description, status, created_at, updated_at
		 FROM services WHERE name = $1`, name).
		Scan(&s.ID, &s.Name, &s.DisplayName, &s.Description,
			&s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying service by name: %w", err)
	}
	return &s, nil
}

func (r *ServiceRepo) Create(ctx context.Context, svc *domain.Service) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO services (id, name, display_name, description, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (name) DO NOTHING`,
		svc.ID, svc.Name, svc.DisplayName, svc.Description, svc.Status, svc.CreatedAt, svc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting service: %w", err)
	}
	return nil
}

func (r *ServiceRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ServiceStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE services SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, id)
	if err != nil {
		return fmt.Errorf("updating service status: %w", err)
	}
	return nil
}
