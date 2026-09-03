package service

import (
	"context"
	"fmt"

	"github.com/faultline/faultline/internal/domain"
	"github.com/faultline/faultline/internal/engine/propagation"
	"github.com/faultline/faultline/internal/repository"
	"github.com/google/uuid"
)

type DependencyDTO struct {
	ID             uuid.UUID `json:"id"`
	SourceID       uuid.UUID `json:"source_id"`
	SourceName     string    `json:"source_name"`
	TargetID       uuid.UUID `json:"target_id"`
	TargetName     string    `json:"target_name"`
	DependencyType string    `json:"dependency_type"`
}

type DependencyService interface {
	GetAll(ctx context.Context) ([]DependencyDTO, error)
	Create(ctx context.Context, dep *domain.Dependency) error
	GetDependencyImpact(ctx context.Context, serviceID uuid.UUID) (*domain.DependencyImpactResult, error)
}

type dependencyService struct {
	depRepo     repository.DependencyRepository
	serviceRepo repository.ServiceRepository
	riskRepo    repository.RiskRepository
}

func NewDependencyService(depRepo repository.DependencyRepository, serviceRepo repository.ServiceRepository, riskRepo repository.RiskRepository) DependencyService {
	return &dependencyService{
		depRepo:     depRepo,
		serviceRepo: serviceRepo,
		riskRepo:    riskRepo,
	}
}

func (s *dependencyService) GetAll(ctx context.Context) ([]DependencyDTO, error) {
	deps, err := s.depRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("dependency service getAll: %w", err)
	}

	services, err := s.serviceRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("dependency service get services: %w", err)
	}

	nameMap := make(map[uuid.UUID]string)
	for _, svc := range services {
		nameMap[svc.ID] = svc.DisplayName
	}

	result := make([]DependencyDTO, len(deps))
	for i, d := range deps {
		result[i] = DependencyDTO{
			ID:             d.ID,
			SourceID:       d.SourceID,
			SourceName:     nameMap[d.SourceID],
			TargetID:       d.TargetID,
			TargetName:     nameMap[d.TargetID],
			DependencyType: d.DependencyType,
		}
	}

	return result, nil
}

func (s *dependencyService) Create(ctx context.Context, dep *domain.Dependency) error {
	if dep.SourceID == uuid.Nil || dep.TargetID == uuid.Nil {
		return fmt.Errorf("%w: source and target service IDs are required", domain.ErrInvalidInput)
	}
	if dep.SourceID == dep.TargetID {
		return fmt.Errorf("%w: circular self-dependency is invalid", domain.ErrInvalidInput)
	}
	return s.depRepo.Create(ctx, dep)
}

func (s *dependencyService) GetDependencyImpact(ctx context.Context, serviceID uuid.UUID) (*domain.DependencyImpactResult, error) {
	services, err := s.serviceRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch services for impact analysis: %w", err)
	}

	deps, err := s.depRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch dependencies for impact analysis: %w", err)
	}

	riskScores := make(map[uuid.UUID]int)
	if s.riskRepo != nil {
		latestScores, _ := s.riskRepo.GetAllLatest(ctx)
		for _, rs := range latestScores {
			riskScores[rs.ServiceID] = rs.OverallScore
		}
	}

	graph := propagation.NewGraph(services, deps)
	impact := graph.AnalyzeImpact(serviceID, riskScores)

	return &impact, nil
}
