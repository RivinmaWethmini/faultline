package simulator

// Scenario defines a predefined failure scenario.
type Scenario struct {
	Name        string             `json:"name"`
	DisplayName string             `json:"display_name"`
	Description string             `json:"description"`
	Target      string             `json:"target_service"`
	Duration    int                `json:"duration_seconds"`
	Modifiers   map[string]float64 `json:"modifiers"`
}

// DefaultScenarios returns the predefined failure scenarios.
func DefaultScenarios() []Scenario {
	return []Scenario{
		{
			Name:        "payment_latency_spike",
			DisplayName: "Payment Latency Spike",
			Description: "Simulates a severe latency increase in the Payment Service, causing cascading timeouts through Order Service and API Gateway.",
			Target:      "payment-service",
			Duration:    60,
			Modifiers: map[string]float64{
				"latency_multiplier": 5.0,
				"timeout_rate_add":   40.0,
				"cpu_add":            25.0,
			},
		},
		{
			Name:        "database_slowdown",
			DisplayName: "Database Slowdown",
			Description: "Simulates a database performance degradation affecting the Payment Service's dependency latency and error rates.",
			Target:      "payment-service",
			Duration:    90,
			Modifiers: map[string]float64{
				"dep_latency_multiplier": 8.0,
				"dep_error_rate_add":     30.0,
				"latency_multiplier":     2.0,
				"error_rate_add":         15.0,
			},
		},
		{
			Name:        "auth_failure",
			DisplayName: "Authentication Failure",
			Description: "Simulates an authentication service failure causing widespread request errors across the system.",
			Target:      "auth-service",
			Duration:    45,
			Modifiers: map[string]float64{
				"error_rate_add":   60.0,
				"timeout_rate_add": 30.0,
				"cpu_add":          15.0,
			},
		},
		{
			Name:        "network_delay",
			DisplayName: "Network Delay",
			Description: "Simulates a network partition or congestion affecting latency across all services.",
			Target:      "api-gateway",
			Duration:    60,
			Modifiers: map[string]float64{
				"latency_multiplier":     3.0,
				"dep_latency_multiplier": 4.0,
				"timeout_rate_add":       15.0,
			},
		},
		{
			Name:        "traffic_surge",
			DisplayName: "Traffic Surge",
			Description: "Simulates a 10x traffic spike hitting the API Gateway, causing resource exhaustion and cascading latency.",
			Target:      "api-gateway",
			Duration:    60,
			Modifiers: map[string]float64{
				"request_rate_multiplier": 10.0,
				"latency_multiplier":      2.0,
				"cpu_add":                 35.0,
				"error_rate_add":          5.0,
			},
		},
	}
}
