package service

import (
	"context"
	"fmt"

	"github.com/faultline/faultline/internal/domain"
	"github.com/faultline/faultline/internal/repository"
	"github.com/google/uuid"
)

type ServiceService interface {
	GetAll(ctx context.Context) ([]domain.Service, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Service, error)
	GetByName(ctx context.Context, name string) (*domain.Service, error)
	Create(ctx context.Context, svc *domain.Service) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ServiceStatus) error
}

type serviceService struct {
	repo repository.ServiceRepository
}

func NewServiceService(repo repository.ServiceRepository) ServiceService {
	return &serviceService{repo: repo}
}

func (s *serviceService) GetAll(ctx context.Context) ([]domain.Service, error) {
	services, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("service service getAll: %w", err)
	}
	if services == nil {
		return []domain.Service{}, nil
	}
	return services, nil
}

func (s *serviceService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Service, error) {
	svc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service service getByID: %w", err)
	}
	if svc == nil {
		return nil, domain.ErrNotFound
	}
	return svc, nil
}

func (s *serviceService) GetByName(ctx context.Context, name string) (*domain.Service, error) {
	svc, err := s.repo.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("service service getByName: %w", err)
	}
	if svc == nil {
		return nil, domain.ErrNotFound
	}
	return svc, nil
}

func (s *serviceService) Create(ctx context.Context, svc *domain.Service) error {
	if svc.Name == "" {
		return fmt.Errorf("%w: service name is required", domain.ErrInvalidInput)
	}
	if svc.DisplayName == "" {
		svc.DisplayName = svc.Name
	}
	return s.repo.Create(ctx, svc)
}

func (s *serviceService) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ServiceStatus) error {
	return s.repo.UpdateStatus(ctx, id, status)
}
