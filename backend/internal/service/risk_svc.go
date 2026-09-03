package service

import (
	"context"
	"fmt"
	"time"

	"github.com/faultline/faultline/internal/domain"
	"github.com/faultline/faultline/internal/engine/risk"
	"github.com/faultline/faultline/internal/repository"
	"github.com/faultline/faultline/internal/repository/redis"
	"github.com/google/uuid"
)

type RiskServiceResponse struct {
	OverallRisk int                     `json:"overallRisk"`
	Level       domain.RiskLevel        `json:"level"`
	Factors     []domain.RiskFactor     `json:"factors"`
	Timestamp   time.Time               `json:"timestamp"`
	Breakdown   domain.AnomalyBreakdown `json:"breakdown"`
	History     []domain.RiskScore      `json:"history,omitempty"`
}

type RiskService interface {
	AssessServiceRisk(ctx context.Context, serviceID uuid.UUID) (*RiskServiceResponse, error)
	GetRiskHistory(ctx context.Context, serviceID uuid.UUID, durationMin int) ([]domain.RiskScore, error)
	GetAllLatest(ctx context.Context) ([]domain.RiskScore, error)
	RecordRiskScore(ctx context.Context, rs *domain.RiskScore) error
}

type riskService struct {
	riskRepo   repository.RiskRepository
	metricRepo repository.MetricRepository
	cache      *redis.Cache
}

func NewRiskService(riskRepo repository.RiskRepository, metricRepo repository.MetricRepository, cache *redis.Cache) RiskService {
	return &riskService{
		riskRepo:   riskRepo,
		metricRepo: metricRepo,
		cache:      cache,
	}
}

func (s *riskService) AssessServiceRisk(ctx context.Context, serviceID uuid.UUID) (*RiskServiceResponse, error) {
	// Retrieve recent metrics for baseline and current observation
	since := time.Now().Add(-30 * time.Minute)
	metrics, err := s.metricRepo.GetByService(ctx, serviceID, since, 360)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics for risk assessment: %w", err)
	}

	if len(metrics) == 0 {
		return &RiskServiceResponse{
			OverallRisk: 0,
			Level:       domain.RiskLow,
			Factors: []domain.RiskFactor{
				{
					Name:   "no_data",
					Score:  0,
					Reason: "Insufficient historical metric data accumulated",
				},
			},
			Timestamp: time.Now(),
		}, nil
	}

	// Current observation is the most recent metric
	current := &metrics[0]

	// Extract historical series
	history := risk.MetricHistory{}
	for _, m := range metrics {
		history.Latencies = append(history.Latencies, m.ResponseLatency)
		history.ErrorRates = append(history.ErrorRates, m.ErrorRate)
		history.TimeoutRates = append(history.TimeoutRates, m.TimeoutRate)
		history.RequestRates = append(history.RequestRates, m.RequestRate)
		history.DepLatencies = append(history.DepLatencies, m.DepLatency)
		history.DepErrorRates = append(history.DepErrorRates, m.DepErrorRate)
	}

	assessment := risk.AssessRisk(current, history)

	// Fetch recent history for trend analysis
	historyScores, _ := s.riskRepo.GetByService(ctx, serviceID, time.Now().Add(-1*time.Hour), 60)
	if historyScores == nil {
		historyScores = make([]domain.RiskScore, 0)
	}

	return &RiskServiceResponse{
		OverallRisk: assessment.OverallRisk,
		Level:       assessment.Level,
		Factors:     assessment.Factors,
		Timestamp:   assessment.Timestamp,
		Breakdown:   assessment.Breakdown,
		History:     historyScores,
	}, nil
}

func (s *riskService) GetRiskHistory(ctx context.Context, serviceID uuid.UUID, durationMin int) ([]domain.RiskScore, error) {
	if durationMin <= 0 {
		durationMin = 60
	}
	since := time.Now().Add(-time.Duration(durationMin) * time.Minute)
	scores, err := s.riskRepo.GetByService(ctx, serviceID, since, 500)
	if err != nil {
		return nil, fmt.Errorf("failed to query risk history: %w", err)
	}
	if scores == nil {
		return make([]domain.RiskScore, 0), nil
	}
	return scores, nil
}

func (s *riskService) GetAllLatest(ctx context.Context) ([]domain.RiskScore, error) {
	scores, err := s.riskRepo.GetAllLatest(ctx)
	if err != nil {
		return nil, err
	}
	if scores == nil {
		return make([]domain.RiskScore, 0), nil
	}
	return scores, nil
}

func (s *riskService) RecordRiskScore(ctx context.Context, rs *domain.RiskScore) error {
	if err := s.riskRepo.Create(ctx, rs); err != nil {
		return err
	}
	if s.cache != nil {
		_ = s.cache.SetRiskScore(ctx, rs.ServiceID, rs)
	}
	return nil
}
