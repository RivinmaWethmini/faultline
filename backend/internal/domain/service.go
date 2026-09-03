package domain

import (
	"time"

	"github.com/google/uuid"
)

type ServiceStatus string

const (
	StatusHealthy  ServiceStatus = "healthy"
	StatusDegraded ServiceStatus = "degraded"
	StatusUnhealthy ServiceStatus = "unhealthy"
	StatusCritical ServiceStatus = "critical"
)

type Service struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	DisplayName string        `json:"display_name"`
	Description string        `json:"description"`
	Status      ServiceStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}
