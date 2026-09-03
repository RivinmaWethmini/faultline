package service

import (
	"context"
	"fmt"

	"github.com/faultline/faultline/internal/domain"
	"github.com/faultline/faultline/internal/repository"
	"github.com/google/uuid"
)

type IncidentDTO struct {
	domain.Incident
	ServiceName string                 `json:"service_name"`
	Events      []domain.IncidentEvent `json:"events,omitempty"`
}

type IncidentService interface {
	GetAll(ctx context.Context, status string, limit int) ([]IncidentDTO, error)
	GetByID(ctx context.Context, id uuid.UUID) (*IncidentDTO, error)
	Create(ctx context.Context, inc *domain.Incident) error
	Update(ctx context.Context, inc *domain.Incident) error
	AddEvent(ctx context.Context, evt *domain.IncidentEvent) error
}

type incidentService struct {
	incidentRepo repository.IncidentRepository
	serviceRepo  repository.ServiceRepository
}

func NewIncidentService(incidentRepo repository.IncidentRepository, serviceRepo repository.ServiceRepository) IncidentService {
	return &incidentService{
		incidentRepo: incidentRepo,
		serviceRepo:  serviceRepo,
	}
}

func (s *incidentService) GetAll(ctx context.Context, status string, limit int) ([]IncidentDTO, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	incidents, err := s.incidentRepo.GetAll(ctx, status, limit)
	if err != nil {
		return nil, fmt.Errorf("incident service getAll: %w", err)
	}

	services, _ := s.serviceRepo.GetAll(ctx)
	nameMap := make(map[uuid.UUID]string)
	for _, svc := range services {
		nameMap[svc.ID] = svc.DisplayName
	}

	result := make([]IncidentDTO, len(incidents))
	for i, inc := range incidents {
		result[i] = IncidentDTO{
			Incident:    inc,
			ServiceName: nameMap[inc.ServiceID],
		}
	}

	return result, nil
}

func (s *incidentService) GetByID(ctx context.Context, id uuid.UUID) (*IncidentDTO, error) {
	inc, err := s.incidentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("incident service getByID: %w", err)
	}
	if inc == nil {
		return nil, domain.ErrNotFound
	}

	events, _ := s.incidentRepo.GetEvents(ctx, inc.ID)
	if events == nil {
		events = make([]domain.IncidentEvent, 0)
	}

	serviceName := ""
	svc, _ := s.serviceRepo.GetByID(ctx, inc.ServiceID)
	if svc != nil {
		serviceName = svc.DisplayName
	}

	return &IncidentDTO{
		Incident:    *inc,
		ServiceName: serviceName,
		Events:      events,
	}, nil
}

func (s *incidentService) Create(ctx context.Context, inc *domain.Incident) error {
	return s.incidentRepo.Create(ctx, inc)
}

func (s *incidentService) Update(ctx context.Context, inc *domain.Incident) error {
	return s.incidentRepo.Update(ctx, inc)
}

func (s *incidentService) AddEvent(ctx context.Context, evt *domain.IncidentEvent) error {
	return s.incidentRepo.AddEvent(ctx, evt)
}
