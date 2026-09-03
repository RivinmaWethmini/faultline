package postgres

import (
	"context"
	"fmt"

	"github.com/faultline/faultline/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DependencyRepo struct {
	pool *pgxpool.Pool
}

func NewDependencyRepo(pool *pgxpool.Pool) *DependencyRepo {
	return &DependencyRepo{pool: pool}
}

func (r *DependencyRepo) GetAll(ctx context.Context) ([]domain.Dependency, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, source_id, target_id, dependency_type, created_at
		 FROM dependencies ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("querying dependencies: %w", err)
	}
	defer rows.Close()

	var deps []domain.Dependency
	for rows.Next() {
		var d domain.Dependency
		if err := rows.Scan(&d.ID, &d.SourceID, &d.TargetID, &d.DependencyType, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning dependency: %w", err)
		}
		deps = append(deps, d)
	}
	return deps, rows.Err()
}

func (r *DependencyRepo) Create(ctx context.Context, dep *domain.Dependency) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO dependencies (id, source_id, target_id, dependency_type, created_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (source_id, target_id) DO NOTHING`,
		dep.ID, dep.SourceID, dep.TargetID, dep.DependencyType, dep.CreatedAt)
	if err != nil {
		return fmt.Errorf("inserting dependency: %w", err)
	}
	return nil
}

func (r *DependencyRepo) GetBySource(ctx context.Context, sourceID uuid.UUID) ([]domain.Dependency, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, source_id, target_id, dependency_type, created_at
		 FROM dependencies WHERE source_id = $1`, sourceID)
	if err != nil {
		return nil, fmt.Errorf("querying dependencies by source: %w", err)
	}
	defer rows.Close()

	var deps []domain.Dependency
	for rows.Next() {
		var d domain.Dependency
		if err := rows.Scan(&d.ID, &d.SourceID, &d.TargetID, &d.DependencyType, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning dependency: %w", err)
		}
		deps = append(deps, d)
	}
	return deps, rows.Err()
}

func (r *DependencyRepo) GetByTarget(ctx context.Context, targetID uuid.UUID) ([]domain.Dependency, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, source_id, target_id, dependency_type, created_at
		 FROM dependencies WHERE target_id = $1`, targetID)
	if err != nil {
		return nil, fmt.Errorf("querying dependencies by target: %w", err)
	}
	defer rows.Close()

	var deps []domain.Dependency
	for rows.Next() {
		var d domain.Dependency
		if err := rows.Scan(&d.ID, &d.SourceID, &d.TargetID, &d.DependencyType, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning dependency: %w", err)
		}
		deps = append(deps, d)
	}
	return deps, rows.Err()
}
