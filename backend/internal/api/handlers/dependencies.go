package handlers

import (
	"net/http"

	"github.com/faultline/faultline/internal/api/response"
	"github.com/faultline/faultline/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type DependencyHandler struct {
	depSvc service.DependencyService
}

func NewDependencyHandler(depSvc service.DependencyService) *DependencyHandler {
	return &DependencyHandler{depSvc: depSvc}
}

// List returns all service dependencies.
func (h *DependencyHandler) List(w http.ResponseWriter, r *http.Request) {
	deps, err := h.depSvc.GetAll(r.Context())
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, deps)
}

// Impact returns failure propagation and root cause analysis for a specific service.
func (h *DependencyHandler) Impact(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid service ID format")
		return
	}

	impact, err := h.depSvc.GetDependencyImpact(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, impact)
}
