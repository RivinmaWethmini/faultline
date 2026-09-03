package incident

import (
	"fmt"
	"strings"

	"github.com/faultline/faultline/internal/domain"
)

// GenerateTitle creates a deterministic incident title from the anomaly data.
func GenerateTitle(serviceName string, level domain.RiskLevel, breakdown domain.AnomalyBreakdown) string {
	var primary string
	maxScore := 0

	checks := []struct {
		name  string
		score int
	}{
		{"latency", breakdown.LatencyAnomaly},
		{"error rate", breakdown.ErrorAnomaly},
		{"timeout rate", breakdown.TimeoutAnomaly},
		{"dependency health", breakdown.DependencyAnomaly},
		{"traffic", breakdown.TrafficAnomaly},
	}

	for _, c := range checks {
		if c.score > maxScore {
			maxScore = c.score
			primary = c.name
		}
	}

	return fmt.Sprintf("%s — %s anomaly detected (%s risk)",
		serviceName, primary, strings.ToLower(string(level)))
}

// GenerateExplanation creates a deterministic, template-based explanation
// of an incident from system data. No LLM involved.
func GenerateExplanation(
	serviceName string,
	breakdown domain.AnomalyBreakdown,
	propagation domain.PropagationResult,
) string {
	var parts []string

	// Describe anomalous metrics
	if breakdown.LatencyAnomaly > 30 {
		parts = append(parts, fmt.Sprintf(
			"Latency anomaly score: %d/100.", breakdown.LatencyAnomaly))
	}
	if breakdown.ErrorAnomaly > 30 {
		parts = append(parts, fmt.Sprintf(
			"Error rate anomaly: %d/100.", breakdown.ErrorAnomaly))
	}
	if breakdown.TimeoutAnomaly > 30 {
		parts = append(parts, fmt.Sprintf(
			"Timeout rate anomaly: %d/100.", breakdown.TimeoutAnomaly))
	}
	if breakdown.TrafficAnomaly > 30 {
		parts = append(parts, fmt.Sprintf(
			"Traffic anomaly: %d/100.", breakdown.TrafficAnomaly))
	}
	if breakdown.DependencyAnomaly > 30 {
		parts = append(parts, fmt.Sprintf(
			"Dependency health degraded: %d/100.", breakdown.DependencyAnomaly))
	}

	// Describe root cause
	if propagation.RootCause != "" && propagation.RootCause != serviceName {
		parts = append(parts, fmt.Sprintf(
			"Suspected root cause: %s.", propagation.RootCause))
	}

	// Describe propagation path
	if len(propagation.PropagationPath) > 1 {
		parts = append(parts, fmt.Sprintf(
			"Failure propagation: %s.", strings.Join(propagation.PropagationPath, " → ")))
	}

	// Describe affected services
	if len(propagation.AffectedServices) > 0 {
		parts = append(parts, fmt.Sprintf(
			"Potentially affected services: %s.", strings.Join(propagation.AffectedServices, ", ")))
	}

	if len(parts) == 0 {
		return fmt.Sprintf("Anomalous behaviour detected on %s.", serviceName)
	}

	return strings.Join(parts, " ")
}
