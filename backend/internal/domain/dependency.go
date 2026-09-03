package domain

import (
	"time"

	"github.com/google/uuid"
)

type Dependency struct {
	ID             uuid.UUID `json:"id"`
	SourceID       uuid.UUID `json:"source_id"`
	TargetID       uuid.UUID `json:"target_id"`
	DependencyType string    `json:"dependency_type"`
	CreatedAt      time.Time `json:"created_at"`
}

// DependencyEdge is a simplified view used for graph operations.
type DependencyEdge struct {
	SourceName string `json:"source"`
	TargetName string `json:"target"`
}

type RootCauseCandidate struct {
	ServiceName string  `json:"service_name"`
	RiskScore   int     `json:"risk_score"`
	Distance    int     `json:"distance"`
	Confidence  float64 `json:"confidence"`
	Reason      string  `json:"reason"`
}

// DependencyImpactResult represents the comprehensive failure propagation analysis.
type DependencyImpactResult struct {
	Service            Service              `json:"service"`
	UpstreamServices   []string             `json:"upstream_services"`
	DownstreamServices []string             `json:"downstream_services"`
	PossibleRootCauses []RootCauseCandidate `json:"possible_root_causes"`
	AffectedServices   []string             `json:"affected_services"`
	PropagationPaths   [][]string           `json:"propagation_paths"`
}

// PropagationResult represents the failure propagation analysis for a service.
type PropagationResult struct {
	RootCause        string   `json:"root_cause"`
	PropagationPath  []string `json:"propagation_path"`
	AffectedServices []string `json:"affected_services"`
	UpstreamImpact   []string `json:"upstream_impact"`
}
