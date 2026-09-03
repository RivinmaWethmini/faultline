package postgres

import (
	"context"
	"fmt"

	"github.com/faultline/faultline/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SimulationRepo struct {
	pool *pgxpool.Pool
}

func NewSimulationRepo(pool *pgxpool.Pool) *SimulationRepo {
	return &SimulationRepo{pool: pool}
}

func (r *SimulationRepo) Create(ctx context.Context, sim *domain.Simulation) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO simulations (id, scenario, target_service, parameters, status, started_at, completed_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		sim.ID, sim.Scenario, sim.TargetService, sim.Parameters, sim.Status,
		sim.StartedAt, sim.CompletedAt, sim.CreatedAt)
	if err != nil {
		return fmt.Errorf("inserting simulation: %w", err)
	}
	return nil
}

func (r *SimulationRepo) Update(ctx context.Context, sim *domain.Simulation) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE simulations SET status = $1, started_at = $2, completed_at = $3 WHERE id = $4`,
		sim.Status, sim.StartedAt, sim.CompletedAt, sim.ID)
	if err != nil {
		return fmt.Errorf("updating simulation: %w", err)
	}
	return nil
}

func (r *SimulationRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Simulation, error) {
	var sim domain.Simulation
	err := r.pool.QueryRow(ctx,
		`SELECT id, scenario, target_service, parameters, status, started_at, completed_at, created_at
		 FROM simulations WHERE id = $1`, id).
		Scan(&sim.ID, &sim.Scenario, &sim.TargetService, &sim.Parameters, &sim.Status,
			&sim.StartedAt, &sim.CompletedAt, &sim.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying simulation: %w", err)
	}
	return &sim, nil
}

func (r *SimulationRepo) GetAll(ctx context.Context, limit int) ([]domain.Simulation, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, scenario, target_service, parameters, status, started_at, completed_at, created_at
		 FROM simulations ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("querying simulations: %w", err)
	}
	defer rows.Close()

	var sims []domain.Simulation
	for rows.Next() {
		var sim domain.Simulation
		if err := rows.Scan(&sim.ID, &sim.Scenario, &sim.TargetService, &sim.Parameters, &sim.Status,
			&sim.StartedAt, &sim.CompletedAt, &sim.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning simulation: %w", err)
		}
		sims = append(sims, sim)
	}
	return sims, rows.Err()
}

func (r *SimulationRepo) GetRunning(ctx context.Context) ([]domain.Simulation, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, scenario, target_service, parameters, status, started_at, completed_at, created_at
		 FROM simulations WHERE status = 'running' ORDER BY started_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("querying running simulations: %w", err)
	}
	defer rows.Close()

	var sims []domain.Simulation
	for rows.Next() {
		var sim domain.Simulation
		if err := rows.Scan(&sim.ID, &sim.Scenario, &sim.TargetService, &sim.Parameters, &sim.Status,
			&sim.StartedAt, &sim.CompletedAt, &sim.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning simulation: %w", err)
		}
		sims = append(sims, sim)
	}
	return sims, rows.Err()
}
