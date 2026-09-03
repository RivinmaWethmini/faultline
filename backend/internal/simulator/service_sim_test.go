package simulator

import (
	"math/rand"
	"testing"
)

func TestDefaultProfiles(t *testing.T) {
	profiles := DefaultProfiles()
	if len(profiles) != 5 {
		t.Fatalf("expected 5 default profiles, got %d", len(profiles))
	}

	expectedServices := map[string]bool{
		"api-gateway":       false,
		"auth-service":      false,
		"order-service":     false,
		"payment-service":   false,
		"inventory-service": false,
	}

	for _, p := range profiles {
		if _, exists := expectedServices[p.Name]; !exists {
			t.Errorf("unexpected service profile: %s", p.Name)
		}
		expectedServices[p.Name] = true

		if p.BaseLatency <= 0 {
			t.Errorf("service %s has invalid base latency: %f", p.Name, p.BaseLatency)
		}
		if p.BaseReqRate <= 0 {
			t.Errorf("service %s has invalid base request rate: %f", p.Name, p.BaseReqRate)
		}
	}

	for name, found := range expectedServices {
		if !found {
			t.Errorf("expected service profile %s was not found", name)
		}
	}
}

func TestGenerateMetric_WithinBounds(t *testing.T) {
	profiles := DefaultProfiles()

	for _, profile := range profiles {
		m := GenerateMetric(profile, nil)

		if m.ResponseLatency <= 0 {
			t.Errorf("latency should be positive, got %f", m.ResponseLatency)
		}
		if m.RequestRate < 0 {
			t.Errorf("request rate cannot be negative, got %f", m.RequestRate)
		}
		if m.ErrorRate < 0 || m.ErrorRate > 100 {
			t.Errorf("error rate must be between 0 and 100, got %f", m.ErrorRate)
		}
		if m.TimeoutRate < 0 || m.TimeoutRate > 100 {
			t.Errorf("timeout rate must be between 0 and 100, got %f", m.TimeoutRate)
		}
		if m.CPUUsage < 0 || m.CPUUsage > 100 {
			t.Errorf("cpu usage must be between 0 and 100, got %f", m.CPUUsage)
		}
		if m.MemoryUsage < 0 || m.MemoryUsage > 100 {
			t.Errorf("memory usage must be between 0 and 100, got %f", m.MemoryUsage)
		}
	}
}

func TestGenerateMetric_WithModifiers(t *testing.T) {
	profile := ServiceProfile{
		Name:        "test-service",
		DisplayName: "Test Service",
		BaseLatency: 50.0,
		BaseReqRate: 100.0,
		BaseErrRate: 1.0,
		BaseTimeout: 0.5,
		BaseCPU:     20.0,
		BaseMemory:  30.0,
		BaseDepLat:  20.0,
		BaseDepErr:  0.2,
	}

	modifiers := map[string]float64{
		"latency_multiplier": 4.0,
		"error_rate_add":     30.0,
		"cpu_add":            40.0,
	}

	rng := rand.New(rand.NewSource(42))
	metric := GenerateDeterministicMetric(profile, modifiers, rng)

	if metric.ResponseLatency < 150 {
		t.Errorf("expected latency to be significantly amplified, got %f", metric.ResponseLatency)
	}
	if metric.ErrorRate < 25 {
		t.Errorf("expected error rate to be elevated with modifier, got %f", metric.ErrorRate)
	}
	if metric.CPUUsage < 50 {
		t.Errorf("expected CPU usage to be elevated, got %f", metric.CPUUsage)
	}
}

func TestDeterministicMetricGeneration(t *testing.T) {
	profile := DefaultProfiles()[0]

	rng1 := rand.New(rand.NewSource(12345))
	metric1 := GenerateDeterministicMetric(profile, nil, rng1)

	rng2 := rand.New(rand.NewSource(12345))
	metric2 := GenerateDeterministicMetric(profile, nil, rng2)

	if metric1.ResponseLatency != metric2.ResponseLatency {
		t.Errorf("expected identical latencies with identical seeds: %f vs %f", metric1.ResponseLatency, metric2.ResponseLatency)
	}
	if metric1.RequestRate != metric2.RequestRate {
		t.Errorf("expected identical request rates with identical seeds: %f vs %f", metric1.RequestRate, metric2.RequestRate)
	}
	if metric1.ErrorRate != metric2.ErrorRate {
		t.Errorf("expected identical error rates with identical seeds: %f vs %f", metric1.ErrorRate, metric2.ErrorRate)
	}
}
