package risk

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/faultline/faultline/internal/domain"
)

// Weights for the overall risk score calculation.
const (
	WeightError      = 0.30
	WeightTimeout    = 0.20
	WeightLatency    = 0.20
	WeightDependency = 0.20
	WeightTraffic    = 0.10
)

// MetricHistory holds historical values for computing baselines and trends.
type MetricHistory struct {
	Latencies     []float64
	ErrorRates    []float64
	TimeoutRates  []float64
	RequestRates  []float64
	DepLatencies  []float64
	DepErrorRates []float64
}

// AssessRisk computes the full explainable risk assessment including overall score,
// classification level, anomaly breakdown, and ranked factors with natural reasons.
func AssessRisk(current *domain.Metric, history MetricHistory) domain.RiskAssessment {
	// Compute baselines from historical data
	latencyBaseline := ComputeBaseline(history.Latencies)
	errorBaseline := ComputeBaseline(history.ErrorRates)
	timeoutBaseline := ComputeBaseline(history.TimeoutRates)
	trafficBaseline := ComputeBaseline(history.RequestRates)
	depLatencyBaseline := ComputeBaseline(history.DepLatencies)
	depErrorBaseline := ComputeBaseline(history.DepErrorRates)

	// Raw anomaly scores
	latencyScore := ComputeAnomalyScore(current.ResponseLatency, latencyBaseline)
	errorScore := ComputeAnomalyForErrorRate(current.ErrorRate, errorBaseline)
	timeoutScore := ComputeAnomalyForErrorRate(current.TimeoutRate, timeoutBaseline)
	trafficScore := ComputeAnomalyForTraffic(current.RequestRate, trafficBaseline)

	// Dependency anomaly: max of latency and error anomalies
	depLatencyScore := ComputeAnomalyScore(current.DepLatency, depLatencyBaseline)
	depErrorScore := ComputeAnomalyForErrorRate(current.DepErrorRate, depErrorBaseline)
	depScore := depLatencyScore
	depIsLatency := true
	if depErrorScore > depScore {
		depScore = depErrorScore
		depIsLatency = false
	}

	// Trend amplification
	latencyTrend := ComputeTrend(history.Latencies)
	errorTrend := ComputeTrend(history.ErrorRates)
	timeoutTrend := ComputeTrend(history.TimeoutRates)
	trafficTrend := ComputeTrend(history.RequestRates)
	depTrend := ComputeTrend(history.DepLatencies)

	latencyScore = ApplyTrendAmplification(latencyScore, latencyTrend)
	errorScore = ApplyTrendAmplification(errorScore, errorTrend)
	timeoutScore = ApplyTrendAmplification(timeoutScore, timeoutTrend)
	trafficScore = ApplyTrendAmplification(trafficScore, trafficTrend)
	depScore = ApplyTrendAmplification(depScore, depTrend)

	breakdown := domain.AnomalyBreakdown{
		LatencyAnomaly:    latencyScore,
		ErrorAnomaly:      errorScore,
		TimeoutAnomaly:    timeoutScore,
		TrafficAnomaly:    trafficScore,
		DependencyAnomaly: depScore,
	}

	overall := ComputeOverallScore(breakdown)
	level := domain.ClassifyRisk(overall)

	// Build factors with deterministic explanations
	var factors []domain.RiskFactor

	if latencyScore > 15 {
		ratio := 1.0
		if latencyBaseline.Mean > 0 {
			ratio = current.ResponseLatency / latencyBaseline.Mean
		}
		factors = append(factors, domain.RiskFactor{
			Name:   "response_latency",
			Score:  latencyScore,
			Reason: fmt.Sprintf("Response latency is %.1fx above baseline (%.1fms vs baseline %.1fms)", ratio, current.ResponseLatency, latencyBaseline.Mean),
		})
	}

	if errorScore > 15 {
		diff := current.ErrorRate - errorBaseline.Mean
		factors = append(factors, domain.RiskFactor{
			Name:   "error_rate",
			Score:  errorScore,
			Reason: fmt.Sprintf("Error rate elevated to %.1f%% (+%.1f%% above baseline %.1f%%)", current.ErrorRate, diff, errorBaseline.Mean),
		})
	}

	if timeoutScore > 15 {
		diff := current.TimeoutRate - timeoutBaseline.Mean
		factors = append(factors, domain.RiskFactor{
			Name:   "timeout_rate",
			Score:  timeoutScore,
			Reason: fmt.Sprintf("Timeout rate spiked to %.1f%% (+%.1f%% above baseline %.1f%%)", current.TimeoutRate, diff, timeoutBaseline.Mean),
		})
	}

	if trafficScore > 15 {
		ratio := 1.0
		if trafficBaseline.Mean > 0 {
			ratio = current.RequestRate / trafficBaseline.Mean
		}
		var reason string
		if ratio >= 1.0 {
			reason = fmt.Sprintf("Traffic surge: %.1fx normal volume (%.0f req/s vs baseline %.0f req/s)", ratio, current.RequestRate, trafficBaseline.Mean)
		} else {
			drop := (1.0 - ratio) * 100
			reason = fmt.Sprintf("Traffic drop: %.0f%% below normal (%.0f req/s vs baseline %.0f req/s)", drop, current.RequestRate, trafficBaseline.Mean)
		}
		factors = append(factors, domain.RiskFactor{
			Name:   "traffic_rate",
			Score:  trafficScore,
			Reason: reason,
		})
	}

	if depScore > 15 {
		var reason string
		var name string
		if depIsLatency {
			name = "dependency_latency"
			ratio := 1.0
			if depLatencyBaseline.Mean > 0 {
				ratio = current.DepLatency / depLatencyBaseline.Mean
			}
			reason = fmt.Sprintf("Dependency latency is %.1fx above baseline (%.1fms vs baseline %.1fms)", ratio, current.DepLatency, depLatencyBaseline.Mean)
		} else {
			name = "dependency_error_rate"
			diff := current.DepErrorRate - depErrorBaseline.Mean
			reason = fmt.Sprintf("Dependency error rate elevated to %.1f%% (+%.1f%% above baseline %.1f%%)", current.DepErrorRate, diff, depErrorBaseline.Mean)
		}
		factors = append(factors, domain.RiskFactor{
			Name:   name,
			Score:  depScore,
			Reason: reason,
		})
	}

	// Sort factors descending by anomaly score
	sort.Slice(factors, func(i, j int) bool {
		return factors[i].Score > factors[j].Score
	})

	if len(factors) == 0 {
		factors = append(factors, domain.RiskFactor{
			Name:   "system_nominal",
			Score:  0,
			Reason: "All operational metrics within normal historical baseline tolerances",
		})
	}

	return domain.RiskAssessment{
		OverallRisk: overall,
		Level:       level,
		Factors:     factors,
		Timestamp:   time.Now(),
		Breakdown:   breakdown,
	}
}

// CalculateRiskScore computes the full risk breakdown for a service given current metrics and history.
func CalculateRiskScore(current *domain.Metric, history MetricHistory) domain.AnomalyBreakdown {
	assessment := AssessRisk(current, history)
	return assessment.Breakdown
}

// ComputeOverallScore calculates the weighted overall risk score from an anomaly breakdown.
func ComputeOverallScore(breakdown domain.AnomalyBreakdown) int {
	score := float64(breakdown.ErrorAnomaly)*WeightError +
		float64(breakdown.TimeoutAnomaly)*WeightTimeout +
		float64(breakdown.LatencyAnomaly)*WeightLatency +
		float64(breakdown.DependencyAnomaly)*WeightDependency +
		float64(breakdown.TrafficAnomaly)*WeightTraffic

	return clampScore(int(math.Round(score)))
}
