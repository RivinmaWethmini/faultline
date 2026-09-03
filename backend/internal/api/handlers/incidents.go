package handlers

import (
	"net/http"
	"strconv"

	"github.com/faultline/faultline/internal/api/response"
	"github.com/faultline/faultline/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type IncidentHandler struct {
	incidentSvc service.IncidentService
}

func NewIncidentHandler(incidentSvc service.IncidentService) *IncidentHandler {
	return &IncidentHandler{incidentSvc: incidentSvc}
}

// List returns recent incidents.
func (h *IncidentHandler) List(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	incidents, err := h.incidentSvc.GetAll(r.Context(), status, limit)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, incidents)
}

// Get returns a single incident with its timeline events.
func (h *IncidentHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid incident ID format")
		return
	}

	inc, err := h.incidentSvc.GetByID(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, inc)
}
