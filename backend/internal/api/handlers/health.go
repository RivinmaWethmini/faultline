package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/faultline/faultline/internal/api/response"
	"github.com/faultline/faultline/internal/repository/redis"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	db    *pgxpool.Pool
	cache *redis.Cache
}

func NewHealthHandler(db *pgxpool.Pool, cache *redis.Cache) *HealthHandler {
	return &HealthHandler{db: db, cache: cache}
}

type ComponentHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type HealthResponse struct {
	Status     string                     `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Version    string                     `json:"version"`
	Components map[string]ComponentHealth `json:"components"`
}

// Check returns the system health status including database and cache.
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	overallStatus := "healthy"
	components := make(map[string]ComponentHealth)

	// Database check
	if h.db != nil {
		if err := h.db.Ping(ctx); err != nil {
			overallStatus = "degraded"
			components["postgres"] = ComponentHealth{
				Status:  "unhealthy",
				Message: err.Error(),
			}
		} else {
			components["postgres"] = ComponentHealth{Status: "healthy"}
		}
	}

	// Redis check
	if h.cache != nil {
		if _, err := h.cache.GetServiceStatus(ctx, "health-ping"); err != nil {
			overallStatus = "degraded"
			components["redis"] = ComponentHealth{
				Status:  "unhealthy",
				Message: err.Error(),
			}
		} else {
			components["redis"] = ComponentHealth{Status: "healthy"}
		}
	}

	res := HealthResponse{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Version:    "1.0.0",
		Components: components,
	}

	if overallStatus == "healthy" {
		response.Success(w, res)
	} else {
		response.JSON(w, http.StatusServiceUnavailable, response.APIResponse{
			Data:    res,
			Success: false,
			Error:   "system degraded",
		})
	}
}
