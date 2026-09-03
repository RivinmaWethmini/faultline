package risk

import (
	"math"
)

// Baseline holds the historical mean and standard deviation for a metric.
type Baseline struct {
	Mean   float64
	StdDev float64
	Count  int
}

// UpdateBaseline recalculates the baseline incrementally using Welford's online algorithm.
func UpdateBaseline(b *Baseline, newValue float64) {
	b.Count++
	delta := newValue - b.Mean
	b.Mean += delta / float64(b.Count)
	delta2 := newValue - b.Mean
	if b.Count > 1 {
		// Running variance using Welford's method
		variance := (b.StdDev*b.StdDev*float64(b.Count-1) + delta*delta2) / float64(b.Count)
		b.StdDev = math.Sqrt(variance)
	}
}

// ComputeBaseline calculates the baseline from a slice of values.
func ComputeBaseline(values []float64) Baseline {
	if len(values) == 0 {
		return Baseline{}
	}

	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	var varianceSum float64
	for _, v := range values {
		diff := v - mean
		varianceSum += diff * diff
	}

	stdDev := 0.0
	if len(values) > 1 {
		stdDev = math.Sqrt(varianceSum / float64(len(values)))
	}

	return Baseline{
		Mean:   mean,
		StdDev: stdDev,
		Count:  len(values),
	}
}

// ComputeTrend returns a value indicating the direction of a metric trend.
// Positive values mean the metric is worsening (increasing), negative means improving.
// Range roughly -1.0 to 1.0.
func ComputeTrend(values []float64) float64 {
	n := len(values)
	if n < 3 {
		return 0
	}

	// Use the last 5 points (or fewer if not available)
	start := 0
	if n > 5 {
		start = n - 5
	}
	recent := values[start:]

	// Simple linear regression slope normalized by mean
	rn := len(recent)
	var sumX, sumY, sumXY, sumX2 float64
	for i, v := range recent {
		x := float64(i)
		sumX += x
		sumY += v
		sumXY += x * v
		sumX2 += x * x
	}

	denom := float64(rn)*sumX2 - sumX*sumX
	if denom == 0 {
		return 0
	}

	slope := (float64(rn)*sumXY - sumX*sumY) / denom
	meanY := sumY / float64(rn)

	if meanY == 0 {
		return 0
	}

	// Normalize slope by mean to get a relative trend
	normalized := slope / math.Abs(meanY)

	// Clamp to [-1, 1]
	return math.Max(-1, math.Min(1, normalized))
}
