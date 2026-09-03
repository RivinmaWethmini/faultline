package risk

import (
	"math"
)

// ComputeAnomalyScore calculates an anomaly score (0-100) from a Z-score.
// Higher Z-score (more standard deviations from mean) = higher anomaly score.
func ComputeAnomalyScore(value float64, baseline Baseline) int {
	if baseline.Count < 3 {
		// Not enough data for meaningful anomaly detection
		return 0
	}

	stdDev := baseline.StdDev
	if stdDev < 0.001 {
		// Near-zero std dev: if value differs from mean at all, flag it moderately
		if math.Abs(value-baseline.Mean) > 0.001 {
			return 50
		}
		return 0
	}

	z := (value - baseline.Mean) / stdDev

	// Use absolute Z-score — anomalous in either direction
	absZ := math.Abs(z)

	// Map Z-score to 0-100 using a sigmoid-like curve
	// z=0 -> 0, z=1 -> ~25, z=2 -> ~50, z=3 -> ~75, z=4+ -> ~90+
	score := 100 * (1 - 1/(1+0.3*absZ*absZ))

	return clampScore(int(math.Round(score)))
}

// ApplyTrendAmplification boosts an anomaly score based on the trend direction.
// If the metric is trending in the anomalous direction, the score is amplified by up to 15%.
func ApplyTrendAmplification(score int, trend float64) int {
	if trend <= 0 {
		return score
	}

	// Trend is positive (worsening): amplify up to 15%
	amplification := 1.0 + 0.15*trend
	amplified := float64(score) * amplification

	return clampScore(int(math.Round(amplified)))
}

// ComputeAnomalyForErrorRate handles error rates specially.
// Error rates are only anomalous when they increase (one-sided).
func ComputeAnomalyForErrorRate(value float64, baseline Baseline) int {
	if baseline.Count < 3 {
		return 0
	}

	// For error rates, only care about values above the mean
	if value <= baseline.Mean {
		return 0
	}

	stdDev := baseline.StdDev
	if stdDev < 0.001 {
		if value > baseline.Mean+0.001 {
			return 60
		}
		return 0
	}

	z := (value - baseline.Mean) / stdDev
	score := 100 * (1 - 1/(1+0.3*z*z))

	return clampScore(int(math.Round(score)))
}

// ComputeAnomalyForTraffic detects both spikes and drops.
func ComputeAnomalyForTraffic(value float64, baseline Baseline) int {
	if baseline.Count < 3 {
		return 0
	}

	stdDev := baseline.StdDev
	if stdDev < 0.001 {
		stdDev = baseline.Mean * 0.1
		if stdDev < 0.001 {
			return 0
		}
	}

	z := math.Abs(value-baseline.Mean) / stdDev

	// Traffic anomalies are less severe than error/latency anomalies
	score := 80 * (1 - 1/(1+0.2*z*z))

	return clampScore(int(math.Round(score)))
}

func clampScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}
