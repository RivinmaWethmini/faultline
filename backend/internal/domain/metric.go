package domain

import (
	"time"

	"github.com/google/uuid"
)

type Metric struct {
	ID               uuid.UUID `json:"id"`
	ServiceID        uuid.UUID `json:"service_id"`
	Timestamp        time.Time `json:"timestamp"`
	ResponseLatency  float64   `json:"response_latency_ms"`
	RequestRate      float64   `json:"request_rate"`
	ErrorRate        float64   `json:"error_rate"`
	TimeoutRate      float64   `json:"timeout_rate"`
	CPUUsage         float64   `json:"cpu_usage"`
	MemoryUsage      float64   `json:"memory_usage"`
	DepLatency       float64   `json:"dep_latency_ms"`
	DepErrorRate     float64   `json:"dep_error_rate"`
}
