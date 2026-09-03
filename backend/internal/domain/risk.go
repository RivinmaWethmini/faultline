package domain

import (
	"time"

	"github.com/google/uuid"
)

type RiskLevel string

const (
	RiskLow      RiskLevel = "LOW"
	RiskModerate RiskLevel = "MODERATE"
	RiskHigh     RiskLevel = "HIGH"
	RiskCritical RiskLevel = "CRITICAL"
)

func ClassifyRisk(score int) RiskLevel {
	switch {
	case score >= 80:
		return RiskCritical
	case score >= 60:
		return RiskHigh
	case score >= 30:
		return RiskModerate
	default:
		return RiskLow
	}
}

type RiskFactor struct {
	Name   string `json:"name"`
	Score  int    `json:"score"`
	Reason string `json:"reason"`
}

type RiskAssessment struct {
	OverallRisk int              `json:"overallRisk"`
	Level       RiskLevel        `json:"level"`
	Factors     []RiskFactor     `json:"factors"`
	Timestamp   time.Time        `json:"timestamp"`
	Breakdown   AnomalyBreakdown `json:"breakdown"`
}

type RiskScore struct {
	ID                uuid.UUID    `json:"id"`
	ServiceID         uuid.UUID    `json:"service_id"`
	Timestamp         time.Time    `json:"timestamp"`
	OverallScore      int          `json:"overall_score"`
	Level             RiskLevel    `json:"risk_level"`
	LatencyAnomaly    int          `json:"latency_anomaly"`
	ErrorAnomaly      int          `json:"error_anomaly"`
	TimeoutAnomaly    int          `json:"timeout_anomaly"`
	TrafficAnomaly    int          `json:"traffic_anomaly"`
	DependencyAnomaly int          `json:"dependency_anomaly"`
	Factors           []RiskFactor `json:"factors,omitempty"`
}

type AnomalyBreakdown struct {
	LatencyAnomaly    int `json:"latency_anomaly"`
	ErrorAnomaly      int `json:"error_anomaly"`
	TimeoutAnomaly    int `json:"timeout_anomaly"`
	TrafficAnomaly    int `json:"traffic_anomaly"`
	DependencyAnomaly int `json:"dependency_anomaly"`
}
