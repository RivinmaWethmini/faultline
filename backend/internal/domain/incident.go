package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type IncidentStatus string

const (
	IncidentOpen          IncidentStatus = "OPEN"
	IncidentActive        IncidentStatus = "OPEN" // alias
	IncidentInvestigating IncidentStatus = "INVESTIGATING"
	IncidentResolved      IncidentStatus = "RESOLVED"
)

type Incident struct {
	ID               uuid.UUID       `json:"id"`
	ServiceID        uuid.UUID       `json:"service_id"`
	Title            string          `json:"title"`
	Severity         RiskLevel       `json:"severity"`
	Status           IncidentStatus  `json:"status"`
	RiskScore        int             `json:"risk_score"`
	RootCause        string          `json:"root_cause"`
	PropagationPath  json.RawMessage `json:"propagation_path"`
	ImpactedServices json.RawMessage `json:"impacted_services"`
	Anomalies        json.RawMessage `json:"anomalies"`
	StartedAt        time.Time       `json:"started_at"`
	ResolvedAt       *time.Time      `json:"resolved_at,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}

type IncidentEvent struct {
	ID         uuid.UUID       `json:"id"`
	IncidentID uuid.UUID       `json:"incident_id"`
	EventType  string          `json:"event_type"`
	Message    string          `json:"message"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}
