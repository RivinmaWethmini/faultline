package service

import (
	"context"
	"fmt"
	"time"

	"github.com/faultline/faultline/internal/domain"
	"github.com/faultline/faultline/internal/repository"
	"github.com/faultline/faultline/internal/repository/redis"
	"github.com/google/uuid"
)

type MetricService interface {
	RecordMetric(ctx context.Context, m *domain.Metric) error
	GetByService(ctx context.Context, serviceID uuid.UUID, since time.Time, limit int) ([]domain.Metric, error)
	GetLatestByService(ctx context.Context, serviceID uuid.UUID) (*domain.Metric, error)
}

type metricService struct {
	metricRepo repository.MetricRepository
	cache      *redis.Cache
}

func NewMetricService(metricRepo repository.MetricRepository, cache *redis.Cache) MetricService {
	return &metricService{
		metricRepo: metricRepo,
		cache:      cache,
	}
}

func (s *metricService) RecordMetric(ctx context.Context, m *domain.Metric) error {
	// Persist historical metric to PostgreSQL
	if err := s.metricRepo.Create(ctx, m); err != nil {
		return fmt.Errorf("failed to persist metric: %w", err)
	}

	// Cache short-lived latest metric in Redis
	if s.cache != nil {
		if err := s.cache.SetLatestMetric(ctx, m.ServiceID, m); err != nil {
			// Logged or handled gracefully without failing the write
			return nil
		}
	}

	return nil
}

func (s *metricService) GetByService(ctx context.Context, serviceID uuid.UUID, since time.Time, limit int) ([]domain.Metric, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	metrics, err := s.metricRepo.GetByService(ctx, serviceID, since, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}
	if metrics == nil {
		return make([]domain.Metric, 0), nil
	}
	return metrics, nil
}

func (s *metricService) GetLatestByService(ctx context.Context, serviceID uuid.UUID) (*domain.Metric, error) {
	// Try Redis first
	if s.cache != nil {
		cached, err := s.cache.GetLatestMetric(ctx, serviceID)
		if err == nil && cached != nil {
			return cached, nil
		}
	}

	// Fallback to PostgreSQL
	metric, err := s.metricRepo.GetLatestByService(ctx, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest metric from DB: %w", err)
	}
	return metric, nil
}
