package repository

import (
	"context"
	"time"

	"github.com/faultline/faultline/internal/domain"
	"github.com/google/uuid"
)

type ServiceRepository interface {
	GetAll(ctx context.Context) ([]domain.Service, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Service, error)
	GetByName(ctx context.Context, name string) (*domain.Service, error)
	Create(ctx context.Context, svc *domain.Service) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ServiceStatus) error
}

type MetricRepository interface {
	Create(ctx context.Context, m *domain.Metric) error
	GetByService(ctx context.Context, serviceID uuid.UUID, since time.Time, limit int) ([]domain.Metric, error)
	GetLatestByService(ctx context.Context, serviceID uuid.UUID) (*domain.Metric, error)
}

type RiskRepository interface {
	Create(ctx context.Context, rs *domain.RiskScore) error
	GetByService(ctx context.Context, serviceID uuid.UUID, since time.Time, limit int) ([]domain.RiskScore, error)
	GetLatestByService(ctx context.Context, serviceID uuid.UUID) (*domain.RiskScore, error)
	GetAllLatest(ctx context.Context) ([]domain.RiskScore, error)
}

type DependencyRepository interface {
	GetAll(ctx context.Context) ([]domain.Dependency, error)
	Create(ctx context.Context, dep *domain.Dependency) error
	GetBySource(ctx context.Context, sourceID uuid.UUID) ([]domain.Dependency, error)
	GetByTarget(ctx context.Context, targetID uuid.UUID) ([]domain.Dependency, error)
}

type IncidentRepository interface {
	Create(ctx context.Context, inc *domain.Incident) error
	Update(ctx context.Context, inc *domain.Incident) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Incident, error)
	GetAll(ctx context.Context, status string, limit int) ([]domain.Incident, error)
	GetActiveByService(ctx context.Context, serviceID uuid.UUID) (*domain.Incident, error)
	AddEvent(ctx context.Context, evt *domain.IncidentEvent) error
	GetEvents(ctx context.Context, incidentID uuid.UUID) ([]domain.IncidentEvent, error)
}

type SimulationRepository interface {
	Create(ctx context.Context, sim *domain.Simulation) error
	Update(ctx context.Context, sim *domain.Simulation) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Simulation, error)
	GetAll(ctx context.Context, limit int) ([]domain.Simulation, error)
	GetRunning(ctx context.Context) ([]domain.Simulation, error)
}
