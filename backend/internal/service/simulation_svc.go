package service

import (
	"context"
	"fmt"
	"time"

	"github.com/faultline/faultline/internal/domain"
	"github.com/faultline/faultline/internal/repository"
	"github.com/faultline/faultline/internal/simulator"
	"github.com/google/uuid"
)

type SimulationService interface {
	GetScenarios() []simulator.Scenario
	StartSimulation(ctx context.Context, scenarioName string) (*domain.Simulation, error)
	StopSimulation(ctx context.Context, id uuid.UUID) (*domain.Simulation, error)
	GetAll(ctx context.Context, limit int) ([]domain.Simulation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Simulation, error)
	DegradeService(ctx context.Context, targetService string, modifiers map[string]float64, duration time.Duration) error
}

type simulationService struct {
	simEngine *simulator.Engine
	simRepo   repository.SimulationRepository
}

func NewSimulationService(simEngine *simulator.Engine, simRepo repository.SimulationRepository) SimulationService {
	return &simulationService{
		simEngine: simEngine,
		simRepo:   simRepo,
	}
}

func (s *simulationService) GetScenarios() []simulator.Scenario {
	return s.simEngine.GetScenarios()
}

func (s *simulationService) StartSimulation(ctx context.Context, scenarioName string) (*domain.Simulation, error) {
	if scenarioName == "" {
		return nil, fmt.Errorf("%w: scenario name is required", domain.ErrInvalidInput)
	}
	return s.simEngine.StartSimulation(ctx, scenarioName)
}

func (s *simulationService) StopSimulation(ctx context.Context, id uuid.UUID) (*domain.Simulation, error) {
	return s.simEngine.StopSimulation(ctx, id)
}

func (s *simulationService) GetAll(ctx context.Context, limit int) ([]domain.Simulation, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	sims, err := s.simRepo.GetAll(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get simulations: %w", err)
	}
	if sims == nil {
		return make([]domain.Simulation, 0), nil
	}
	return sims, nil
}

func (s *simulationService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Simulation, error) {
	sim, err := s.simRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get simulation: %w", err)
	}
	if sim == nil {
		return nil, domain.ErrNotFound
	}
	return sim, nil
}

func (s *simulationService) DegradeService(ctx context.Context, targetService string, modifiers map[string]float64, duration time.Duration) error {
	if targetService == "" {
		return fmt.Errorf("%w: target service is required", domain.ErrInvalidInput)
	}
	if duration <= 0 {
		duration = 60 * time.Second
	}
	return s.simEngine.DegradeService(ctx, targetService, modifiers, duration)
}
