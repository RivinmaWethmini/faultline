package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/faultline/faultline/internal/api/response"
	"github.com/faultline/faultline/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type MetricHandler struct {
	metricSvc service.MetricService
}

func NewMetricHandler(metricSvc service.MetricService) *MetricHandler {
	return &MetricHandler{metricSvc: metricSvc}
}

// GetByService returns recent metrics for a service.
func (h *MetricHandler) GetByService(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid service ID format")
		return
	}

	durationMin := 30
	if d := r.URL.Query().Get("duration"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			durationMin = parsed
		}
	}

	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	since := time.Now().Add(-time.Duration(durationMin) * time.Minute)
	metrics, err := h.metricSvc.GetByService(r.Context(), id, since, limit)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, metrics)
}
