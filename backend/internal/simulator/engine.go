package simulator

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/faultline/faultline/internal/domain"
	"github.com/faultline/faultline/internal/repository"
	"github.com/faultline/faultline/internal/repository/redis"
	"github.com/google/uuid"
)

// Engine manages simulation lifecycle and applies modifiers to metrics.
type Engine struct {
	simRepo   repository.SimulationRepository
	cache     *redis.Cache
	scenarios map[string]Scenario
}

func NewEngine(simRepo repository.SimulationRepository, cache *redis.Cache) *Engine {
	scenarios := make(map[string]Scenario)
	for _, s := range DefaultScenarios() {
		scenarios[s.Name] = s
	}
	return &Engine{
		simRepo:   simRepo,
		cache:     cache,
		scenarios: scenarios,
	}
}

// GetScenarios returns all available failure scenarios.
func (e *Engine) GetScenarios() []Scenario {
	return DefaultScenarios()
}

// StartSimulation triggers a predefined failure scenario.
func (e *Engine) StartSimulation(ctx context.Context, scenarioName string) (*domain.Simulation, error) {
	scenario, ok := e.scenarios[scenarioName]
	if !ok {
		return nil, fmt.Errorf("unknown scenario: %s", scenarioName)
	}

	now := time.Now()
	params, _ := json.Marshal(scenario.Modifiers)

	sim := &domain.Simulation{
		ID:            uuid.New(),
		Scenario:      scenarioName,
		TargetService: scenario.Target,
		Parameters:    params,
		Status:        domain.SimRunning,
		StartedAt:     &now,
		CreatedAt:     now,
	}

	if err := e.simRepo.Create(ctx, sim); err != nil {
		return nil, fmt.Errorf("creating simulation: %w", err)
	}

	// Store modifiers in Redis with TTL
	ttl := time.Duration(scenario.Duration) * time.Second
	if err := e.cache.SetSimulationState(ctx, scenario.Target, scenario.Modifiers, ttl); err != nil {
		slog.Error("failed to set simulation state in Redis", "error", err)
	}

	// Schedule completion in background
	go func() {
		time.Sleep(ttl)
		currentSim, err := e.simRepo.GetByID(context.Background(), sim.ID)
		if err == nil && currentSim != nil && currentSim.Status == domain.SimRunning {
			completedAt := time.Now()
			currentSim.Status = domain.SimCompleted
			currentSim.CompletedAt = &completedAt
			_ = e.simRepo.Update(context.Background(), currentSim)
			slog.Info("simulation completed", "scenario", scenarioName, "target", scenario.Target)
		}
	}()

	slog.Info("simulation started", "scenario", scenarioName, "target", scenario.Target, "duration", scenario.Duration)
	return sim, nil
}

// StopSimulation terminates a running simulation and clears its active degradation modifiers.
func (e *Engine) StopSimulation(ctx context.Context, id uuid.UUID) (*domain.Simulation, error) {
	sim, err := e.simRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("retrieving simulation: %w", err)
	}
	if sim == nil {
		return nil, domain.ErrNotFound
	}

	now := time.Now()
	sim.Status = domain.SimCompleted
	sim.CompletedAt = &now

	if err := e.simRepo.Update(ctx, sim); err != nil {
		return nil, fmt.Errorf("updating simulation status: %w", err)
	}

	// Clear active modifiers from Redis immediately to reverse degradation
	if err := e.cache.ClearSimulationState(ctx, sim.TargetService); err != nil {
		slog.Error("failed to clear simulation state from Redis", "error", err)
	}

	slog.Info("simulation stopped and reverted", "id", id, "target", sim.TargetService)
	return sim, nil
}

// DegradeService allows programmatic injection of degradation modifiers for any service.
func (e *Engine) DegradeService(ctx context.Context, targetService string, modifiers map[string]float64, duration time.Duration) error {
	now := time.Now()
	params, _ := json.Marshal(modifiers)

	sim := &domain.Simulation{
		ID:            uuid.New(),
		Scenario:      "programmatic_degradation",
		TargetService: targetService,
		Parameters:    params,
		Status:        domain.SimRunning,
		StartedAt:     &now,
		CreatedAt:     now,
	}

	if err := e.simRepo.Create(ctx, sim); err != nil {
		return fmt.Errorf("creating degradation simulation: %w", err)
	}

	if err := e.cache.SetSimulationState(ctx, targetService, modifiers, duration); err != nil {
		slog.Error("failed to set programmatic simulation in Redis", "error", err)
	}

	go func() {
		time.Sleep(duration)
		currentSim, err := e.simRepo.GetByID(context.Background(), sim.ID)
		if err == nil && currentSim != nil && currentSim.Status == domain.SimRunning {
			completedAt := time.Now()
			currentSim.Status = domain.SimCompleted
			currentSim.CompletedAt = &completedAt
			_ = e.simRepo.Update(context.Background(), currentSim)
			slog.Info("programmatic degradation ended", "target", targetService)
		}
	}()

	slog.Info("programmatic degradation applied", "target", targetService, "duration", duration)
	return nil
}

// GetModifiers retrieves active simulation modifiers for a service from Redis.
func (e *Engine) GetModifiers(ctx context.Context, serviceName string) (map[string]float64, error) {
	return e.cache.GetSimulationState(ctx, serviceName)
}
