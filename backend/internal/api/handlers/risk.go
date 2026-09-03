package handlers

import (
	"net/http"

	"github.com/faultline/faultline/internal/api/response"
	"github.com/faultline/faultline/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RiskHandler struct {
	riskSvc service.RiskService
}

func NewRiskHandler(riskSvc service.RiskService) *RiskHandler {
	return &RiskHandler{riskSvc: riskSvc}
}

// GetByService returns explainable risk assessment and factors for a service.
func (h *RiskHandler) GetByService(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid service ID format")
		return
	}

	assessment, err := h.riskSvc.AssessServiceRisk(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, assessment)
}
