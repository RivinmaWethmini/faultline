package simulator

import (
	"math"
	"math/rand"
	"time"

	"github.com/faultline/faultline/internal/domain"
	"github.com/google/uuid"
)

// ServiceProfile defines the baseline metrics for a simulated service.
type ServiceProfile struct {
	Name         string  `json:"name"`
	DisplayName  string  `json:"display_name"`
	Description  string  `json:"description"`
	BaseLatency  float64 `json:"base_latency_ms"` // ms
	BaseReqRate  float64 `json:"base_request_rate"` // req/s
	BaseErrRate  float64 `json:"base_error_rate"` // percentage (0-100)
	BaseTimeout  float64 `json:"base_timeout_rate"` // percentage (0-100)
	BaseCPU      float64 `json:"base_cpu_usage"` // percentage (0-100)
	BaseMemory   float64 `json:"base_memory_usage"` // percentage (0-100)
	BaseDepLat   float64 `json:"base_dep_latency_ms"` // ms
	BaseDepErr   float64 `json:"base_dep_error_rate"` // percentage (0-100)
}

// DefaultProfiles returns the 5 simulated service profiles.
func DefaultProfiles() []ServiceProfile {
	return []ServiceProfile{
		{
			Name:        "api-gateway",
			DisplayName: "API Gateway",
			Description: "Edge gateway handling all inbound API traffic",
			BaseLatency: 45, BaseReqRate: 1200, BaseErrRate: 0.8,
			BaseTimeout: 0.2, BaseCPU: 35, BaseMemory: 42,
			BaseDepLat: 30, BaseDepErr: 0.5,
		},
		{
			Name:        "auth-service",
			DisplayName: "Auth Service",
			Description: "Handles user authentication and token validation",
			BaseLatency: 25, BaseReqRate: 800, BaseErrRate: 0.5,
			BaseTimeout: 0.1, BaseCPU: 22, BaseMemory: 35,
			BaseDepLat: 10, BaseDepErr: 0.2,
		},
		{
			Name:        "order-service",
			DisplayName: "Order Service",
			Description: "Manages order lifecycle and processing",
			BaseLatency: 80, BaseReqRate: 400, BaseErrRate: 1.2,
			BaseTimeout: 0.5, BaseCPU: 45, BaseMemory: 55,
			BaseDepLat: 50, BaseDepErr: 0.8,
		},
		{
			Name:        "payment-service",
			DisplayName: "Payment Service",
			Description: "Processes payment transactions and settlements",
			BaseLatency: 120, BaseReqRate: 300, BaseErrRate: 1.5,
			BaseTimeout: 0.8, BaseCPU: 38, BaseMemory: 48,
			BaseDepLat: 80, BaseDepErr: 1.0,
		},
		{
			Name:        "inventory-service",
			DisplayName: "Inventory Service",
			Description: "Tracks product availability and stock levels",
			BaseLatency: 35, BaseReqRate: 500, BaseErrRate: 0.6,
			BaseTimeout: 0.3, BaseCPU: 28, BaseMemory: 40,
			BaseDepLat: 15, BaseDepErr: 0.3,
		},
	}
}

// DefaultDependencies returns the service dependency edges (source depends on target).
func DefaultDependencies() []struct{ Source, Target string } {
	return []struct{ Source, Target string }{
		{"api-gateway", "auth-service"},
		{"api-gateway", "order-service"},
		{"order-service", "inventory-service"},
		{"order-service", "payment-service"},
	}
}

// GenerateMetric produces a realistic synthetic metric for a given service profile
// with optional simulation modifiers applied.
func GenerateMetric(profile ServiceProfile, modifiers map[string]float64) domain.Metric {
	return GenerateDeterministicMetric(profile, modifiers, nil)
}

// GenerateDeterministicMetric produces metric with an optional specific *rand.Rand for reproducible test runs.
func GenerateDeterministicMetric(profile ServiceProfile, modifiers map[string]float64, rng *rand.Rand) domain.Metric {
	now := time.Now()

	// Apply natural variance (jitter)
	latency := applyJitterRNG(profile.BaseLatency, 0.15, rng)
	reqRate := applyJitterRNG(profile.BaseReqRate, 0.10, rng)
	errRate := applyJitterRNG(profile.BaseErrRate, 0.20, rng)
	timeout := applyJitterRNG(profile.BaseTimeout, 0.25, rng)
	cpu := applyJitterRNG(profile.BaseCPU, 0.08, rng)
	memory := applyJitterRNG(profile.BaseMemory, 0.05, rng)
	depLat := applyJitterRNG(profile.BaseDepLat, 0.15, rng)
	depErr := applyJitterRNG(profile.BaseDepErr, 0.20, rng)

	// Apply simulation modifiers
	if modifiers != nil {
		if m, ok := modifiers["latency_multiplier"]; ok && m > 0 {
			latency *= m
		}
		if m, ok := modifiers["request_rate_multiplier"]; ok && m > 0 {
			reqRate *= m
		}
		if m, ok := modifiers["error_rate_add"]; ok {
			errRate += m
		}
		if m, ok := modifiers["timeout_rate_add"]; ok {
			timeout += m
		}
		if m, ok := modifiers["cpu_add"]; ok {
			cpu += m
		}
		if m, ok := modifiers["dep_latency_multiplier"]; ok && m > 0 {
			depLat *= m
		}
		if m, ok := modifiers["dep_error_rate_add"]; ok {
			depErr += m
		}
	}

	// Clamp values to realistic physiological bounds
	errRate = math.Max(0, math.Min(100, errRate))
	timeout = math.Max(0, math.Min(100, timeout))
	cpu = math.Max(0, math.Min(100, cpu))
	memory = math.Max(0, math.Min(100, memory))
	depErr = math.Max(0, math.Min(100, depErr))
	latency = math.Max(1.0, latency)
	reqRate = math.Max(0.0, reqRate)
	depLat = math.Max(0.0, depLat)

	return domain.Metric{
		ID:              uuid.New(),
		Timestamp:       now,
		ResponseLatency: math.Round(latency*100) / 100,
		RequestRate:     math.Round(reqRate*100) / 100,
		ErrorRate:       math.Round(errRate*100) / 100,
		TimeoutRate:     math.Round(timeout*100) / 100,
		CPUUsage:        math.Round(cpu*100) / 100,
		MemoryUsage:     math.Round(memory*100) / 100,
		DepLatency:      math.Round(depLat*100) / 100,
		DepErrorRate:    math.Round(depErr*100) / 100,
	}
}

func applyJitterRNG(base, jitterPct float64, rng *rand.Rand) float64 {
	var norm float64
	if rng != nil {
		norm = rng.NormFloat64()
	} else {
		norm = rand.NormFloat64()
	}
	jitter := base * jitterPct * norm
	result := base + jitter
	if result < base*0.1 {
		return base * 0.1 // Floor at 10% of base
	}
	return result
}
