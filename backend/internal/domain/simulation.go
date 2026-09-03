package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type SimulationStatus string

const (
	SimPending   SimulationStatus = "pending"
	SimRunning   SimulationStatus = "running"
	SimCompleted SimulationStatus = "completed"
)

type Simulation struct {
	ID            uuid.UUID        `json:"id"`
	Scenario      string           `json:"scenario"`
	TargetService string           `json:"target_service"`
	Parameters    json.RawMessage  `json:"parameters,omitempty"`
	Status        SimulationStatus `json:"status"`
	StartedAt     *time.Time       `json:"started_at,omitempty"`
	CompletedAt   *time.Time       `json:"completed_at,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
}
