package risk

import (
	"testing"
	"time"

	"github.com/faultline/faultline/internal/domain"
	"github.com/google/uuid"
)

func generateHealthyHistory(n int) MetricHistory {
	h := MetricHistory{}
	for i := 0; i < n; i++ {
		h.Latencies = append(h.Latencies, 45.0+(float64(i%5)-2.0))
		h.ErrorRates = append(h.ErrorRates, 0.5+float64(i%3)*0.1)
		h.TimeoutRates = append(h.TimeoutRates, 0.1+float64(i%2)*0.05)
		h.RequestRates = append(h.RequestRates, 1000.0+float64(i%10)*20.0)
		h.DepLatencies = append(h.DepLatencies, 25.0+float64(i%4)*1.0)
		h.DepErrorRates = append(h.DepErrorRates, 0.2+float64(i%2)*0.1)
	}
	return h
}

func TestNormalBehaviour(t *testing.T) {
	history := generateHealthyHistory(30)

	current := &domain.Metric{
		ID:              uuid.New(),
		Timestamp:       time.Now(),
		ResponseLatency: 46.0,
		ErrorRate:       0.6,
		TimeoutRate:     0.1,
		RequestRate:     1020.0,
		DepLatency:      26.0,
		DepErrorRate:    0.2,
	}

	assessment := AssessRisk(current, history)

	if assessment.OverallRisk >= 30 {
		t.Errorf("expected normal behaviour risk to be < 30 (LOW), got %d (level %s)", assessment.OverallRisk, assessment.Level)
	}
	if assessment.Level != domain.RiskLow {
		t.Errorf("expected level LOW, got %s", assessment.Level)
	}
	if len(assessment.Factors) == 0 {
		t.Errorf("expected at least one factor explanation, got none")
	}
}

func TestGradualDegradation(t *testing.T) {
	// History showing a gradual upward trend over the last several samples
	h := MetricHistory{}
	for i := 0; i < 20; i++ {
		h.Latencies = append(h.Latencies, 45.0+float64(i)*6.0) // 45ms up to 165ms
		h.ErrorRates = append(h.ErrorRates, 0.5+float64(i)*0.4) // 0.5% up to 8.5%
		h.TimeoutRates = append(h.TimeoutRates, 0.1+float64(i)*0.2)
		h.RequestRates = append(h.RequestRates, 1000.0)
		h.DepLatencies = append(h.DepLatencies, 25.0+float64(i)*4.0)
		h.DepErrorRates = append(h.DepErrorRates, 0.2+float64(i)*0.3)
	}

	current := &domain.Metric{
		ID:              uuid.New(),
		Timestamp:       time.Now(),
		ResponseLatency: 180.0,
		ErrorRate:       9.5,
		TimeoutRate:     4.5,
		RequestRate:     1000.0,
		DepLatency:      110.0,
		DepErrorRate:    6.5,
	}

	assessment := AssessRisk(current, h)

	if assessment.OverallRisk < 30 {
		t.Errorf("expected gradual degradation to elevate risk score >= 30, got %d", assessment.OverallRisk)
	}
	if assessment.Level == domain.RiskLow {
		t.Errorf("expected risk level MODERATE, HIGH or CRITICAL, got %s", assessment.Level)
	}

	// Verify explainable factors returned
	foundLatency := false
	for _, f := range assessment.Factors {
		if f.Name == "response_latency" || f.Name == "error_rate" || f.Name == "dependency_latency" {
			foundLatency = true
			if f.Reason == "" {
				t.Errorf("factor %s is missing human-readable reason", f.Name)
			}
		}
	}
	if !foundLatency {
		t.Errorf("expected degradation factors to be identified, got %+v", assessment.Factors)
	}
}

func TestSuddenDegradation(t *testing.T) {
	history := generateHealthyHistory(30)

	// Sudden catastrophic failure: 8x latency, 45% error rate, 30% timeout rate
	current := &domain.Metric{
		ID:              uuid.New(),
		Timestamp:       time.Now(),
		ResponseLatency: 450.0,
		ErrorRate:       45.0,
		TimeoutRate:     30.0,
		RequestRate:     1000.0,
		DepLatency:      250.0,
		DepErrorRate:    35.0,
	}

	assessment := AssessRisk(current, history)

	if assessment.OverallRisk < 70 {
		t.Errorf("expected sudden severe failure to have risk >= 70, got %d", assessment.OverallRisk)
	}
	if assessment.Level != domain.RiskHigh && assessment.Level != domain.RiskCritical {
		t.Errorf("expected HIGH or CRITICAL level, got %s", assessment.Level)
	}

	// Verify top factor explains the anomaly
	if len(assessment.Factors) == 0 {
		t.Fatalf("expected explanatory factors, got none")
	}
	topFactor := assessment.Factors[0]
	if topFactor.Score < 60 {
		t.Errorf("expected top factor to have high anomaly score, got %d", topFactor.Score)
	}
	if topFactor.Reason == "" {
		t.Errorf("expected top factor reason to be populated")
	}
}
